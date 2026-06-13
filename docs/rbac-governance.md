# RBAC Governance

The MOH SSO Dashboard uses Keycloak as the identity manager and keeps portal RBAC as the application authorization layer.

Keycloak owns:
- users
- realm roles
- client roles
- clients/systems
- authentication and tokens

Portal RBAC owns:
- portal permission keys
- system display metadata
- system ownership metadata
- mapping Keycloak realm roles to portal permissions
- mapping Keycloak client roles to portal permissions
- access request and RBAC change request governance
- drift detection, seed import/export, and environment promotion

## Onboarding A System

1. Create or confirm the Keycloak client.
2. Create client roles in Keycloak.
3. Run drift detection from `/admin/rbac` or the CLI.
4. Preview sync before applying it.
5. Apply sync to create inert system and role records in portal RBAC.
6. Fill system ownership metadata:
   - owner team
   - owner name
   - owner email
   - support URL
   - documentation URL
   - environment
   - criticality
7. Map portal permissions to system roles.
8. Use effective access preview to confirm the access model.

Newly discovered roles do not grant portal permissions until an admin maps them.

## Drift Detection

Use the admin UI or CLI to compare Keycloak with portal RBAC.

```bash
system-rbac sync-keycloak --realm-export ../keycloak/realm-export.json --draft-file /tmp/rbac-draft.yaml
system-rbac unmapped --realm-export ../keycloak/realm-export.json
```

The UI supports:
- live/current drift check
- pasted realm export drift check
- sync preview
- sync apply

## Import And Export

Export the current database state:

```bash
system-rbac export-db --file config/system-rbac.seed.yaml
```

Preview or apply imports from `/admin/rbac`.

Imports support YAML or JSON. Destructive pruning is intentionally disabled unless explicitly implemented and enabled later.

## Promotion Flow

Recommended flow:

1. Export from development:
   ```bash
   system-rbac export-db --file rbac-dev.yaml
   ```
2. Review changes in version control.
3. Compare the approved file with another environment:
   ```bash
   system-rbac diff --left rbac-staging.yaml --right rbac-dev.yaml
   ```
4. Dry-run promotion:
   ```bash
   system-rbac promote --file approved-rbac.yaml
   ```
5. Apply promotion:
   ```bash
   system-rbac promote --file approved-rbac.yaml --apply
   ```

## Approvals And Requests

Access requests let users request system roles. Admins can approve, reject, or cancel requests in the RBAC console.

RBAC change requests let admins stage sensitive changes for review. High-risk examples include:
- deleting a role
- disabling a system
- removing access roles
- bulk removing permissions
- importing destructive changes

Automatic Keycloak role assignment on access approval is a future integration point. Until configured, approved requests represent a governed decision and can be completed manually in Keycloak.

## Auditing

RBAC write operations emit audit events for:
- system metadata updates
- system roles
- access roles
- permission assignments
- realm role permission mappings
- sync apply
- seed import apply
- access request decisions
- RBAC change request decisions

The audit trail is available from `/admin/rbac`.

## Troubleshooting Access

Use Effective Access in `/admin/rbac` to explain access for a user.

Check:
- Keycloak realm roles
- Keycloak client roles
- portal permission mappings
- accessible systems
- grant sources for each permission

If a user has a Keycloak role but no portal access, confirm that the role is mapped to portal permissions and that the system is enabled.
