# Documents App

The documents app handles document upload and document processing workflows.

It is a standalone React microfrontend exposed as `@moh-sso/documents` and mounted by the shell at document-related routes such as `/apps/dwh/filesvr`.

## Responsibilities

- Upload files and documents.
- Display document records and details.
- Support document processing flows.
- Provide document upload routes used by DWH and self-service workflows.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/documents
npm run build -w @moh-sso/documents
```
