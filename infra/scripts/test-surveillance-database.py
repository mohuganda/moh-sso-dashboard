#!/usr/bin/env python3
"""Real PostgreSQL import regression; opt in with OBSERVATORY_DOCKER_TEST=1."""
import csv
import os
from pathlib import Path
import subprocess
import tempfile
import time
import unittest
import uuid

ROOT=Path(__file__).resolve().parents[2]

@unittest.skipUnless(os.environ.get('OBSERVATORY_DOCKER_TEST')=='1','requires local Docker')
class ObservatoryDatabaseTest(unittest.TestCase):
 def test_isolated_import_preserves_conflicts_and_week_zero_and_retries(self):
  container='observatory-test-'+uuid.uuid4().hex[:10]
  subprocess.run(['docker','run','-d','--name',container,'--tmpfs','/var/lib/postgresql/data','-e','POSTGRES_USER=observatory','-e','POSTGRES_DB=observatory','-e','POSTGRES_PASSWORD=fixture-only','postgres:15-alpine'],check=True,stdout=subprocess.DEVNULL)
  self.addCleanup(lambda: subprocess.run(['docker','rm','-f',container],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL))
  def query(sql):
   return subprocess.check_output(['docker','exec','-i',container,'psql','-X','-v','ON_ERROR_STOP=1','-U','observatory','-d','observatory','-Atq'],input=sql,text=True).strip()
  for _ in range(60):
   result=subprocess.run(['docker','exec',container,'pg_isready','-h','127.0.0.1','-U','observatory','-d','observatory'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
   if result.returncode==0:break
   time.sleep(0.5)
  # pg_isready can briefly see the bootstrap server before its final restart.
  time.sleep(1)
  for migration in sorted((ROOT/'backend/internal/db/surveillance_migrations').glob('*.up.sql')):
   query(migration.read_text())
  with tempfile.TemporaryDirectory() as directory:
   directory=Path(directory)
   def write(name,headers,rows):
    with (directory/name).open('w',newline='') as file:
     writer=csv.writer(file);writer.writerow(headers);writer.writerows(rows)
   write('data.csv',['record_id','facility','region','district','sub_county','disease','value','year','weeks'],[
    ['a','Facility A','Region A','District A','Area A','Measles Cases',2,2026,1],
    ['b','Facility A','Region A','District A','Area A','Measles Cases',2,2026,1],
    ['c','Facility A','Region A','District A','Area A','Measles Cases',3,2026,2],
    ['d','Facility A','Region A','District A','Area A','Measles Cases',4,2026,2],
    ['e','Facility A','Region A','District A','Area A','Cases Tested with RDT',5,2026,1],
    ['f','Facility A','Region A','District A','Area A','Measles Cases','',2026,3]])
   write('eidsr.csv',['created','narrative','district','weeks','submitted_by','disease'],[['2026-01-02','Synthetic fixture','District A',1,'fixture','Measles']])
   write('flags.csv',['weeks','maroon','red','yellow','green'],[[0,"['Measles']",'[]','[]','[]'],[1,"['Measles']",'[]','[]','[]']])
   write('region_flags.csv',['weeks','region','maroon','red','yellow','green'],[[1,'Region A','[]',"['Measles']",'[]','[]']])
   write('district_flags.csv',['weeks','district','maroon','red','yellow','green'],[[1,'District A','[]','[]',"['Measles']",'[]']])
   command=['python3',str(ROOT/'infra/scripts/load-surveillance.py'),'--container',container,'--db-user','observatory','--directory',str(directory),'--flag-year','2026']
   subprocess.run(command,check=True,stdout=subprocess.DEVNULL)
   self.assertEqual(query('SELECT count(*) FROM facility_weekly_metrics;'),'2')
   self.assertEqual(query('SELECT count(*) FROM weekly_status;'),'3')
   self.assertEqual(query('SELECT count(*) FROM alerts;'),'1')

   # Weekly map totals must retain a facility's district/region when no
   # sub-county is recorded. Execute the actual repository SQL for both paths.
   query("UPDATE facilities SET sub_county_id=NULL WHERE name='Facility A';")
   week=query('SELECT id FROM epi_weeks WHERE epi_year=2026 AND epi_week=1;')
   facility=query("SELECT id FROM facilities WHERE name='Facility A';")
   source=(ROOT/'backend/internal/db/query/facility_weekly_metrics.sql').read_text()
   for name in ('ListFacilityWeeklyDiseaseMetricsByWeek','ListFacilityWeeklyDiseaseMetricsByWeekInHealthContext'):
    statement=source.split('-- name: '+name+' :many\n',1)[1].split('-- name:',1)[0].strip().rstrip(';')
    statement=statement.replace("sqlc.arg('epi_week_id')", "'"+week+"'::uuid").replace('$1', "'"+week+"'::uuid")
    statement=statement.replace("sqlc.arg('allowed_facility_ids')", "ARRAY['"+facility+"']")
    self.assertEqual(query("SELECT count(*) FROM ("+statement+") result WHERE district_name='District A' AND region_name='Region A' AND sub_county_id IS NULL;"),'1')
   self.assertEqual(query("SELECT count(*) FROM surveillance_import_raw_rows WHERE status='FAILED';"),'4')
   self.assertEqual(query("SELECT success_rows||'|'||failed_rows FROM surveillance_import_batches WHERE file_name='data.csv';"),'3|3')
   self.assertEqual(query("SELECT payload->>'weeks' FROM surveillance_import_raw_rows WHERE payload->>'weeks'='0';"),'0')
   self.assertEqual(query("SELECT count(*) FROM information_schema.table_constraints WHERE table_name='surveillance_import_batches' AND constraint_type='FOREIGN KEY';"),'0')
   subprocess.run(command,check=True,stdout=subprocess.DEVNULL)
   self.assertEqual(query('SELECT count(*) FROM surveillance_import_batches;'),'5')
   subprocess.run(command+['--reprocess'],check=True,stdout=subprocess.DEVNULL)
   self.assertEqual(query('SELECT count(*) FROM facility_weekly_metrics;'),'2')
   self.assertEqual(query('SELECT count(*) FROM weekly_status;'),'3')
   self.assertEqual(query('SELECT count(*) FROM alerts;'),'1')

if __name__=='__main__':unittest.main()
