-- Run only against the isolated surveillance database; selected_batches is supplied by the loader.
-- Keep every raw row, including successful records, for reconciliation and retry.
SET LOCAL work_mem='128MB';
\echo Reading staged source rows
CREATE TEMP TABLE source_rows AS
SELECT r.*, b.dataset_type, b.source_name
FROM surveillance_import_raw_rows r
JOIN surveillance_import_batches b ON b.id=r.batch_id
JOIN selected_batches selected ON selected.id=b.id
WHERE r.status='PENDING';
ANALYZE source_rows;

CREATE TEMP TABLE metrics AS
SELECT id, batch_id, row_number, payload, source_name,
 nullif(nullif(btrim(payload->>'region'),''),'nan') AS region,
 nullif(nullif(btrim(payload->>'district'),''),'nan') AS district,
 nullif(nullif(btrim(payload->>'sub_county'),''),'nan') AS sub_county,
 nullif(nullif(btrim(payload->>'facility'),''),'nan') AS facility,
 btrim(payload->>'disease') AS subject,
 CASE WHEN payload->>'year' ~ '^\d{4}$' THEN (payload->>'year')::int END AS year,
 CASE WHEN payload->>'weeks' ~ '^\d{1,2}$' THEN (payload->>'weeks')::int END AS week,
 CASE WHEN payload->>'value' ~ '^\d+(\.\d+)?$' THEN (payload->>'value')::numeric END AS value
FROM source_rows WHERE dataset_type='facility_metrics';
ANALYZE metrics;

UPDATE surveillance_import_raw_rows r SET status='FAILED',error_message='Missing location/subject, invalid year/week, or missing/invalid numeric value'
FROM metrics m WHERE r.id=m.id AND (m.region IS NULL OR m.district IS NULL OR m.facility IS NULL OR coalesce(m.subject,'')='' OR m.year IS NULL OR m.year NOT BETWEEN 1900 AND 2200 OR m.week IS NULL OR m.week NOT BETWEEN 1 AND 53 OR m.value IS NULL OR m.value>999999999999.99);
DELETE FROM metrics m USING surveillance_import_raw_rows r WHERE r.id=m.id AND r.status='FAILED';

\echo Resolving locations and subjects
-- Shared reference data is primary-owned and replicated. A standalone bulk loader
-- must not create an Observatory-only registry that later synchronization overwrites.
UPDATE surveillance_import_raw_rows r SET status='FAILED',
 error_message='Unknown geography: register the location in primary and synchronize Observatory before retrying'
FROM metrics m WHERE r.id=m.id AND NOT EXISTS (
 SELECT 1 FROM regions reg JOIN districts dist ON dist.region_id=reg.id
 JOIN facilities f ON f.district_id=dist.id
 WHERE reg.name=m.region AND dist.name=m.district AND f.name=m.facility
 AND (m.sub_county IS NULL OR EXISTS (
  SELECT 1 FROM sub_counties sub WHERE sub.district_id=dist.id AND sub.name=m.sub_county
 ))
);
DELETE FROM metrics m USING surveillance_import_raw_rows r WHERE r.id=m.id AND r.status='FAILED';

-- Exact seeded measurement names are indicators; '<disease> Cases' maps to the disease.
INSERT INTO diseases(name)
SELECT DISTINCT regexp_replace(subject,' Cases$','') FROM metrics m
WHERE NOT EXISTS(SELECT 1 FROM indicators i WHERE i.name=m.subject)
ON CONFLICT(name) DO NOTHING;

CREATE TEMP TABLE flag_rows AS
SELECT s.id,s.batch_id,s.payload,s.source_name,s.dataset_type,f.key AS color, names.name AS subject
FROM source_rows s CROSS JOIN LATERAL jsonb_each(coalesce(s.payload->'_flags','{}')) f
CROSS JOIN LATERAL jsonb_array_elements_text(f.value) names(name)
WHERE s.dataset_type IN ('national_status','region_status','district_status');
ANALYZE flag_rows;
-- Preserve source disease labels in raw rows; this known naming alias is an indicator.
UPDATE flag_rows SET subject='Maternal Deaths' WHERE subject='Maternal Death';
INSERT INTO diseases(name)
SELECT DISTINCT subject FROM flag_rows f WHERE NOT EXISTS(SELECT 1 FROM indicators i WHERE i.name=f.subject)
ON CONFLICT(name) DO NOTHING;
INSERT INTO diseases(name)
SELECT DISTINCT btrim(payload->>'disease') FROM source_rows WHERE dataset_type='alerts' AND coalesce(btrim(payload->>'disease'),'')<>'' ON CONFLICT(name) DO NOTHING;

