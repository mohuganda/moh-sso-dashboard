# Surveillance App

The surveillance app provides disease surveillance dashboards, imports, and related reporting views.

It is a standalone React microfrontend exposed as `@moh-sso/surveillance` and mounted by the shell at `/apps/dwh/surveillance`.

## Responsibilities

- Display surveillance dashboard views.
- Support disease details and surveillance metrics.
- Provide CSV upload/import workflows where available.
- Integrate with Data & Statistics navigation and RBAC permissions.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/surveillance
npm run build -w @moh-sso/surveillance
```
