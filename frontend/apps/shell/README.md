# Shell App

The shell app is the frontend composition root for the MOH Integrated Health Portal.

It owns global routing, layouts, providers, route guards, runtime configuration, and microfrontend orchestration.

## Responsibilities

- Render public, user, and admin layouts.
- Register and mount standalone microfrontend apps.
- Provide auth bootstrap and route guards.
- Load runtime config and import maps in production.
- Compose shared navigation, headers, panels, modals, and toasts.

## Entry Points

- `src/main.tsx`
- `src/app/App.tsx`
- `src/app/routes`
- `src/app/microfrontends`

## Development

```bash
npm run dev -w @moh-sso/shell
npm run build -w @moh-sso/shell
```