-- ISO week dates match the existing epi_weeks migration.
INSERT INTO epi_weeks(epi_year,epi_week,week_start_date,week_end_date)
SELECT DISTINCT y,w,(make_date(y,1,4)-((extract(isodow FROM make_date(y,1,4)))::int-1)+(w-1)*7),
(make_date(y,1,4)-((extract(isodow FROM make_date(y,1,4)))::int-1)+(w-1)*7+6)
FROM (
 SELECT year y,week w FROM metrics
 UNION SELECT CASE WHEN payload->>'year' ~ '^\d{4}$' THEN (payload->>'year')::int END,CASE WHEN payload->>'weeks' ~ '^\d{1,2}$' THEN (payload->>'weeks')::int END FROM source_rows WHERE dataset_type LIKE '%_status'
 UNION SELECT CASE WHEN payload->>'created' ~ '^\d{4}-\d{2}-\d{2}$' THEN left(payload->>'created',4)::int END,CASE WHEN payload->>'weeks' ~ '^\d{1,2}$' THEN (payload->>'weeks')::int END FROM source_rows WHERE dataset_type='alerts'
) dates WHERE y BETWEEN 1900 AND 2200 AND w BETWEEN 1 AND 53 ON CONFLICT(epi_year,epi_week) DO NOTHING;

\echo Reconciling facility observations
CREATE TEMP TABLE resolved_metrics AS
SELECT m.*,f.id facility_id,d.id disease_id,i.id indicator_id,w.id epi_week_id,
row_number() OVER(subject_week ORDER BY m.row_number) AS duplicate_number,
min(m.value) OVER(subject_week) AS minimum_value,max(m.value) OVER(subject_week) AS maximum_value
FROM metrics m JOIN districts dist ON dist.name=m.district JOIN facilities f ON f.name=m.facility AND f.district_id=dist.id
LEFT JOIN indicators i ON i.name=m.subject
LEFT JOIN diseases d ON i.id IS NULL AND d.name=regexp_replace(m.subject,' Cases$','')
JOIN epi_weeks w ON w.epi_year=m.year AND w.epi_week=m.week
WINDOW subject_week AS (PARTITION BY f.id,coalesce(d.id,i.id),w.id);
ANALYZE resolved_metrics;
UPDATE surveillance_import_raw_rows r SET status='FAILED',error_message='Conflicting facility/subject/week values; all members held for review'
FROM resolved_metrics m WHERE r.id=m.id AND m.minimum_value<>m.maximum_value;
INSERT INTO facility_weekly_metrics(source_record_id,facility_id,disease_id,epi_week_id,metric_value,source_name)
SELECT nullif(payload->>'record_id',''),facility_id,disease_id,epi_week_id,value,source_name FROM resolved_metrics WHERE duplicate_number=1 AND minimum_value=maximum_value AND disease_id IS NOT NULL
ON CONFLICT(facility_id,disease_id,epi_week_id) WHERE disease_id IS NOT NULL DO UPDATE SET metric_value=EXCLUDED.metric_value,source_record_id=EXCLUDED.source_record_id,source_name=EXCLUDED.source_name;
INSERT INTO facility_weekly_metrics(source_record_id,facility_id,indicator_id,epi_week_id,metric_value,source_name)
SELECT nullif(payload->>'record_id',''),facility_id,indicator_id,epi_week_id,value,source_name FROM resolved_metrics WHERE duplicate_number=1 AND minimum_value=maximum_value AND indicator_id IS NOT NULL
ON CONFLICT(facility_id,indicator_id,epi_week_id) WHERE indicator_id IS NOT NULL DO UPDATE SET metric_value=EXCLUDED.metric_value,source_record_id=EXCLUDED.source_record_id,source_name=EXCLUDED.source_name;
UPDATE surveillance_import_raw_rows r SET status='PROCESSED',error_message=NULL FROM resolved_metrics m WHERE r.id=m.id AND m.minimum_value=m.maximum_value;

