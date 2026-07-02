# Data Visualizer App

The data visualizer app supports exploratory data viewing and charting for platform datasets.

It is a standalone React microfrontend exposed as `@moh-sso/data-visualizer` and mounted by the shell at `/apps/dwh/data-visualizer`.

## Responsibilities

- Display data tables and visualizations.
- Provide data model, period, organisation unit, and visualization controls.
- Support DWH-style analysis workflows.
- Expose single-spa lifecycle functions for shell orchestration.

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev -w @moh-sso/data-visualizer
npm run build -w @moh-sso/data-visualizer
```
