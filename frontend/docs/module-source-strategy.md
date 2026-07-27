# Module Source Strategy

The frontend supports two module-source models.

## Source Mode

Source mode is the default production path.

Docker builds everything from this monorepo:

```bash
npm ci
npm run build:docker
```

The build creates one static `dist/` folder containing:

- shell assets
- app bundles under `/mf/<app>/`
- shared package bundles under `/packages/<package>/`
- `import-map.json`
- `version-manifest.json`

This is the safest production mode today because the Docker image and source commit are the release unit.

## NPM Package Mode

NPM package mode is the supported future path for independently versioned modules.

In this mode:

- apps and shared packages are published as `@moh-sso/*`
- a Docker/build pipeline installs selected package versions from npm
- `stage:mf:npm` copies installed package `dist/` folders from `node_modules`
- the shell still produces one static deployment artifact

This lets one app or shared package be rolled forward or back by changing package versions, without requiring every app to become independently hosted on day one.

Production package mode requires exact versions for all 22 published
workspaces. Floating ranges and npm dist-tags are rejected.

```bash
export MOH_SSO_NPM_MODULES="$(npm run --silent npm:module-specs)"

DOCKER_BUILDKIT=1 docker build \
  --secret id=npmrc,src="$HOME/.npmrc" \
  --build-arg MOH_SSO_NPM_MODULES="$MOH_SSO_NPM_MODULES" \
  -f Dockerfile.npm-modules \
  -t moh-sso-dashboard-frontend:npm .
```

The npm credential is mounted only during dependency installation. It is not
copied into an image layer or the final nginx image.

Before building a production image, verify the selected versions:

```bash
npm view @moh-sso/users@0.1.1 version
npm view @moh-sso/ui@0.1.1 version
```

The generated `version-manifest.json` records `deploymentMode: "npm"` and the
installed app/package versions.

## Version Control

Changesets owns package versions:

```bash
npm run changeset
npm run version:packages
npm run release:packages
```

Docker image tags should identify the shell/runtime deployment. Package versions identify app/package artifacts.

Example:

```text
Docker image: ghcr.io/mohuganda/moh-sso-dashboard-frontend:2026.06.07
Users app:   @moh-sso/users@0.4.1
API package: @moh-sso/api@0.2.0
```

## Rollback

Source mode rollback:

- redeploy the previous Docker image tag

NPM package mode rollback:

- install a previous `@moh-sso/*` package version
- rebuild/stage the static artifact
- redeploy the Docker image

Remote import-map deployment rollback:

- point `import-map.json` back to previous app/package URLs
- redeploy or invalidate the import map

## Import Maps

Source mode import maps point to staged local paths:

```text
/mf/users/single-spa.js
/packages/api/index.js
```

Versioned remote import maps can point to versioned asset paths:

```text
/mf/users/0.4.1/single-spa.js
/packages/api/0.2.0/index.js
```

Source mode remains the fallback and local-development default. npm mode may be
enabled after the first package release and a successful npm-based image smoke
test.
