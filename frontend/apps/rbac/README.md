# RBAC App

The RBAC app manages dynamic, system-aware authorization governance.

It is a standalone React microfrontend exposed as `@moh-sso/rbac` and mounted by the shell at `/admin/rbac`.

## Responsibilities

- Manage systems, roles, permissions, and role templates.
- Inspect Keycloak-to-portal sync and drift reports.
- Support RBAC policy simulation and governance workflows.
- Provide administrative UI for system-aware access control.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/rbac
npm run build -w @moh-sso/rbac
```
