# E-Services App

The e-services app is the product module for electronic service workflows.

It is a standalone React microfrontend exposed as `@moh-sso/e-services` and mounted by the shell at `/apps/eservices`.

## Responsibilities

- Provide the e-services landing experience.
- Host service request workflows as they are implemented.
- Integrate with shell routing and RBAC guards.
- Expose single-spa lifecycle functions for future independent deployment.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/e-services
npm run build -w @moh-sso/e-services
```
