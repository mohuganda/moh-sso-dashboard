# Announcements App

The announcements app manages platform notices, scheduled announcements, and announcement attachments.

It is a standalone React microfrontend exposed as `@moh-sso/announcements` and mounted by the shell at `/admin/announcements`.

## Responsibilities

- List and manage announcements.
- Create, update, publish, schedule, pin, and archive announcements.
- Manage announcement attachments used in news feed and email delivery.
- Expose single-spa lifecycle functions for shell orchestration.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/announcements
npm run build -w @moh-sso/announcements
```
