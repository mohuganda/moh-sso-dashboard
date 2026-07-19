# Helm Chart Packaging

The MOH SSO chart lives in `charts/moh-sso`. It can be linted, rendered, packaged as a `.tgz`, and later installed on any Kubernetes server with Helm.

## Version Model

Use separate versions for the chart and the application images:

| Value | Meaning |
| --- | --- |
| `Chart.yaml version` | Helm chart package version, for example `0.1.0`. |
| `Chart.yaml appVersion` | Informational app version shown by Helm. |
| `backend.image.tag` | Immutable backend Docker image tag to deploy. |
| `frontend.image.tag` | Immutable frontend Docker image tag to deploy. |

Production deployments must use explicit immutable image tags. Do not deploy `latest`.

## Local Validation

From the repository root:

```bash
make helm-lint
make helm-template
make helm-template-local
make helm-template-dev
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 make helm-template-prod
```

`helm-template-prod` requires image tags because production values must prove they can render with deployable images.

## Package The Chart

Package with the versions from `Chart.yaml`:

```bash
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 make helm-package
```

Package with temporary chart metadata overrides:

```bash
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 \
CHART_VERSION=0.1.1 APP_VERSION=1.2.3 \
make helm-package
```

The package is written to:

```text
dist/helm/moh-sso-<chart-version>.tgz
dist/helm/checksums.txt
```

`dist/` is ignored by Git. CI uploads the chart package as an artifact.

## Install From A Package

Copy the `.tgz` and the production values file to the target server, then run:

```bash
kubectl create namespace moh-sso --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install moh-sso-dashboard ./dist/helm/moh-sso-0.1.0.tgz \
  -n moh-sso \
  -f charts/moh-sso/values-prod.yaml \
  --set backend.image.tag=1.2.3 \
  --set frontend.image.tag=1.2.3
```

Check the deployment:

```bash
helm status moh-sso-dashboard -n moh-sso
kubectl get pods -n moh-sso
kubectl rollout status deployment/moh-sso-dashboard-backend -n moh-sso
kubectl rollout status deployment/moh-sso-dashboard-frontend -n moh-sso
```

## Rollback

List releases and roll back to a known-good revision:

```bash
helm history moh-sso-dashboard -n moh-sso
helm rollback moh-sso-dashboard <revision> -n moh-sso
```

## CI Packaging

`.github/workflows/helm-package.yml` validates and packages the chart on chart-related pull requests and pushes to `main`.

The workflow runs:

1. `helm lint`
2. base, local, dev, and prod template renders
3. `helm package`
4. checksum generation
5. chart artifact upload

## Future OCI Registry Publishing

When you are ready to publish chart packages to an OCI registry:

```bash
helm registry login ghcr.io
helm push dist/helm/moh-sso-0.1.0.tgz oci://ghcr.io/mohuganda/charts
helm pull oci://ghcr.io/mohuganda/charts/moh-sso --version 0.1.0
```

Keep registry publishing separate from validation until the registry permissions and release process are confirmed.

## Production Notes

- Store secrets outside the chart package and provide them through Kubernetes secrets or sealed/external secret tooling.
- Use `values-prod.yaml` as the production baseline and override environment-specific settings through a private values file.
- Keep backend and frontend image tags immutable.
- Do not commit packaged `.tgz` files unless a release process explicitly requires it.
