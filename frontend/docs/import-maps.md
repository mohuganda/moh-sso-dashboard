# Import Maps

Import maps prepare the shell for independently deployed microfrontends.

## Files

```text
frontend/public/import-map.json
frontend/public/import-map.local.json
```

`import-map.json` is the production-style map. `import-map.local.json` maps apps to local dev server ports.

## Vendor Dependencies

The production import map uses pinned ESM CDN URLs for shared platform dependencies:

- `react`
- `react-dom`
- `react-dom/client`
- `react-redux`
- `single-spa`
- `single-spa-react`

This avoids missing `/vendor/*.js` files and keeps React shared in remote mode.

## App And Package Entries

App bundles:

```text
/mf/<app>/single-spa.js
```

Shared packages:

```text
/packages/<package>/index.js
```

## Versioned Import Map Mode

Default mode writes local/static paths:

```text
/mf/users/single-spa.js
/packages/api/index.js
```

Versioned mode writes CDN/base URLs with package versions:

```bash
FRONTEND_ASSET_BASE_URL=https://cdn.example.com FRONTEND_VERSIONED_IMPORTS=true npm run generate:import-map
```

Example:

```json
{
  "imports": {
    "@moh-sso/users": "https://cdn.example.com/mf/users/0.4.1/single-spa.js",
    "@moh-sso/documents": "https://cdn.example.com/mf/documents/0.8.0/single-spa.js",
    "@moh-sso/api": "https://cdn.example.com/packages/api/0.2.0/index.js"
  }
}
```

Environment variables:

- `FRONTEND_VERSIONED_IMPORTS=true`
- `FRONTEND_ASSET_BASE_URL=https://cdn.example.com`

## Local Import Map

The local map points apps to dev servers:

- users: `http://localhost:4101/single-spa.js`
- clients: `http://localhost:4102/single-spa.js`
- announcements: `http://localhost:4103/single-spa.js`
- audit: `http://localhost:4104/single-spa.js`
- email: `http://localhost:4105/single-spa.js`
- documents: `http://localhost:4106/single-spa.js`
- surveillance: `http://localhost:4107/single-spa.js`
- data-visualizer: `http://localhost:4108/single-spa.js`
- issue-tracker: `http://localhost:4109/single-spa.js`
- report-browser: `http://localhost:4110/single-spa.js`
- e-services: `http://localhost:4111/single-spa.js`
- utilities: `http://localhost:4112/single-spa.js`

## Import Map Overrides

`import-map-overrides` can replace one deployed app with a local dev bundle. This is useful when testing a single app against deployed shell/config.

## Enabling Remote Mode

The shell `index.html` keeps import maps passive by default. To enable remote mode in a deployment template:

1. Set `window.__APP_CONFIG__.singleSpaOrchestration = true`.
2. Set `window.__APP_CONFIG__.microfrontendMode = "remote"`.
3. Enable the import map script in the shell HTML or deployment template:

```html
<script type="importmap" src="/import-map.json"></script>
```

4. Optionally enable import-map-overrides:

```html
<script async src="https://unpkg.com/import-map-overrides@6.1.0/dist/import-map-overrides.js"></script>
```

## Rollback

Rollback is import-map driven:

1. Point the affected app entry back to a previous versioned bundle URL.
2. Redeploy the import map or shell config.
3. Clear CDN/browser cache if required.

## Cache Busting

Use one of these strategies:

- versioned bundle URLs
- hashed deployment folders
- immutable CDN assets plus mutable import map

## CDN Vs Self-Hosted Vendors

The current map uses pinned CDN ESM URLs for simplicity. If an offline or fully self-hosted deployment is required, replace vendor entries with self-hosted files and ensure those files are served with the shell.
