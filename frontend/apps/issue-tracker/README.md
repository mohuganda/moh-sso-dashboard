# Issue Tracker App

The issue tracker app supports tracking and reviewing data quality issues.

It is a standalone React microfrontend exposed as `@moh-sso/issue-tracker` and mounted by the shell at `/apps/dwh/issue-tracker`.

## Responsibilities

- Display issue tracking workflows for DWH/data quality users.
- Support investigation and resolution of flagged records.
- Integrate with shared API, UI, and type packages.
- Expose single-spa lifecycle functions for shell orchestration.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/issue-tracker
npm run build -w @moh-sso/issue-tracker
```
