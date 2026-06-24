# Utils Package

`@moh-sso/utils` contains small shared utility functions used by frontend apps and packages.

## Responsibilities

- Provide generic reusable helpers.
- Avoid large feature-specific utility dumping.
- Keep feature-owned helpers inside their app when they are not truly shared.

## Usage

```ts
import { formatDateTime } from "@moh-sso/utils";
```

## Development

```bash
npm run build -w @moh-sso/utils
```
