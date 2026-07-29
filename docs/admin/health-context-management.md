# Health Context Administration

Administrators need the corresponding permissions:

- `health_contexts:read`
- `health_contexts:write`
- `health_contexts:assign`
- `health_contexts:sync`
- `health_contexts:audit`

## API

Current hierarchy endpoints:

```text
GET    /api/v1/admin/health-contexts
POST   /api/v1/admin/health-contexts
GET    /api/v1/admin/health-contexts/:contextId
PUT    /api/v1/admin/health-contexts/:contextId
DELETE /api/v1/admin/health-contexts/:contextId
GET    /api/v1/admin/health-contexts/:contextId/children
POST   /api/v1/admin/health-contexts/:contextId/children
GET    /api/v1/admin/health-contexts/:contextId/descendants
GET    /api/v1/admin/health-contexts/:contextId/aliases
PUT    /api/v1/admin/health-contexts/:contextId/aliases
DELETE /api/v1/admin/health-contexts/:contextId/aliases/:aliasId
```

Assignment endpoints:

```text
GET /api/v1/admin/users/:userId/health-contexts
PUT /api/v1/admin/users/:userId/health-contexts
GET /api/v1/admin/rbac/groups/:groupId/health-contexts
PUT /api/v1/admin/rbac/groups/:groupId/health-contexts
```

Updates require the current `version`; stale updates return `409`. Reparenting
maintains the closure table and rejects cycles. Deletion returns `409` while a
node has children or direct/group assignments.

## User endpoints

```text
GET  /api/v1/me/health-contexts
GET  /api/v1/me/health-contexts/:contextId
GET  /api/v1/me/health-contexts/:contextId/explanation
POST /api/v1/me/health-contexts/select
```

Use direct assignments for exceptional individual scope. Prefer explicit
Keycloak group mappings for managed teams. Select one direct assignment as the
default where appropriate.

## Administration screen

RBAC Management includes a Health Context workspace for:

- browsing and searching hierarchy nodes;
- creating, editing, enabling, and disabling nodes;
- selecting parents without creating hierarchy cycles;
- assigning direct user scope;
- mapping Keycloak-backed groups to context scope;
- distinguishing direct and group-derived access;
- reviewing effective-access explanations; and
- previewing and applying group mapping drift reconciliation.

Destructive actions use shared confirmation modals. Mutations report through
the shared toast system. Drift application is preview-first and requires
`health_contexts:sync`.

Announcement and email audience editors can select one or more authorized
health contexts. Announcement authors can preview the recipient count before
publishing. The preview returns a count only and does not expose recipient
personal data.
