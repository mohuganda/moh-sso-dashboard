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
role defaults to the existing `surveillance` role; no database role is renamed. The container, role, network and existing data volume
are retained. Existing explicit database names remain supported for rollback.

## Local container

Use a private environment file based on `infra/env/surveillance.env.example`.
The current workspace configuration is `.local-sso/surveillance.env`, mode 0600;
credentials are not tracked. Start only the new database:

```sh
docker compose --env-file .local-sso/surveillance.env -f infra/compose/docker-compose.surveillance.yml up -d --wait
```

The container remains `sso-surveillance-db`, the database is `observatory`, the
database role remains `surveillance`, and
host access is restricted to `127.0.0.1:55434`. Its named volume persists data.
Do not use `down -v` to stop it; that removes its data volume.

Load the private variables into the process environment without printing them, then,
from `backend/`, initialize the schema with `go run ./cmd/surveillance-db`.
The backend also applies these migrations automatically when its dedicated connection
is configured. Runtime images include the dedicated migration directory.

The current local database has been renamed in place to `observatory`; database
and relation identifiers plus exact counts for all 13 tables were preserved. A
private pre-rename dump and count snapshot are retained under `.local-sso/`.

The current local development backend has been switched using
`docker-compose.surveillance-backend.override.yml`; its readiness check passed.
A private snapshot of its previous environment is stored in
`.local-sso/backend-before-surveillance.override.json` (mode 0600). Neither private
file belongs in Git. The standalone database and its import were verified after a
container recreation.

For a backend running directly on the host, use the private file's connection settings.
For a backend in Docker, attach it to the surveillance container's Docker network and
use `sso-surveillance-db:5432` instead of `127.0.0.1:55434`. Only enable the new connection
after importing or transferring and checking the data. File-backed passwords are
supported with `OBSERVATORY_DB_PASSWORD_FILE` (or the legacy
`SURVEILLANCE_DB_PASSWORD_FILE`) by the backend entrypoint.

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
using the updated override. Compare table counts and verify readiness. Keep the
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

To return the current local development backend to its saved primary-database settings:

```sh
docker compose -p moh-sso-dashboard-dev -f infra/compose/docker-compose.dev.yml -f .local-sso/backend-before-surveillance.override.json up -d --no-deps backend
```

To reapply the isolated connection, add `--env-file .local-sso/surveillance.env`
and `-f infra/compose/docker-compose.surveillance-backend.override.yml` to that
Compose command before `up`. This is local-only and does not alter production.

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
