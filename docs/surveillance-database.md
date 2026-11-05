# Isolated surveillance database

The surveillance module can use a separate PostgreSQL connection while authentication,
documents, processes, notifications and other features remain on the primary database.
An empty `SURVEILLANCE_DB_HOST` preserves the previous single-database behavior.
A configured host requires its own complete connection settings; failure stops startup
and readiness fails if the configured surveillance database becomes unavailable.

## Local container

Use a private environment file based on `infra/env/surveillance.env.example`.
The current workspace configuration is `.local-sso/surveillance.env`, mode 0600;
credentials are not tracked. Start only the new database:

```sh
docker compose --env-file .local-sso/surveillance.env -f infra/compose/docker-compose.surveillance.yml up -d --wait
```

The container is `sso-surveillance-db`, the database/user are `surveillance`, and
host access is restricted to `127.0.0.1:55434`. Its named volume persists data.
Do not use `down -v` to stop it; that removes its data volume.

Load the private variables into the process environment without printing them, then,
from `backend/`, initialize the schema with `go run ./cmd/surveillance-db`.
The backend also applies these migrations automatically when its dedicated connection
is configured. Runtime images include the dedicated migration directory.

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
supported with `SURVEILLANCE_DB_PASSWORD_FILE` by the backend entrypoint.

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
