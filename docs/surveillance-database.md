# Observatory database for surveillance

The surveillance module can use a separate PostgreSQL connection while authentication,
documents, processes, notifications and other features remain on the primary database.
An empty effective Observatory host preserves the previous single-database behavior.
A configured host requires its own complete connection settings; failure stops startup
and readiness fails if the configured surveillance database becomes unavailable.

## Relationship to the DQA connection

DQA uses `Config.DqaDbSource()` and the shared `db.InitDB` connection-pool
initializer. It targets the DWH host, port and database, optionally using its own
`DQA_DB_USERNAME`/`DQA_DB_PASSWORD`. When those are unset it uses DWH credentials.
`NewDQAStore` creates its `dqa` schema/tables idempotently on first use.

The observatory connection uses the same pool initializer and repository injection
pattern but keeps its own PostgreSQL container and database. Versioned surveillance
migrations remain in place. Authentication and other primary stores are unaffected.
`OBSERVATORY_DB_*` is the canonical connection namespace. Existing
`SURVEILLANCE_DB_*` files/environment remain supported as compatibility defaults;
explicit Observatory settings take precedence, including `ENABLE_SSL=false`. The
role defaults to the existing `surveillance` role; no database role is renamed. The container, role and existing data volume are retained; the local service now
shares the main development network. Existing explicit database names remain supported for rollback.

## Local container

The `observatory-db` service lives in `infra/compose/docker-compose.dev.yml`, alongside
`remote-postgres`. The backend uses `observatory-db:5432` on the shared `sso_network`
and waits for its health check. Starting the main development stack starts both databases.
A fresh stack creates the `observatory_data` named volume. Development-only default
credentials can be overridden using `OBSERVATORY_DB_USER`, `OBSERVATORY_DB_NAME`
and `OBSERVATORY_DB_PASSWORD`; legacy `SURVEILLANCE_DB_*` values remain supported.
The same values configure PostgreSQL and the backend.

```sh
docker compose -f infra/compose/docker-compose.dev.yml up -d --wait
```

The container remains `sso-surveillance-db` for the CSV loader. Host access is limited
to `127.0.0.1:55434`; `OBSERVATORY_DB_LOCAL_PORT` overrides that published port.
The database defaults to `observatory` and the existing role to `surveillance`.
A backend running directly on the host uses this loopback address/port instead.
Do not use `down -v`; it removes managed data volumes.

### Adopting the existing standalone volume

For an existing standalone container, back up the database first and inspect its
`/var/lib/postgresql/data` volume. Stop the backend and old database, then remove
only the old database container. Do not remove its volume. Use a private override
so the main project adopts the populated volume instead of creating an empty one:

```yaml
volumes:
  observatory_data:
    external: true
    name: moh-surveillance_surveillance_data
```

The current workspace keeps this override at `.local-sso/observatory-volume.override.json`
and credentials at `.local-sso/surveillance.env`, both mode 0600. Its saved backend
environment remains private in `.local-sso/backend-before-surveillance.override.json`.
Start the adopted service and backend with:

```sh
docker compose --env-file .local-sso/surveillance.env -f infra/compose/docker-compose.dev.yml -f .local-sso/backend-before-surveillance.override.json -f .local-sso/observatory-volume.override.json up -d --no-deps --wait observatory-db backend
```

Check readiness and compare all table counts before resuming writes. The main
project now manages both containers; the separate surveillance Compose files and
external network override are no longer needed. The existing data volume remains
external and is retained by Compose teardown. No production configuration is changed.

Load private variables into the environment without printing them, then, from
`backend/`, initialize the schema with `go run ./cmd/surveillance-db`. The backend
also applies the dedicated versioned migrations automatically. File-backed passwords
remain supported through `OBSERVATORY_DB_PASSWORD_FILE` and its legacy alias.

## Rename an existing local database

Changing `POSTGRES_DB` does not rename a database in an existing volume. For an
already populated container, take a backup, stop the local backend, and rename the
database while connected to `postgres`:

```sh
docker exec sso-surveillance-db psql -X -v ON_ERROR_STOP=1 -U surveillance -d postgres -c 'ALTER DATABASE surveillance RENAME TO observatory;'
```

All connections to the old database must be closed first. This preserves tables,
IDs, ownership, migration history, staging and facts; it does not reimport data.
Set the private `SURVEILLANCE_DB_NAME=observatory`, recreate the database container
without removing its volume to update its health check, then recreate the backend
using the updated database name. Compare table counts and verify readiness. Keep the
backup private. Rollback uses the inverse rename while the backend is stopped and
restores the prior connection name. Production database changes are separate from
this local procedure.

The CSV loader defaults to `observatory`; `--database surveillance` supports an
existing database before its rename, or an explicit alternate database during
verification. The PostgreSQL role remains `surveillance`.

## Table ownership

The isolated schema contains 13 tables: `regions`, `districts`, `sub_counties`,
`facilities`, `diseases`, `indicators`, `disease_indicators`, `epi_weeks`,
`facility_weekly_metrics`, `weekly_status`, `alerts`, `surveillance_import_batches`
and `surveillance_import_raw_rows`. It also owns `risk_level` and `alert_status`.
Staging retains its batch lookup index without a JSON GIN index, keeping bulk status
updates efficient. Subject/calendar reference seeds are included; locations come from imports or a data
transfer. The schema has its own migration history. Existing primary migrations remain
unchanged so upgrades and existing IDs are not silently invalidated.

