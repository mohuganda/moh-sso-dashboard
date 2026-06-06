# Import Map Strategy

Import maps are planned for independent microfrontend deployment, but they are not active in the current shell build.

## Current Mode

- The shell builds with Vite.
- Microfrontends resolve through workspace-local imports.
- `frontend/public/import-map.json` is a future deployment contract only.

## Future Local Development

Use import-map-overrides to point one app to a local dev bundle while the shell uses deployed versions for the rest.

Example:

```json
{
  "imports": {
    "@moh-sso/users": "http://localhost:4101/single-spa.js"
  }
}
```

Local development ports are reserved as follows:

- shell: `3000`
- users: `4101`
- clients: `4102`
- announcements: `4103`
- audit: `4104`
- email: `4105`
- documents: `4106`
- surveillance: `4107`
- data visualizer: `4108`
- issue tracker: `4109`
- report browser: `4110`
- e-services: `4111`
- utilities: `4112`

Use `public/import-map.local.json` as the starting point for local overrides.

## Staging And Production

Each environment should publish its own import map:

- local: points to local dev servers or local static bundles
- staging: points to staging CDN/app hosts
- production: points to versioned production CDN/app hosts

## Shared Dependencies

When remote loading is enabled, these should be shared import-map entries:

- `react`
- `react-dom`
- `single-spa`
- `react-redux`
- `@moh-sso/api`
- `@moh-sso/auth`
- `@moh-sso/config`
- `@moh-sso/microfrontend`
- `@moh-sso/state`
- `@moh-sso/ui`
- `@moh-sso/types`
- `@moh-sso/utils`

## Rollback

Rollback should be possible by repointing an app entry in the import map to a previous `single-spa.js` bundle.
