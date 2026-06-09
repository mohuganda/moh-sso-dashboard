# MOH SSO UI Package

## Rules

- Do not import from `@moh-sso/api`, `@moh-sso/state`, or app folders.
- Components should be presentation-only.
- Business data fetching stays in apps.
- Use Carbon tokens and MOH CSS variables.
- Avoid inline styles.
- Use `PageShell`, `SectionCard`, `EmptyState`, and `ErrorState` for consistency.