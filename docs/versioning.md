# Versioning

## Backend

The backend follows Semantic Versioning and uses component-scoped Git tags:

```text
backend/v1.0.0
backend/v1.1.0
backend/v2.0.0
```

- PATCH: compatible fixes with no contract or migration break.
- MINOR: backward-compatible functionality.
- MAJOR: breaking API, configuration, database migration, security, or operational contract changes.

Frontend workspace packages and apps have their own npm versions. A backend tag never changes frontend package versions automatically.

## Build Metadata

Every backend binary exposes:

- service name
- normalized version without a leading `v`
- Git commit
- RFC3339 UTC build time
- dirty worktree status
- Go runtime version

Inspect it with:

```bash
./bin/moh-sso-dashboard --version
./bin/moh-sso version --output json
curl http://localhost:9000/version
```

Development builds default to `dev` and use Go embedded VCS information when available. Release values are injected with linker flags and take precedence over VCS fallbacks.

## Image Tags

Backend images use tags without the component prefix:

```text
ghcr.io/mohuganda/moh-sso-dashboard-backend:1.2.3
ghcr.io/mohuganda/moh-sso-dashboard-backend:sha-d3062e2
```

`latest` may exist for discovery on the default branch but is forbidden for production deployment. Production records the immutable tag and image digest.

## Compatibility

Minor and patch releases should preserve API and configuration compatibility. Breaking response shapes, removed endpoints, mandatory configuration changes, destructive migrations, or authorization semantic changes require a major release or a documented compatibility window.

Database migrations must remain forward-safe for rolling deployment whenever possible. A release with an irreversible migration must document backup and rollback constraints before it is tagged.
