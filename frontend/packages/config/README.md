# Config Package

`@moh-sso/config` contains shared frontend configuration helpers and constants.

## Responsibilities

- Provide runtime and build-time configuration access.
- Keep shared constants in one package.
- Support shell and app configuration without importing shell internals.

## Usage

```ts
import { API } from "@moh-sso/config";
```

## Development

```bash
npm run build -w @moh-sso/config
```
