# Microfrontend Strategy

The frontend is moving toward a single-spa-ready modular architecture while continuing to build as one Vite shell application.

## Current Phase

- Keep the shell as the single deployed frontend.
- Use workspace-local imports for microfrontend pilots.
- Expose single-spa-compatible lifecycles from pilot apps.
- Avoid import maps and remote bundles until the lifecycle pattern is stable.

## Orchestration Choice

Use single-spa as the future orchestration layer because it supports multiple frameworks through the same lifecycle contract:

- `bootstrap`
- `mount`
- `unmount`

React apps can use `single-spa-react` once the package is installed. Other frameworks can use their matching adapters later.

## Shared Dependency Strategy

For now, dependencies resolve from the workspace:

- `react`
- `react-dom`
- `@carbon/react`
- `@moh-sso/api`
- `@moh-sso/auth`
- `@moh-sso/state`
- `@moh-sso/ui`
- `@moh-sso/types`

When apps become independently deployed, these shared dependencies should move into import-map entries.

## Pilot App

The pilot app is `@moh-sso/users` because it has a clear route boundary at `/admin/users` and is less complex than documents, surveillance, or data visualizer.

## Future Work

- Install `single-spa` and `single-spa-react` with the package manager.
- Replace the manual lifecycle implementation with `single-spa-react`.
- Add import maps only when independent deployment begins.
- Repeat the lifecycle pattern feature-by-feature.
