# Data Validation App

The data validation app manages validation rules used for data quality checks.

It is a standalone React microfrontend exposed as `@moh-sso/data-validation` and mounted by the shell at `/apps/dwh/data-validation`.

## Responsibilities

- Display built-in and custom validation rules.
- Provide rule search and pagination.
- Add custom validation rules through the shared header panel.
- Keep the table and panel UX consistent with Users, Clients, and Audit.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/data-validation
npm run build -w @moh-sso/data-validation
```
