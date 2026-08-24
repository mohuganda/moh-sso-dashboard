# Publishing Frontend Packages

The frontend publishes independently consumable apps and shared libraries under
the private npm scope `@moh-sso`. The shell is deliberately private and is
distributed only through the frontend runtime image.

## Published Workspaces

The release contains 22 packages:

- 14 single-spa apps from `apps/*`, excluding `apps/shell`
- 8 shared packages from `packages/*`

Every published package contains compiled `dist` artifacts, declarations,
README documentation, and package metadata. Source files, local configuration,
credentials, and development artifacts are excluded.

## npm Organization Setup

Before the first release:

1. Create or confirm the npm organization named `moh-sso`.
2. Grant maintainers permission to publish restricted packages.
3. Require two-factor authentication for organization members.
4. Create the GitHub environment `npm-production` and protect it with required
   reviewers.
5. Configure npm trusted publishing for
   `mohuganda/moh-sso-dashboard` and workflow
   `.github/workflows/frontend-packages-release.yml`.

Trusted publishing is preferred in CI. Do not commit npm tokens or a populated
user `.npmrc`.

The repository `.npmrc` contains only:

```ini
@moh-sso:registry=https://registry.npmjs.org/
```

Local credentials stay in `~/.npmrc`:

```ini
//registry.npmjs.org/:_authToken=${NPM_TOKEN}
```

## First Package Bootstrap

npm trusted publishing may require each package to exist before its trusted
publisher can be configured. For the first release only:

1. Authenticate interactively with an authorized maintainer account.
2. Run the full validation and tarball checks.
3. Publish each package as restricted.
4. Configure its trusted publisher on npm.
5. Use CI for all subsequent releases.

Never publish `@moh-sso/shell`.

## Create A Changeset

From `frontend/`:

```bash
source ~/.nvm/nvm.sh
nvm use 20.20.1
npm ci
npm run changeset
```

Select only apps and packages with a public change. Shared packages are a fixed
release train, so Changesets keeps their versions aligned.

For repository-only work that intentionally does not release a package, create
an explicit empty changeset:

```bash
npx changeset --empty
```

## Validate A Release

```bash
npm run release:check
npm run verify:consumer-install
```

The checks validate dependency declarations, package metadata, compatibility,
types, lint, production artifacts, package contents, and installation into a
clean consumer project.

Useful focused commands:

```bash
npm run audit:packages
npm run audit:publishability
npm run compatibility:check
npm run changeset:status
npm run build:all
npm run pack:all
npm run verify:tarballs
npm run verify:consumer-install
```

Tarballs are written to `frontend/.artifacts/npm/`. This directory is a local
verification artifact and must not be committed.

Tarballs default to a 10 MiB compressed and 40 MiB unpacked limit. Intentional
exceptions must be reviewed and can be tested without changing source:

```bash
NPM_PACKAGE_MAX_TARBALL_BYTES=15728640 \
NPM_PACKAGE_MAX_UNPACKED_BYTES=52428800 \
npm run verify:tarballs
```

## Automated Release Flow

On `main`, `Frontend Packages Release`:

1. validates pull requests and every publishable workspace with
   `npm run release:check`;
2. opens or updates a Changesets version PR;
3. publishes after the version PR is merged;
4. uses npm trusted publishing through OIDC.

The workflow publishes only from `mohuganda/moh-sso-dashboard`. Forks cannot
publish packages.

## Manual Emergency Release

Normal releases must use CI. For an approved emergency release:

```bash
npm run release:check
npm run version:packages
git commit -am "chore(frontend): release npm packages"
NPM_TAG=latest npm run release:packages
```

`release:packages` refuses dirty worktrees, non-main CI branches, non-canonical
CI repositories, and prerelease versions sent to `latest`.

Use a prerelease tag for preview versions:

```bash
NPM_TAG=next npm run release:packages
```

## Installing Packages

For an authenticated consumer:

```bash
npm install @moh-sso/ui@1.2.3
npm install @moh-sso/users@1.2.3
```

Apps expose:

```text
@moh-sso/<app>
@moh-sso/<app>/single-spa
@moh-sso/<app>/routes
```

Shared packages expose only their documented public entrypoints. Consumers must
not import package internals.

## Rollback And Deprecation

npm versions are immutable. Never overwrite or unpublish a production version.

- Roll back a runtime by rebuilding with the previous exact package versions.
- Deprecate a broken version:

```bash
npm deprecate @moh-sso/users@1.2.3 "Use 1.2.4"
```

- Publish a patch release for the fix.
- Keep the previous frontend Docker image available for runtime rollback.

See [Module Source Strategy](module-source-strategy.md) and
[Versioning And Releases](versioning.md).
