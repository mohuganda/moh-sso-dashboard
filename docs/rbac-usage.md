# RBAC Usage

The dashboard uses dynamic, system-aware RBAC.

## Source of Truth

The frontend reads access from `/auth/me`:

- `permissions`
- `systems`
- `accessibleSystems`
- `realmRoles`
- `clientRoles`

Realm roles and client roles are inputs. Permission checks are the preferred authorization model.

## Backend Rules

Backend routes must enforce permissions with middleware from `backend/internal/middleware`.

Use:

- `RequirePermission(permission)` for one required permission.
- `RequireAnyPermission(...)` when several read permissions can open a shared API.
- `RequireAllPermissions(...)` when all listed permissions are required.
- `RequireSystem(systemClientID)` for system-scoped modules.
- `RequireSystemRole(systemClientID, role)` only when behavior is truly role-specific.

Do not protect new APIs with raw role checks unless the business rule is specifically role-based.

## Frontend Rules

Use helpers from `@moh-sso/auth`:

- `hasPermission`
- `hasAnyPermission`
- `hasAllPermissions`
- `hasSystem`
- `hasSystemRole`
- `canLaunchSystem`
- `getAccessibleSystems`

Use `PermissionRoute` for shell route boundaries. Use `PermissionGuard` for page actions and buttons.

## Microfrontend Metadata

Every standalone app route should declare access metadata in `frontend/apps/<app>/src/routes.tsx`.

Example:

```ts
export const surveillanceRoute = {
  appName: "@moh-sso/surveillance",
  path: "/apps/dwh/surveillance",
  requiredAnyPermissions: ["surveillance:read", "outbreak:access"],
  requiredSystems: ["integrated-outbreak-system"],
};
```

The shell uses this metadata to guard routes. `npm run rbac:doctor` checks that route metadata exists.

## System Launching

App launcher and side navigation should use:

- `systems:launch`
- `accessibleSystems`

A Keycloak client should not appear only because it exists. It must be mapped through portal RBAC and assigned to the user.

## Debugging

Use the backend CLI to explain expected access:

```sh
cd backend
GOCACHE=/private/tmp/moh-sso-go-build go run ./cmd/cli system-rbac explain \
  --realm-role user \
  --client-role integrated-outbreak-system:viewer
```

Use frontend checks:

```sh
cd frontend
source ~/.nvm/nvm.sh
nvm use v20.20.1
npm run rbac:doctor
```
