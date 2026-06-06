# Changesets

Use Changesets to record public frontend package/app changes.

## Create A Changeset

```bash
npm run changeset
```

Select every package/app affected by the public change.

## Versioning Model

Shared packages are fixed together and release as one train:

- `@moh-sso/api`
- `@moh-sso/auth`
- `@moh-sso/config`
- `@moh-sso/microfrontend`
- `@moh-sso/state`
- `@moh-sso/types`
- `@moh-sso/ui`
- `@moh-sso/utils`

Apps version independently:

- `@moh-sso/shell`
- `@moh-sso/users`
- `@moh-sso/documents`
- other `@moh-sso/*` apps under `apps/`

## What Needs A Changeset

Create a changeset for:

- public exports
- runtime props or lifecycle contract changes
- route ownership changes
- API/type contract changes
- app-to-app dependency contract changes
- user-facing app features or fixes that should appear in release notes

Do not create a changeset for purely internal/no-public-effect edits unless release notes are useful.

## Release Commands

```bash
npm run release:check
npm run version:packages
npm run release:dry-run
npm run release:packages
```

Install Changesets first if the command is unavailable:

```bash
npm install -D @changesets/cli
```
