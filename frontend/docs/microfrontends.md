# Microfrontends

The frontend is single-spa-ready. The default path is hybrid local mode, where the shell builds normally and apps resolve through workspace imports. Remote/import-map mode is prepared for independent deployment.

## Current Modes

- Hybrid local mode: default. Shell uses local workspace app lifecycles.
- Remote mode: enabled by runtime/env config. Shell can load app bundles through import maps.

## App Contract

Every standalone app exposes:

```text
src/root.component.tsx
src/routes.tsx
src/single-spa.tsx
```

Every app lifecycle exports:

```ts
export async function bootstrap(props) {}
export async function mount(props) {}
export async function unmount(props) {}
```

React apps use `createReactMicrofrontendLifecycle` from `@moh-sso/microfrontend`.

## Runtime Props

Apps accept runtime props from `@moh-sso/microfrontend`:

- `basename`
- `auth`
- `apiBaseUrl`
- `eventBus`

`basename` is important for nested routes inside each app.

## Route Ownership

Examples:

- users: `/admin/users`
- clients: `/admin/clients`
- announcements: `/admin/announcements`
- audit: `/admin/audit-logs`
- email: `/admin/emails`

Documents owns:

- `/apps/dwh/filesvr`
- `/apps/utilities/self-service/eservice/document-upload`
- `/apps/utilities/self-service/eservice/document-upload/:id`

Surveillance owns:

- `/apps/dwh/surveillance`
- `/apps/dwh/surveillance/:diseaseName`

Utilities owns:

- `/apps/utilities/self-service/timesheet`
- `/apps/utilities/self-service/elearning`
- `/apps/utilities/self-service/leave-plan`
- `/apps/utilities/self-service/absence-requests`
- `/apps/utilities/self-service/absence-dashboard`

## Shell Integration

`SingleSpaApp` lives at:

```text
frontend/apps/shell/src/app/microfrontends/SingleSpaApp.tsx
```

It mounts lifecycle objects inside shell-owned React Router routes.

The orchestrator lives at:

```text
frontend/apps/shell/src/app/microfrontends/orchestrator.ts
```

Related files:

- `registerApps.ts`: app route registry
- `routes.ts`: route constants
- `lifecycles.ts`: local lifecycle imports

## Local Vs Remote

Local mode uses workspace imports and local lifecycle modules.

Remote mode uses import maps. The shell loads `@moh-sso/<app>` bundle URLs from the import map.

## Non-React Apps

Future Vue, Angular, Svelte, or plain JavaScript apps can participate by exposing the same lifecycle contract:

- `bootstrap`
- `mount`
- `unmount`

They should accept the same runtime props and avoid importing shell internals.
