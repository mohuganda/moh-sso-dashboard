# Identity environments and client onboarding

The portal manages one configured realm per deployment. Its credentials never select
an arbitrary realm from a browser request. Onboarding is blocked when the deployment
and realm do not match `backend/internal/keycloak/identity-environments.json`:

| Portal ENVIRONMENT | KEYCLOAK_REALM | Intended hosting |
| --- | --- | --- |
| development | moh-dev | Local isolated identity stack |
| test | moh-test | Disposable CI container |
| staging | moh-staging | Persistent non-production stack |
| production | moh-realm | Existing production instance; dedicated database |

The production realm name stays unchanged to preserve its issuer and integrations.
Different realms on a shared server separate identities, but still share its failure
and administrative boundaries. Keep production on separate infrastructure. Use a
separate Compose project, database volume, credentials, hostname and port for each
persistent environment. Do not connect non-production applications to production
identity databases. CI uses disposable fixtures with no production credentials.

## Bootstrap a fresh identity stack

Generate an environment-specific import file. Substitute the real staging portal
and API URLs; the API URL includes `/api/v1`:

```sh
python3 infra/scripts/prepare-keycloak-realm.py \
  --environment staging \
  --portal-url https://portal-staging.example.org/portal \
  --api-url https://portal-staging.example.org/api/v1 \
  --output /private/identity/moh-staging.json
```

Create the output directory first. The generator refuses to overwrite an existing
file. It preserves platform role/scope definitions and service-account role bindings,
removes human accounts and all credentials/IDs, and excludes external application
clients. Keycloak generates fresh client credentials and signing keys. External apps
must be onboarded independently in each environment. Only core portal clients start
enabled. Other retained platform client definitions remain disabled and have their
old browser URLs removed until configured for this environment.

Copy `infra/env/identity.env.example` to a private file outside the repository. Set a
unique project, identity hostname, loopback port, generated import file path and
private database/bootstrap passwords. Then:

```sh
docker compose --env-file /private/identity/staging.env \
  -f infra/compose/docker-compose.identity.yml config --quiet
docker compose --env-file /private/identity/staging.env \
  -f infra/compose/docker-compose.identity.yml up -d
```

The stack uses Keycloak production mode and PostgreSQL, binds HTTP to loopback and
expects a trusted HTTPS reverse proxy at `IDENTITY_HOSTNAME`. Route that environment's
hostname to its loopback port; forward the configured proxy headers. Avoid routing
untrusted traffic directly to the internal listener. For local development use a
loopback HTTP hostname and generate the realm with loopback portal/API URLs.

After first boot, set environment-specific secrets for the portal's existing
administrative and confidential web clients in the Keycloak admin console. Store
these in that portal deployment's private environment. Set its `ENVIRONMENT`,
`KEYCLOAK_REALM`, internal URL, external URL, callback and web/admin client credentials
consistently. Each application must validate its own realm's exact issuer. Configure
SMTP separately; staging email should go to a test inbox/sink. Provision human
administrators explicitly and grant realm-specific administration permissions.

These commands prepare fresh installations. They do not replace live realms or
migrate users. Startup import skips existing realms. The existing local dev Compose
stack is retained for compatibility; opting into `moh-dev` requires configuring that
portal deployment to use the new identity stack. Do not point both realm names at
the same application's configuration or reuse old cookies/sessions after switching.

## Admin console walkthrough

Open **Clients → Onboard application** (requires `clients:write`). Register the
application name, owner, support email and client profile. Review callbacks, logout
addresses, browser origins and the optional external launcher entry. Select
**Validate & preview**, inspect the environment and configuration, confirm review,
then provision the disabled client. The result provides its issuer/client ID and
next steps. **Export onboarding draft** produces a credential-free JSON document.
For managed disabled clients, **Edit** opens the same reviewed update workflow;
existing unmanaged clients retain their normal editor.

## Client lifecycle and promotion

The Clients administration workflow uses the same public-web, confidential-web and
service profiles as the operator CLI. Browser profiles use authorization code/S256,
RS256, `profile email`, stable subject/email claims, and exact callbacks/logout/origins.
Service profiles use client credentials with no browser flow or automatic role grants.
All new or updated onboarded clients stay disabled. Access mapping and activation
are explicit follow-up steps; configuring a launcher entry does not grant access.

Export an onboarding draft from the source deployment, import it into the target,
review/change its environment and all URLs, preview the result, then provision it.
Only the configuration is portable: users, credentials, signing keys and sessions
are never included. An existing target client must be owned by the same onboarding
profile and be disabled before an update. The preview shows its previous settings;
changes to the draft or target state require a new review. Audit records contain the
client ID, environment, realm, profile and operation, never credentials.

For an operator-assisted promotion, use the exported version-1 document:

```sh
python3 infra/scripts/reconcile-keycloak-clients.py \
  --environment staging --spec /private/identity/reviewed-app.json
python3 infra/scripts/reconcile-keycloak-clients.py \
  --environment staging --spec /private/identity/reviewed-app.json \
  --apply --backup /private/identity/app-before.json
```

Load only that environment's existing `KEYCLOAK_*` administrative credentials. The
CLI refuses realm/environment mismatches, cross-type client conversion and updates
to enabled or unmanaged clients. Apply requires a new mode-0600 backup. Provisioning
is repeatable; managed scopes/mappers are checked after apply. If a partial create
fails, leave it disabled, inspect the next preview and retry. For recovery restore
only the affected client's saved settings/scopes/mappers, never an entire realm.

Before enabling: test code exchange, wrong/missing PKCE, logout, issuer/azp validation,
email-verification refusal and portal access mapping where applicable. Confidential
client secrets are retrieved/rotated privately through the realm's administration;
this workflow never exports them. Promote configuration through development → test
→ staging → production with the existing production review/release process.
