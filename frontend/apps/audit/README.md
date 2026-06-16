# Audit App

The audit app provides administrative visibility into platform activity and security events.

It is a standalone React microfrontend exposed as `@moh-sso/audit` and mounted by the shell at `/admin/audit-logs`.

## Responsibilities

- Display audit logs and audit metrics.
- Filter logs by time range, actor, client, action, IP, and result.
- Export audit log data.
- Open audit log details in the shared header panel.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/audit
npm run build -w @moh-sso/audit
```
