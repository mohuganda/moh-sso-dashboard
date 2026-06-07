# Package Publishing

Every frontend app and shared package has npm-style metadata so it can build and be published independently.

Changesets controls normal version bumps. Do not manually edit app/package versions unless doing an emergency/manual release.

## Package Requirements

Each app/package should have:

- `package.json`
- `version`
- `license`
- `exports`
- `publishConfig`
- `dependencies`
- `peerDependencies`
- `files` containing `dist`

Published packages should ship built `dist` artifacts, not source files.

Current internal version convention:

```text
@moh-sso/* -> 0.1.0
```

## Dependency Rules

- Shared platform dependencies are peer dependencies:
  - `react`
  - `react-dom`
  - `react-redux`
  - `react-router-dom`
  - `single-spa`
  - `single-spa-react`
  - `@carbon/react`
  - `@reduxjs/toolkit`
- Direct `@moh-sso/*` imports are dependencies.
- Type-only packages may be dev dependencies.

Run the audit after import or manifest changes:

```bash
npm run audit:packages
npm run audit:publishability
```

## Build A Package

```bash
npm run build -w @moh-sso/api
npm run build -w @moh-sso/ui
```

## Build An App

```bash
npm run build -w @moh-sso/users
npm run build -w @moh-sso/documents
```

## Publish Dry Run

Run release checks before any publish attempt:

```bash
npm run release:check
npm run audit:publishability
```

```bash
npm publish --dry-run -w @moh-sso/users
```

## Publish

```bash
npm publish -w @moh-sso/users
```

The preferred release command is:

```bash
npm run release:packages
```

## Registry And Auth

Configure npm before publishing:

```bash
npm login
npm config set registry <registry-url>
```

For private scoped packages, keep:

```json
{
  "publishConfig": {
    "access": "restricted"
  }
}
```

## Lockfile Rule

After changing any app/package `package.json`, update the lockfile:

```bash
npm install
```

Commit both `package.json` and `package-lock.json`. Docker uses `npm ci`, so stale lockfiles will fail builds.

## Changesets Commands

```bash
npm run changeset
npm run version:packages
npm run release:dry-run
npm run release:packages
```

Install Changesets first if needed:

```bash
npm install -D @changesets/cli
```

More detail: [Versioning And Releases](versioning.md).
