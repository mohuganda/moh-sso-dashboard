#!/bin/sh

set -eu

CHART_DIR=${CHART_DIR:-charts/moh-sso}
RELEASE_NAME=${HELM_RELEASE:-moh-sso-dashboard}
OUT_DIR=${HELM_OUT_DIR:-dist/helm}
BACKEND_TAG=${BACKEND_TAG:-}
FRONTEND_TAG=${FRONTEND_TAG:-}
CHART_VERSION=${CHART_VERSION:-}
APP_VERSION=${APP_VERSION:-}

if ! command -v helm >/dev/null 2>&1; then
  echo "helm is required to package the chart" >&2
  exit 1
fi

if [ ! -f "$CHART_DIR/Chart.yaml" ]; then
  echo "chart not found: $CHART_DIR/Chart.yaml" >&2
  exit 1
fi

chart_name=$(awk -F': *' '$1 == "name" { print $2; exit }' "$CHART_DIR/Chart.yaml" | tr -d '"')
chart_version=$(awk -F': *' '$1 == "version" { print $2; exit }' "$CHART_DIR/Chart.yaml" | tr -d '"')
app_version=$(awk -F': *' '$1 == "appVersion" { print $2; exit }' "$CHART_DIR/Chart.yaml" | tr -d '"')

if [ -z "$chart_name" ] || [ -z "$chart_version" ]; then
  echo "Chart.yaml must define name and version" >&2
  exit 1
fi

if ! printf '%s' "$chart_version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'; then
  echo "chart version must be SemVer: $chart_version" >&2
  exit 1
fi

if [ -n "$CHART_VERSION" ] && ! printf '%s' "$CHART_VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'; then
  echo "CHART_VERSION must be SemVer: $CHART_VERSION" >&2
  exit 1
fi

if [ -z "$BACKEND_TAG" ] || [ -z "$FRONTEND_TAG" ]; then
  echo "BACKEND_TAG and FRONTEND_TAG are required for production template validation" >&2
  echo "example: BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 $0" >&2
  exit 1
fi

case "$BACKEND_TAG" in
  latest|"")
    echo "BACKEND_TAG must be immutable and cannot be latest" >&2
    exit 1
    ;;
esac

case "$FRONTEND_TAG" in
  latest|"")
    echo "FRONTEND_TAG must be immutable and cannot be latest" >&2
    exit 1
    ;;
esac

package_chart_dir=$CHART_DIR
tmp_dir=

cleanup() {
  if [ -n "$tmp_dir" ] && [ -d "$tmp_dir" ]; then
    rm -rf "$tmp_dir"
  fi
}
trap cleanup EXIT INT TERM

if [ -n "$CHART_VERSION" ] || [ -n "$APP_VERSION" ]; then
  tmp_dir=$(mktemp -d)
  package_chart_dir="$tmp_dir/$chart_name"
  cp -R "$CHART_DIR" "$package_chart_dir"

  if [ -n "$CHART_VERSION" ]; then
    sed -i.bak "s/^version: .*/version: $CHART_VERSION/" "$package_chart_dir/Chart.yaml"
    rm -f "$package_chart_dir/Chart.yaml.bak"
    chart_version=$CHART_VERSION
  fi

  if [ -n "$APP_VERSION" ]; then
    sed -i.bak "s/^appVersion: .*/appVersion: \"$APP_VERSION\"/" "$package_chart_dir/Chart.yaml"
    rm -f "$package_chart_dir/Chart.yaml.bak"
    app_version=$APP_VERSION
  fi
fi

echo "Packaging Helm chart:"
echo "  chart: $chart_name"
echo "  version: $chart_version"
echo "  appVersion: ${app_version:-unknown}"
echo "  backend image tag: $BACKEND_TAG"
echo "  frontend image tag: $FRONTEND_TAG"

if [ -f "$package_chart_dir/Chart.lock" ] || grep -Eq '^dependencies:' "$package_chart_dir/Chart.yaml"; then
  helm dependency update "$package_chart_dir"
fi

helm lint "$package_chart_dir"

helm template "$RELEASE_NAME" "$package_chart_dir" >/dev/null
helm template "$RELEASE_NAME" "$package_chart_dir" -f "$package_chart_dir/values-local.yaml" >/dev/null
helm template "$RELEASE_NAME" "$package_chart_dir" -f "$package_chart_dir/values-dev.yaml" >/dev/null
helm template "$RELEASE_NAME" "$package_chart_dir" -f "$package_chart_dir/values-prod.yaml" \
  --set "backend.image.tag=$BACKEND_TAG" \
  --set "frontend.image.tag=$FRONTEND_TAG" >/dev/null

mkdir -p "$OUT_DIR"
rm -f "$OUT_DIR/$chart_name-"*.tgz "$OUT_DIR/checksums.txt"
helm package "$package_chart_dir" --destination "$OUT_DIR"

(
  cd "$OUT_DIR"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum ./*.tgz > checksums.txt
  else
    shasum -a 256 ./*.tgz > checksums.txt
  fi
)

echo "Helm chart package created:"
ls -1 "$OUT_DIR/$chart_name-"*.tgz
echo "Checksums:"
cat "$OUT_DIR/checksums.txt"
