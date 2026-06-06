# Frontend Development

## Prerequisites

- Node `20.19+` or `22.12+`
- npm
- Docker, optional

Vite 7 warns on Node 18. Use Node 20.19+ or 22.12+ for the clean path.

## Install

```bash
cd frontend
npm install
```

## Shell-Only Development

```bash
npm run dev:shell
```

This starts the shell on port `3000` and resolves apps through the local workspace.

## Remote Microfrontend Development

Run shell plus selected app dev servers:

```bash
npm run dev:remote:users
npm run dev:remote:clients
npm run dev:remote:documents
```

Run shell plus all app dev servers:

```bash
npm run dev:remote:all
```

These commands set remote microfrontend mode for local testing.

## Ports

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

## Environment Variables

- `VITE_SINGLE_SPA_ORCHESTRATION`: set to `true` to enable single-spa registration.
- `VITE_MICROFRONTEND_MODE`: `local` uses workspace imports; `remote` uses import maps/runtime imports.
- `FRONTEND_DEV_COMMAND`: Docker dev command override.

## Docker Development

Shell-only:

```bash
docker compose -f docker-compose.dev.yml up frontend
```

All remote app dev servers:

```bash
FRONTEND_DEV_COMMAND="npm run dev:remote:all" docker compose -f docker-compose.dev.yml up frontend
```

## Dependency Audit

```bash
npm run audit:packages
```

Run this after:

- adding imports
- moving code between apps/packages
- editing any app/package `package.json`

## Daily Checks

Use these before opening a PR or handing off work:

```bash
npm run audit:packages
npm run typecheck
npm run lint
npm run build:shell
```

Run app/package builds when changing shared package exports, app entrypoints, route ownership, or package metadata.
