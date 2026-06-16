# Auth Package

`@moh-sso/auth` contains authentication state, authorization helpers, permission constants, and route guard support.

## Responsibilities

- Store authenticated user state.
- Expose selectors and auth bootstrap logic.
- Provide permission and system constants.
- Provide authorization helpers used by shell routes and app navigation.

## Usage

```ts
import { PERMISSIONS, useAuthorization } from "@moh-sso/auth";
```

## Development

```bash
npm run build -w @moh-sso/auth
```
