# Email App

The email app manages email outbox and delivery operations for platform messaging.

It is a standalone React microfrontend exposed as `@moh-sso/email` and mounted by the shell at `/admin/emails`.

## Responsibilities

- Display email outbox messages.
- Send and inspect email delivery records.
- Support attachments and scheduled email workflows as backend capabilities evolve.
- Use shared panel, toast, and table UI patterns.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/email
npm run build -w @moh-sso/email
```
