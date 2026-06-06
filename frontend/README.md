# MOH SSO Dashboard Frontend

This frontend is a workspace-based React/Vite application prepared for single-spa microfrontends. The default development and deployment path still works as one shell application, while standalone apps can also build independently and be staged under `dist/mf`.

## Architecture

```text
frontend/
  apps/
    shell/
    announcements/
    audit/
    clients/
    data-visualizer/
    documents/****
    e-services/
    email/
    issue-tracker/
    report-browser/
    surveillance/
    users/
    utilities/
  packages/
    api/
    auth/
    config/
    microfrontend/
    state/
    types/
    ui/
    utils/
```

- `apps/shell` owns layouts, route composition, auth guards, providers, and microfrontend orchestration.
- `apps/<app>` own product pages, feature components, nested app routes, and single-spa lifecycle entrypoints.
- `packages/<package>` own shared API hooks, auth/state contracts, UI, config, types, utilities, and microfrontend contracts.

More detail: [docs/architecture.md](docs/architecture.md).

## Prerequisites

Use Node `20.19+` or `22.12+` for Vite 7. Node 18 may still build in some environments, but Vite will warn.

```bash
cd frontend
npm install
```

## Development

Shell-only mode:

```bash
npm run dev:shell
```

Shell plus selected remote microfrontend dev servers:

```bash
npm run dev:remote:users
npm run dev:remote:clients
npm run dev:remote:documents
```

Shell plus all remote microfrontend dev servers:

```bash
npm run dev:remote:all
```

Local ports:

- shell: `3000`
- users: `4101`
- clients: `4102`
- announcements: `4103`
- audit: `4104`
- email: `4105`
- documents: `4106`
- surveillance: `4107`
- data-visualizer: `4108`
- issue-tracker: `4109`
- report-browser: `4110`
- e-services: `4111`
- utilities: `4112`

More detail: [docs/development.md](docs/development.md).

## Common Commands

```bash
npm run audit:packages
npm run typecheck
npm run lint
npm run build:shell
npm run build:packages
npm run build:apps
npm run build:all
npm run build:docker
```

`npm run audit:packages` verifies that every app/package declares the packages it imports. Run it after changing imports or package manifests.

## Docker

From the repository root:

```bash
docker build -f frontend/Dockerfile-nginx frontend
docker build -f frontend/Dockerfile.dev frontend
```

Docker dev defaults to shell-only mode. To run all remote dev servers in Docker:

```bash
FRONTEND_DEV_COMMAND="npm run dev:remote:all" docker compose -f docker-compose.dev.yml up frontend
```

## Deployment

`npm run build:docker` builds packages, generates the import map, builds apps, builds the shell, and stages microfrontend bundles into `dist`.

Final static layout:

```text
dist/index.html
dist/assets/*
dist/import-map.json
dist/mf/<app>/single-spa.js
dist/packages/<package>/index.js
```

More detail: [docs/build-and-deployment.md](docs/build-and-deployment.md).

## Documentation

- [Architecture](docs/architecture.md)
- [Development](docs/development.md)
- [Microfrontends](docs/microfrontends.md)
- [Import Maps](docs/import-maps.md)
- [Package Publishing](docs/package-publishing.md)
- [Build And Deployment](docs/build-and-deployment.md)
- [Verification](docs/verification.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Adding A New App](docs/adding-a-new-app.md)

## Troubleshooting

- Vite Node warning: use Node `20.19+` or `22.12+`.
- Missing dependency or app build failure: run `npm run audit:packages`.
- `npm ci` fails in Docker: run `npm install` locally and commit the updated `package-lock.json`.
- Remote app loads blank: check the import map URL, served `single-spa.js`, basename, and browser console.

More detail: [docs/troubleshooting.md](docs/troubleshooting.md).
