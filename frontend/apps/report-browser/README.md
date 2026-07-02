# Report Browser App

The report browser app provides access to platform and DWH reports.

It is a standalone React microfrontend exposed as `@moh-sso/report-browser` and mounted by the shell at `/apps/dwh/dashboards`.

## Responsibilities

- Provide report browsing entry points.
- Support report access controlled by RBAC and system launch permissions.
- Integrate with the Data & Statistics navigation group.
- Expose single-spa lifecycle functions for shell orchestration.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/report-browser
npm run build -w @moh-sso/report-browser
```
