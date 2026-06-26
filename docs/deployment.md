# Backend Deployment

## Immutable Inputs

Production Compose requires `BACKEND_TAG`. Accepted examples:

```env
BACKEND_TAG=1.2.3
BACKEND_TAG=sha-d3062e2
```

An empty tag or `latest` is not accepted for production. Frontend tags remain independently managed.

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
