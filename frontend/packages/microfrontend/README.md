# Microfrontend Package

`@moh-sso/microfrontend` contains shared contracts and helpers for single-spa-compatible apps.

## Responsibilities

- Define microfrontend runtime props and lifecycle types.
- Provide route metadata contracts.
- Provide React lifecycle helpers used by standalone apps.
- Keep microfrontend integration framework-neutral where possible.

## Usage

```ts
import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";
```

## Development

```bash
npm run build -w @moh-sso/microfrontend
```
