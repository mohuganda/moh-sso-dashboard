# Release CI Outline

This is a CI/CD outline for future automation. It is documentation only; no workflow file is required yet.

## Validate Job

```bash
npm ci
npm run audit:packages
npm run typecheck
npm run lint
npm run build:all
npm run audit:publishability
```

## Version Job

```bash
npm run version:packages
```

This applies Changesets version bumps and updates internal dependency ranges.

## Publish Job

```bash
npm run release:packages
```

This publishes npm packages according to Changesets.

## Source Docker Deploy Job

Build the default production image from source:

```bash
docker build -f frontend/Dockerfile frontend
```

This path remains the default production deployment.

## Optional NPM Module Docker Job

After packages are published, build the optional npm-module image with exact package versions:

```bash
docker build \
  -f frontend/Dockerfile.npm-modules \
  --build-arg MOH_SSO_NPM_MODULES="@moh-sso/users@0.4.1 @moh-sso/api@0.2.0" \
  frontend
```

Use this only when package publishing and version pinning are stable.

## Deploy Job

Generate version metadata:

```bash
npm run generate:versions
```

Generate a versioned import map:

```bash
FRONTEND_ASSET_BASE_URL=https://cdn.example.com FRONTEND_VERSIONED_IMPORTS=true npm run generate:import-map
```

Then upload app and package bundles to versioned paths:

```text
/mf/users/0.4.1/single-spa.js
/mf/documents/0.8.0/single-spa.js
/packages/api/0.2.0/index.js
```

Deploy:

- shell static files
- import map
- `version-manifest.json`
- app/package bundles

## Rollback Job

Rollback by updating the import map to previous bundle versions.

Example:

```json
{
  "imports": {
    "@moh-sso/users": "https://cdn.example.com/mf/users/0.4.0/single-spa.js"
  }
}
```

Redeploy the import map or shell config and clear cache if required.
