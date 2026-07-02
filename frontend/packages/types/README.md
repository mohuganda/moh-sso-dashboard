# Types Package

`@moh-sso/types` contains cross-cutting TypeScript contracts used by shared frontend
packages.

## Responsibilities

- Define shared contracts that packages like `@moh-sso/ui`, `@moh-sso/state`,
  `@moh-sso/config`, and `@moh-sso/utils` can safely consume.
- Keep feature-specific API DTOs inside the owning app under `src/types`.
- Avoid dependencies from shared packages back into feature apps.
- Provide stable public type exports for publishable shared packages.

## Usage

```ts
import type { Client, Notification } from "@moh-sso/types";
```

Feature-specific types should be imported from the owning module, for example:

```ts
import type { User } from "@moh-sso/users";
import type { Announcement } from "@moh-sso/announcements";
```

## Development

```bash
npm run build -w @moh-sso/types
```
