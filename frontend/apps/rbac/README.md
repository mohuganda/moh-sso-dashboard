# RBAC App

The RBAC app manages dynamic, system-aware authorization governance.

It is a standalone React microfrontend exposed as `@moh-sso/rbac` and mounted by the shell at `/admin/rbac`.

## Responsibilities

- Manage systems, roles, permissions, and role templates.
- Manage Keycloak group mappings and explain inherited group access.
- Inspect Keycloak-to-portal sync and drift reports.
- Support RBAC policy simulation and governance workflows.
- Provide administrative UI for system-aware access control.

## Groups

Keycloak remains the source of truth for groups and group membership. The portal syncs groups into the RBAC cache so admins can explain inherited access, assign portal-only group permissions, and audit group changes.

See [RBAC Groups](../../../docs/rbac-groups.md) for the operator workflow.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/rbac
npm run build -w @moh-sso/rbac
```
