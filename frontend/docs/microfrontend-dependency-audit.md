# Microfrontend Dependency Audit

This audit tracks direct cross-app imports and whether they should remain.

## Current Intentional Dependencies

- `users` imports `clients` role assignment UI.
- `home` is shell-owned and composes panels from users, clients, and announcements.
- `issue-tracker` imports data visualizer utilities and table components.
- `surveillance` imports document API/types for disease detail documents.

## Target Direction

- Shell should import apps through lifecycle modules or public index exports.
- Apps should avoid importing shell internals.
- Reusable components should move to `frontend/packages/ui`.
- Shared types should move to `frontend/packages/types`.
- Shared API hooks should stay in `frontend/packages/api`.

## Next Cleanup Candidates

- Move user/client role assignment UI into a shared admin package or `packages/ui`.
- Move data visualizer table utilities used by issue tracker into `packages/ui` or `packages/utils`.
- Keep surveillance/document API coupling documented until domain ownership is clearer.
