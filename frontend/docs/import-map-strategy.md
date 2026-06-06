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
- `@moh-sso/api`
- `@moh-sso/auth`
- `@moh-sso/state`
- `@moh-sso/ui`
- `@moh-sso/types`

## Rollback

Rollback should be possible by repointing an app entry in the import map to a previous `single-spa.js` bundle.