`surveillance_import_batches.document_id` is an opaque ID into the primary documents
service, not a foreign key across databases. Documents/processes stay on the primary
connection while the import repository uses the surveillance connection.

## Import the root folder

```sh
python3 infra/scripts/load-surveillance.py --flag-year 2026
```

Each file gets a batch and each CSV record is retained as JSON in raw staging. The
loader uses a disk spool and PostgreSQL COPY, not one database request per source row.
The filename, SHA-256 checksum and explicit flag year identify a batch, preventing an
unchanged file from being staged again. `--stage-only` stops before transformation.
`--reprocess` transactionally rebuilds only the selected loader-owned source batches;
all source rows remain available for audit. Ordinary retries are idempotent.

- `data.csv`: location references and facility disease/indicator metrics. Original
  years are preserved. Exact known indicator names remain indicators; the suffix
  ` Cases` maps case labels to disease names. The loader does not sum malaria testing
  indicators into disease cases or invent epidemiological definitions.
- `eidsr.csv`: alerts, matching district names case-insensitively and deriving the
  year from the created date. Missing district/date/week/subject/narrative is held.
  Identical alert identities map to one alert while all source rows remain staged.
- `flags.csv`, `region_flags.csv`, `district_flags.csv`: national, regional and district
  weekly status. Each subject in a color list becomes a status. Lists are parsed with
  `ast.literal_eval`, never evaluated as executable code. `Maternal Death` is mapped
  to the seeded `Maternal Deaths` indicator; source text stays in raw JSON.

The agreed flag year is **2026**. Weeks **1–52 remain unchanged**. Source week **0**
is marked FAILED with a review reason and is never folded into week 1.

Identical facility/subject/week values collapse to one fact with a deterministic
source record, while each identical source row is counted as processed. If the values
conflict, **every member of that group is held for review**. Missing numeric values are
not treated as zero. Source row counts therefore need not equal normalized fact counts.
A batch can be COMPLETED with failed rows; inspect `success_rows`, `failed_rows`, and
raw-row error messages rather than treating batch status as full data acceptance.

This loader is an operational folder-import path. It does not repair the existing
surveillance UI's metadata-only upload action or its incomplete async CSV dispatch.
Those paths must be addressed separately before promising the same behavior for UI uploads.

## Existing database cutover

The local CSV import does not copy or delete existing production database records.
Before a real migration, inventory and back up all 13 tables, preserve their UUIDs,
restore them together into the new schema with constraints intact, and reconcile row
counts, foreign keys, batch outcomes and dashboard results. Seeded reference IDs in a
new database must not be mixed with old fact IDs. Pause surveillance writes during the
final copy/cutover, change only its connection, and retain the original tables for
rollback. Removal of legacy tables requires a separate reviewed migration.

## Verification

```sh
python3 infra/scripts/test-load-surveillance.py
SURVEILLANCE_DOCKER_TEST=1 python3 infra/scripts/test-surveillance-database.py
cd backend && go test ./...
```

The Docker test uses a disposable PostgreSQL container and synthetic data; it checks
schema independence from documents, all five file mappings, identical/conflicting
duplicates, zero-week quarantine, and repeat/reprocess behavior. It runs in PR CI.

## Verified local import (2026 flag year)

The initial import retained 796,427 source rows and produced 648,111 facility
metrics, 14,693 weekly status records and 579 alerts. 789,000 source rows were
processed; 7,427 were held for review: 6,771 conflicting metrics, 440 invalid
metrics, 51 alerts and 165 source-week-zero flag rows. There were no pending rows.
Identical observations and alert identities explain why processed source counts
are larger than normalized record counts. These are local fixture/import results,
not evidence of a production migration or cutover.

To temporarily use the primary database again, set the backend's explicit
`OBSERVATORY_DB_HOST` to an empty string in a private Compose override and recreate
only the backend. Restoring `observatory-db` reconnects the dedicated database.
Keep the adopted-volume override on subsequent local Compose commands.

## Observatory configuration names

Canonical settings are `OBSERVATORY_DB_HOST`, `OBSERVATORY_DB_PORT`,
`OBSERVATORY_DB_NAME`, `OBSERVATORY_DB_USER`, `OBSERVATORY_DB_PASSWORD`, and
`OBSERVATORY_DB_ENABLE_SSL`. Legacy `SURVEILLANCE_DB_*` values are used when the
corresponding canonical key is unset. An explicit empty canonical host disables
the isolated connection. A configured host requires all credentials; there is no
fallback to the primary database after a connection or migration failure.

The standalone `cmd/surveillance-db` initializer accepts either namespace. Its
command and the `internal/db/surveillance_migrations` directory retain their module
names; existing migration versions and data are preserved. The import integration
test accepts `OBSERVATORY_DOCKER_TEST=1` and the legacy CI flag
`SURVEILLANCE_DOCKER_TEST=1`. The loader's `--db-user` selects a non-default role
without renaming the role in an existing container.
