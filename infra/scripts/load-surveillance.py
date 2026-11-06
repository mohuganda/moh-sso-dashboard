#!/usr/bin/env python3
"""Stage and process the repository CSVs in the isolated local surveillance DB."""
import argparse
import ast
import csv
import hashlib
import json
from pathlib import Path
import subprocess
import uuid

ROOT = Path(__file__).resolve().parents[2]
DATASETS = {
    'data.csv': 'facility_metrics', 'eidsr.csv': 'alerts',
    'flags.csv': 'national_status', 'region_flags.csv': 'region_status',
    'district_flags.csv': 'district_status',
}


def prepare_row(row, dataset, year):
    payload = {key: value.strip() for key, value in row.items()}
    error = None
    if dataset.endswith('_status'):
        payload['year'] = str(year)
        try:
            flags = {}
            for color in ('maroon', 'red', 'yellow', 'green'):
                subjects = ast.literal_eval(payload[color])
                if not isinstance(subjects, list) or any(not isinstance(s, str) for s in subjects):
                    raise ValueError('flag values must be lists of names')
                flags[color.upper()] = [s.strip() for s in subjects if s.strip()]
            payload['_flags'] = flags
            if int(payload['weeks']) == 0:
                error = 'Source week 0 retained for separate review; not mapped to week 1'
            elif not 1 <= int(payload['weeks']) <= 53:
                error = 'Week must be between 1 and 53'
        except (ValueError, SyntaxError, KeyError, TypeError):
            error = 'Invalid flag list or week'
    return payload, error


def psql(container, sql, capture=False, database="observatory"):
    return subprocess.run(['docker', 'exec', '-i', container, 'psql', '-X', '-v',
                           'ON_ERROR_STOP=1', '-U', 'surveillance', '-d', database, '-Atq'],
                          input=sql, text=True, check=True,
                          stdout=subprocess.PIPE if capture else None)


def stage(container, path, dataset, year, database="observatory"):
    with path.open('rb') as source_file:
        digest = hashlib.file_digest(source_file, 'sha256').hexdigest()
    source = f'folder:{path.name}:{digest}:year={year}'
    existing = psql(container, "SELECT id FROM surveillance_import_batches WHERE source_name='" + source + "';", True, database=database).stdout.strip()
    if existing:
        print(f'{path.name}: already staged; reusing batch {existing}')
        return existing
    batch = str(uuid.uuid4())
    # Spool to disk instead of retaining 787k JSON documents in memory. A single
    # COPY transaction avoids individual database round trips for every row.
    import tempfile
    with tempfile.TemporaryFile(mode='w+', encoding='utf-8', newline='') as spool:
        spool.write("BEGIN;\nINSERT INTO surveillance_import_batches(id,source_name,file_name,dataset_type,status,notes) VALUES "
                    f"('{batch}','{source}','{path.name}','{dataset}','PENDING','Flag year {year}; source week 0 quarantined');\n")
        spool.write('COPY surveillance_import_raw_rows(batch_id,row_number,payload,status,error_message) FROM STDIN WITH (FORMAT csv);\n')
        writer = csv.writer(spool, lineterminator='\n')
        count = failed = 0
        with path.open(newline='', encoding='utf-8-sig') as file:
            for count, row in enumerate(csv.DictReader(file), 1):
                payload, error = prepare_row(row, dataset, year)
                failed += error is not None
                writer.writerow([batch, count, json.dumps(payload), 'FAILED' if error else 'PENDING', error])
        spool.write('\\.\n')
        spool.write(f"UPDATE surveillance_import_batches SET total_rows={count}, failed_rows={failed} WHERE id='{batch}';\nCOMMIT;\n")
        spool.seek(0)
        subprocess.run(['docker', 'exec', '-i', container, 'psql', '-X', '-q', '-v',
                        'ON_ERROR_STOP=1', '-U', 'surveillance', '-d', database],
                       stdin=spool, check=True)
    print(f'{path.name}: staged {count} rows ({failed} held for review), batch {batch}')
    return batch


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--container', default='sso-surveillance-db')
    parser.add_argument('--database', default='observatory', help='Target database; use surveillance for an existing database before renaming')
    parser.add_argument('--directory', type=Path, default=ROOT / 'surveillance')
    parser.add_argument('--flag-year', type=int, required=True)
    parser.add_argument('--stage-only', action='store_true')
    parser.add_argument('--reprocess', action='store_true', help='Rebuild only this loader’s selected source batches; keeps raw rows')
    args = parser.parse_args()
    if not 1900 <= args.flag_year <= 2200:
        parser.error('--flag-year must be an explicit year between 1900 and 2200')
    # Fixed filenames, not arbitrary input names interpolated into SQL.
    batches = [stage(args.container, args.directory / name, dataset, args.flag_year, database=args.database)
               for name, dataset in DATASETS.items()]
    if not args.stage_only:
        ids = ','.join("'" + str(uuid.UUID(batch)) + "'::uuid" for batch in batches)
        sql = (ROOT / 'infra/sql/process-surveillance.sql').read_text()
        if args.reprocess:
            sql = """
\\echo Resetting only selected source batches
DELETE FROM facility_weekly_metrics WHERE source_name IN (SELECT b.source_name FROM surveillance_import_batches b JOIN selected_batches s ON s.id=b.id);
DELETE FROM weekly_status WHERE source_name IN (SELECT b.source_name FROM surveillance_import_batches b JOIN selected_batches s ON s.id=b.id);
DELETE FROM alerts WHERE source_name IN (SELECT b.source_name FROM surveillance_import_batches b JOIN selected_batches s ON s.id=b.id);
UPDATE surveillance_import_raw_rows r SET status='PENDING',error_message=NULL FROM surveillance_import_batches b JOIN selected_batches s ON s.id=b.id
WHERE r.batch_id=b.id AND (b.dataset_type IN ('facility_metrics','alerts') OR (r.payload ? '_flags' AND r.payload->>'weeks' ~ '^[0-9]+$' AND (r.payload->>'weeks')::int BETWEEN 1 AND 53));
""" + sql
        psql(args.container, 'BEGIN;\nCREATE TEMP TABLE selected_batches AS SELECT unnest(ARRAY[' + ids + ']) AS id;\n' + sql + '\nCOMMIT;\n', database=args.database)
    report = psql(args.container, "SELECT file_name,total_rows,success_rows,failed_rows,status FROM surveillance_import_batches WHERE id IN (" + ','.join("'" + str(uuid.UUID(batch)) + "'" for batch in batches) + ") ORDER BY file_name;", True, database=args.database).stdout
    print('file|total|processed|review|status\n' + report)


if __name__ == '__main__':
    main()