\echo Processing weekly status flags
CREATE TEMP TABLE resolved_flags AS
SELECT f.*,d.id disease_id,i.id indicator_id,w.id epi_week_id,coalesce(reg.id,dist.region_id) region_id,dist.id district_id
FROM flag_rows f LEFT JOIN indicators i ON i.name=f.subject LEFT JOIN diseases d ON i.id IS NULL AND d.name=f.subject
LEFT JOIN regions reg ON lower(reg.name)=lower(btrim(f.payload->>'region'))
LEFT JOIN districts dist ON lower(dist.name)=lower(btrim(f.payload->>'district'))
LEFT JOIN epi_weeks w ON w.epi_year=(f.payload->>'year')::int AND w.epi_week=(f.payload->>'weeks')::int;
ANALYZE resolved_flags;
UPDATE surveillance_import_raw_rows r SET status='FAILED',error_message='Unknown location/week in flag row'
FROM resolved_flags f WHERE r.id=f.id AND (f.epi_week_id IS NULL OR (f.dataset_type='region_status' AND f.region_id IS NULL) OR (f.dataset_type='district_status' AND (f.district_id IS NULL OR f.region_id IS NULL)));
-- Flag lists can contain the same subject more than once; conflicting colours require review.
UPDATE surveillance_import_raw_rows r SET status='FAILED',error_message='Conflicting colors for the same subject in one row'
FROM (SELECT id,subject FROM resolved_flags GROUP BY id,subject HAVING count(DISTINCT color)>1) bad WHERE r.id=bad.id;
INSERT INTO weekly_status(region_id,district_id,disease_id,indicator_id,epi_week_id,status,source_name)
SELECT DISTINCT f.region_id,f.district_id,f.disease_id,f.indicator_id,f.epi_week_id,f.color::risk_level,f.source_name
FROM resolved_flags f JOIN surveillance_import_raw_rows r ON r.id=f.id WHERE r.status='PENDING'
ON CONFLICT DO NOTHING;
UPDATE surveillance_import_raw_rows r SET status='PROCESSED',error_message=NULL
FROM source_rows s WHERE r.id=s.id AND s.dataset_type IN ('national_status','region_status','district_status') AND r.status='PENDING';

-- eIDSR source weeks are parsed from staged strings; preserve narratives in the database only.
\echo Processing eIDSR alerts
CREATE TEMP TABLE resolved_alerts AS
SELECT s.*,d.id disease_id,dist.id district_id,w.id epi_week_id,
CASE WHEN payload->>'created' ~ '^\d{4}-\d{2}-\d{2}$' THEN (payload->>'created')::date END AS created_on
FROM source_rows s LEFT JOIN diseases d ON d.name=btrim(payload->>'disease')
LEFT JOIN districts dist ON lower(dist.name)=lower(btrim(payload->>'district'))
LEFT JOIN epi_weeks w ON w.epi_year=CASE WHEN payload->>'created' ~ '^\d{4}-\d{2}-\d{2}$' THEN left(payload->>'created',4)::int END AND w.epi_week=CASE WHEN payload->>'weeks' ~ '^\d{1,2}$' THEN (payload->>'weeks')::int END
WHERE s.dataset_type='alerts';
ANALYZE resolved_alerts;
UPDATE surveillance_import_raw_rows r SET status='FAILED',error_message='Alert missing disease, district, date/week or narrative'
FROM resolved_alerts a WHERE r.id=a.id AND (a.disease_id IS NULL OR a.district_id IS NULL OR a.epi_week_id IS NULL OR a.created_on IS NULL OR coalesce(btrim(a.payload->>'narrative'),'')='');
INSERT INTO alerts(external_id,disease_id,district_id,epi_week_id,occurred_on,created_on,narrative,submitted_by,source_name)
SELECT 'folder:'||md5(concat_ws('|',a.payload->>'created',a.payload->>'district',a.payload->>'disease',a.payload->>'weeks',a.payload->>'narrative')),disease_id,district_id,epi_week_id,created_on,created_on,a.payload->>'narrative',a.payload->>'submitted_by',source_name
FROM resolved_alerts a JOIN surveillance_import_raw_rows r ON r.id=a.id WHERE r.status='PENDING' ON CONFLICT(external_id) DO NOTHING;
UPDATE surveillance_import_raw_rows r SET status='PROCESSED',error_message=NULL FROM resolved_alerts a WHERE r.id=a.id AND r.status='PENDING';

\echo Reconciling batch counts
UPDATE surveillance_import_batches b SET status='COMPLETED', completed_at=now(),
 success_rows=(SELECT count(*) FROM surveillance_import_raw_rows r WHERE r.batch_id=b.id AND r.status='PROCESSED'),
 failed_rows=(SELECT count(*) FROM surveillance_import_raw_rows r WHERE r.batch_id=b.id AND r.status='FAILED')
FROM selected_batches selected WHERE b.id=selected.id;
