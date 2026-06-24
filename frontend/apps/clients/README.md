# Clients App

The clients app manages Keycloak clients, which represent integrated systems on the platform.

It is a standalone React microfrontend exposed as `@moh-sso/clients` and mounted by the shell at `/admin/clients`.

## Responsibilities

- List registered clients/systems.
- Create and update client metadata.
- Enable or disable clients.
- Manage client roles where supported.
- Use shared table, modal, toast, and header panel UI patterns.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/clients
npm run build -w @moh-sso/clients
```
