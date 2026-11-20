# Report Scheduler Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler`.
Go package: `reportscheduler`. Requires Go 1.25.4.

Version `v0.1.0` is the first standalone scheduler release. Its Go module tag is
`backend/modules/report-scheduler/v0.1.0`. It has independent dependency metadata
and tests and builds outside the dashboard tree. Hosts select the immutable
version in their dependency metadata and supply platform providers.

## Public API and responsibilities

| API | Responsibility |
| --- | --- |
| `Service`, `NewService` | Validate schedules and health scope, generate reports, retry, resolve recipients, deliver artifacts and record events |
| `Repository`, `NewRepository` | Persist scheduler state; optional PostgreSQL implementation against the host-owned schema |
| `HealthBIClient`, `NewKeycloakHealthBIClient` | Report catalogue, generation, polling and bounded artifact downloads using service-account tokens |
| `EmailDelivery`, `UserLookup`, `ArtifactStorage` | Host-supplied durable email outbox, identity directory and storage adapters |
| `Notifications`, `Audit` | Host-supplied event recording and failure notifications |
| `HealthHierarchy`, `NewPostgresHealthHierarchy` | Validate district/facility relationships against the host DWH |
| `WorkerService`, `Worker` | Explicit readiness, immediate processing, polling, cancellation and joined shutdown |
| `NewHTTPModule`, `HTTPOptions` | Package-owned protected and public routes, JSON responses and scheduler access policies |
| Request/result types | Public scheduling, execution, artifact, delivery and recipient models without host types |
| `CalculateNextRun`, `ResolveReportingPeriod` | Timezone-aware scheduling and reporting-period resolution |
| `ModulePath`, `Version` | Selected consumer dependency version; local/unversioned builds report `dev` |

The HTTP module exposes the existing `/report-scheduler` routes relative to the
host's API prefix. Mount `ProtectedPatterns()` with `ProtectedHandler()` behind
the host's authentication middleware. Mount `PublicPatterns()` with
`PublicHandler()` for anonymous delivery links. Strip the host API prefix from
the request before dispatch. The package enforces system access and per-route
permissions and passes verified identity and health scope to the service.
The host supplies a verified actor resolver and a per-IP download limiter.
The optional request-ID provider preserves upstream request correlation.
The package uses Go's standard HTTP library.

Construction does not start background work or run migrations. Construct the
service with explicit adapters, configure notifications/audit, public download URL,
delivery-link TTL and attachment limit, then create a worker. `Worker.Initialize`
returns after a successful persistence heartbeat with a ten-second readiness
deadline. Its cleanup cancels operations and waits for worker exit within the
supplied shutdown context. A worker instance starts once; construct a fresh worker
after a restart. `Worker.Start` is the blocking alternative for standalone callers.

Email attachments use base64 bytes so the host outbox can persist them. The host
adapter owns conversion to its email model. Health BI requests use client
credentials and do not need a user's session. Existing SQL claim exclusivity,
retry timing and stale polling/delivery claim recovery are preserved.

## Schema ownership

`schema-contract.json` records the scheduler's host-owned migration contract:
scheduler migrations 40 through 47, including the polling-status repair, must be
applied by the dashboard migration runner before starting a worker. Later host
migrations may also be present. The package contains no SQL migration files or
migration runner and does not advance host migration history.

Scheduler tables are `report_schedules`, `report_schedule_recipients`,
`report_executions`, `report_artifacts`, `report_deliveries` and
`report_scheduler_runtime`. The supplied PostgreSQL adapters also read the shared
`users`, `ihp_rbac_groups` and `dwh.dim_org_hierarchy` tables for directory and scope
validation. Hosts can supply alternative implementations through the ports.
Disabling the scheduler retains schedules, pending work and historical data.

## Verification

From this module directory:

```sh
GOWORK=off go test -race -mod=readonly ./...
GOWORK=off go vet -mod=readonly ./...
```

Database tests skip when their fixture URLs are absent. From the repository root,
run the complete isolated verification:

```sh
python3 infra/scripts/test-report-scheduler-package.py
```

It copies only this module to a temporary directory, disables Go workspaces,
checks its dependency boundary, builds/vets it, creates disposable PostgreSQL,
applies the host-owned scheduler migrations, and runs the suite under the race
detector. Database checks cover concurrent schedule/poll/retry/delivery claims,
stale-claim recovery, delivery idempotence, deletion lifecycle and recipient
directory behavior. Fixture connections replace inherited test database URLs.
No operator settings or running application databases are used.

The repository's `make -C backend modules-check` discovers this module alongside
Metrics. HTTP permissions, public token-download routes and registry rollback
remain covered by the dashboard suite and authenticated distribution fixture.

To integrate a release, verify a clean module download, select its version in the
host's `go.mod`, supply host adapters and report `Version()` in the compiled
inventory. Production builds use the selected dependency without local workspace
replacements. Schema changes continue through the host-owned migration runner.
