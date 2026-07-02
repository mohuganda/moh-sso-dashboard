# API Package

`@moh-sso/api` contains the shared RTK Query base API used by frontend apps.

## Responsibilities

- Provide the shared `baseApi`.
- Centralize request credentials, refresh handling, and backend base URL behavior.
- Keep feature endpoint definitions out of the shared package.
- Let each app inject its own endpoints into `baseApi` from its local `src/api` folder.

## Usage

Feature modules should import `baseApi`, inject endpoints locally, and export their hooks from
their own app package.

```ts
import { baseApi } from "@moh-sso/api";

export const usersApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    // feature-owned endpoints
  }),
});
```

Consumers should import hooks from the owning module, for example `@moh-sso/users`,
`@moh-sso/clients`, or a local `src/api` barrel inside that module.

## Development

```bash
npm run build -w @moh-sso/api
```
