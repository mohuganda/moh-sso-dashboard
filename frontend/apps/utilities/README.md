# Utilities App

The utilities app hosts self-service utility workflows for portal users.

It is a standalone React microfrontend exposed as `@moh-sso/utilities` and mounted by the shell under `/apps/utilities`.

## Responsibilities

- Provide utility and self-service workflow entry points.
- Support timesheet, e-learning, leave, absence, and related utility routes.
- Integrate with the shell side navigation and RBAC guards.
- Expose single-spa lifecycle functions for shell orchestration.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/utilities
npm run build -w @moh-sso/utilities
```
