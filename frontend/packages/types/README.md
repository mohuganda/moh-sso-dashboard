# Types Package

`@moh-sso/types` contains shared TypeScript types used across frontend apps and packages.

## Responsibilities

- Define shared API and domain DTO types.
- Avoid duplicating type definitions across apps.
- Provide stable public type exports for publishable packages.

## Usage

```ts
import type { User, Client } from "@moh-sso/types";
```

## Development

```bash
npm run build -w @moh-sso/types
```
