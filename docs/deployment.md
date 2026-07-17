# Backend Deployment

## Immutable Inputs

Production Compose requires immutable `BACKEND_TAG` and `FRONTEND_TAG` values. Accepted examples:

```env
BACKEND_TAG=1.2.3
BACKEND_TAG=sha-d3062e2
FRONTEND_TAG=1.2.3
FRONTEND_TAG=sha-d3062e2
```

An empty tag or `latest` is not accepted for production. Backend and frontend tags remain independently managed, but both must be explicit.

## Compose Production Files

Use `docker-compose-nginx.yml` when the server should expose a reverse proxy on port `80`.
It mounts the checked-in Nginx config from `./nginx`, proxies `/api/*` to the backend, and sends all other traffic to the frontend container.

Use `docker-compose.yml` when another reverse proxy already exists on the host and you only want the backend/frontend bound to localhost.

Both compose files expect:

- `./app.env` for runtime environment
- `./secrets/db_password.txt`
- `./secrets/keycloak_admin_client_secret.txt`
- `./secrets/keycloak_web_client_secret.txt`

For syntax validation without real production values:

```bash
APP_ENV_FILE=./app.env.example BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 \
  docker compose -f docker-compose-nginx.yml config --quiet
```

For deployment:

```bash
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 \
  docker compose -f docker-compose-nginx.yml pull

BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 \
  docker compose -f docker-compose-nginx.yml up -d
```

## Helm Packaged Chart Deployment

Kubernetes deployments can use the packaged Helm chart from `charts/moh-sso`.
Compose remains supported; Helm is the portable Kubernetes packaging path.

Validate chart changes locally:

```bash
make helm-lint
make helm-template
make helm-template-local
make helm-template-dev
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 make helm-template-prod
```

Package the chart:

```bash
BACKEND_TAG=1.2.3 FRONTEND_TAG=1.2.3 make helm-package
```

This creates:

```text
dist/helm/moh-sso-<chart-version>.tgz
dist/helm/checksums.txt
```

Install or upgrade from the package:

```bash
kubectl create namespace moh-sso --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install moh-sso-dashboard ./dist/helm/moh-sso-0.1.0.tgz \
  -n moh-sso \
  -f charts/moh-sso/values-prod.yaml \
  --set backend.image.tag=1.2.3 \
  --set frontend.image.tag=1.2.3
```

See [`docs/helm-packaging.md`](helm-packaging.md) for chart versioning, packaging, CI artifacts, OCI registry notes, and rollback commands.

## Automated Deployment

The deployment workflow supports:

- merged backend release PRs and backend release workflow runs using the published backend release tag
- manual deployment of an explicitly supplied immutable backend image tag or `backend/v<version>` tag, which is normalized to the image tag

The remote deployment:

1. Records the previously deployed backend/frontend tags.
2. Restarts using the requested immutable tags.
3. Calls `/version` and verifies expected version and commit.
4. Calls `/health/live`.
5. Records the image digest, actor, timestamp, version, and commit.
6. Restores the previous tags when verification fails.

The deploy workflow now reads the release metadata artifact emitted by the backend release workflow, so production follows the exact published backend release instead of a generic build output. The artifact carries both the Git release tag and the Docker image tag, and deployment uses the immutable image tag while recording the matching release tag in the deployment history.

The server-side `restart` script must honor exported `BACKEND_TAG` and `FRONTEND_TAG` values and use the production Compose configuration.

## Manual Verification

```bash
curl -fsS http://localhost:9000/version
curl -fsS http://localhost:9000/health/live
docker image inspect ghcr.io/mohuganda/moh-sso-dashboard-backend:1.2.3
```

The binary metadata, OCI labels, requested tag, and expected Git commit must agree.

## Database Migrations

Review migration compatibility before deployment. Back up the database before irreversible migrations. Application rollback cannot reverse an incompatible database migration automatically.
