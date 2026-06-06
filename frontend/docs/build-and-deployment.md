# Build And Deployment

## Build Commands

Build shared packages:

```bash
npm run build:packages
```

Build standalone apps:

```bash
npm run build:apps
```

Generate import map:

```bash
npm run generate:import-map
```

Build shell:

```bash
npm run build:shell
```

Build everything:

```bash
npm run build:all
```

Build the static Docker-ready artifact:

```bash
npm run build:docker
```

Generate version metadata:

```bash
npm run generate:versions
```

## What `build:docker` Does

`build:docker` runs:

1. build packages
2. generate import map
3. build apps
4. build shell
5. generate version manifest
6. stage microfrontends into `dist`

## Final `dist` Layout

```text
dist/index.html
dist/assets/*
dist/import-map.json
dist/version-manifest.json
dist/mf/<app>/single-spa.js
dist/packages/<package>/index.js
```

This layout supports single static deployment and prepares remote import-map deployment.

## Docker Production

From the repository root:

```bash
docker build -f frontend/Dockerfile-nginx frontend
```

The production image builds the frontend and serves the final static files with nginx.

## Docker Dev

From the repository root:

```bash
docker build -f frontend/Dockerfile.dev frontend
```

Run with compose:

```bash
docker compose -f docker-compose.dev.yml up frontend
```

Run all remote dev servers in Docker:

```bash
FRONTEND_DEV_COMMAND="npm run dev:remote:all" docker compose -f docker-compose.dev.yml up frontend
```

## Deployment Modes

### Single Static Deployment

Deploy one artifact containing:

- shell
- staged app bundles
- staged package bundles
- import map

This is the simplest deployment mode.

### Remote Microfrontend Deployment

Deploy shell and apps independently. The import map controls app and package versions.

In this mode:

- shell reads the import map
- apps are hosted separately
- changing an app version can be done by updating the import map

Versioned deployment paths should look like:

```text
/mf/users/0.4.1/single-spa.js
/packages/api/0.2.0/index.js
```

## Static Hosting Notes

The static server must:

- serve `dist`
- provide SPA fallback to `index.html`
- serve `/mf/*`
- serve `/packages/*`
- serve `/config.js`
- serve `/import-map.json` when remote mode is enabled
