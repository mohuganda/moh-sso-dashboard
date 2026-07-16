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
- a persistent `upload_data` volume mounted into the backend at `/data/uploads` when `STORAGE_PROVIDER=local`

The backend document and announcement upload flows use the configured storage provider. For local filesystem storage, production Compose must include both:

```yaml
volumes:
  upload_data:

services:
  backend:
    volumes:
      - ${APP_ENV_FILE:-./app.env}:/app/app.env:ro
      - upload_data:/data/uploads
```

and `app.env` must include:

```env
STORAGE_PROVIDER=local
LOCAL_BASE_PATH=/data/uploads
APP_BASE_URL=https://dashboards.health.go.ug/ssobackend
```

If the server is still using an older Compose file that only mounts `./app.env`, uploads can fail in production with `UPLOAD_FAILED` even though they work locally.

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

After deployment, verify the backend can write to the upload volume:

```bash
docker compose -f docker-compose-nginx.yml exec backend sh -lc \
  'mkdir -p /data/uploads/.healthcheck && echo ok > /data/uploads/.healthcheck/write-test && cat /data/uploads/.healthcheck/write-test'
```

Expected output:

```text
ok
```

If this fails with `permission denied`, recreate the volume with the current backend image or switch production uploads to S3/MinIO. If multiple backend containers are used, prefer S3/MinIO or a shared filesystem outside Compose; a named Docker volume is local to one Docker host.

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

For upload failures, match the frontend `requestId` to backend logs:

```bash
docker logs sso-backend 2>&1 | grep '50c50103-32ab-4887-9e18-8877af4484b5'
```

The backend logs storage upload failures with the storage provider, storage location ID, object key, filename, and file size.

## Database Migrations

Review migration compatibility before deployment. Back up the database before irreversible migrations. Application rollback cannot reverse an incompatible database migration automatically.
