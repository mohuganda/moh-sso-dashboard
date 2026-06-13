# System Onboarding

Use this flow when a new Keycloak client/system is added to the Integrated Health Portal.

## 1. Create or Confirm Keycloak Client

Create the client in Keycloak using the admin UI or CLI. Add the client roles that represent the system's access model.

Examples:

- `<system>_access`
- `viewer`
- `data_entry`
- `manager`
- `admin`

## 2. Draft Portal RBAC Mapping

From `backend/`:

```sh
go run ./cmd/cli system-rbac sync-keycloak \
  --base-url "$KEYCLOAK_BASE_URL" \
  --realm "$KEYCLOAK_REALM" \
  --admin-client-id "$KEYCLOAK_ADMIN_CLIENT_ID" \
  --admin-client-secret "$KEYCLOAK_ADMIN_CLIENT_SECRET" \
  --draft-file /tmp/system-rbac.seed.yaml
```

Review the draft and assign permissions to each system role. A generated role without permissions is intentionally inert until mapped.

## 3. Validate the Seed

```sh
go run ./cmd/cli system-rbac validate --file /tmp/system-rbac.seed.yaml
```

## 4. Apply to the Portal Database

```sh
go run ./cmd/cli system-rbac seed --file /tmp/system-rbac.seed.yaml
```

## 5. Manage or Adjust in the Admin UI

After the first seed, portal administrators can manage system metadata, access roles, system roles, and permission mappings at:

```text
/admin/rbac
```

The signed-in administrator needs `rbac:read` to open the page. Write actions require `rbac:write`, `rbac:roles:write`, or `rbac:permissions:write` depending on the operation.

## 6. Explain a User's Expected Access

```sh
go run ./cmd/cli system-rbac explain \
  --realm-role user \
  --client-role integrated-outbreak-system:viewer
```

## 7. Check for Drift

```sh
go run ./cmd/cli system-rbac unmapped \
  --base-url "$KEYCLOAK_BASE_URL" \
  --realm "$KEYCLOAK_REALM" \
  --admin-client-id "$KEYCLOAK_ADMIN_CLIENT_ID" \
  --admin-client-secret "$KEYCLOAK_ADMIN_CLIENT_SECRET" \
  --file config/system-rbac.seed.yaml
```

Any unmapped client or role should either be added to the seed or intentionally ignored in an onboarding note.
