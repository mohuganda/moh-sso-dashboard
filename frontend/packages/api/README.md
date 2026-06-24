# API Package

`@moh-sso/api` contains shared frontend API clients and RTK Query endpoints.

## Responsibilities

- Provide the shared `baseApi`.
- Export API hooks used by apps and shell modules.
- Keep endpoint definitions out of feature UI where they are reused.
- Centralize request/response integration with backend APIs.

## Usage

Import API hooks from the package public entrypoint:

```ts
import { useListUsersQuery } from "@moh-sso/api";
```

## Development

```bash
npm run build -w @moh-sso/api
```
