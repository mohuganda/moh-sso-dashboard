#!/usr/bin/env python3
import importlib.util
from pathlib import Path
import unittest

spec=importlib.util.spec_from_file_location('loader',Path(__file__).with_name('load-surveillance.py'))
loader=importlib.util.module_from_spec(spec)
spec.loader.exec_module(loader)

class SurveillanceRowTests(unittest.TestCase):
 def test_week_zero_is_retained_without_merging_with_week_one(self):
  original={'weeks':'0','maroon':"['Measles']",'red':'[]','yellow':'[]','green':'[]'}
  payload,error=loader.prepare_row(original,'national_status',2026)
  self.assertEqual(payload['year'],'2026');self.assertEqual(payload['weeks'],'0')
  self.assertIn('separate review',error)
  original['weeks']='1'
  payload,error=loader.prepare_row(original,'national_status',2026)
  self.assertIsNone(error);self.assertEqual(payload['weeks'],'1')
 def test_invalid_flag_data_is_retained_for_review(self):
  original={'weeks':'1','maroon':'not a list','red':'[]','yellow':'[]','green':'[]'}
  payload,error=loader.prepare_row(original,'region_status',2026)
  self.assertIsNotNone(error);self.assertEqual(payload['maroon'],'not a list')
 def test_flag_names_are_parsed_without_executing_source(self):
  original={'weeks':'52','maroon':"['Measles', 'Malaria']",'red':'[]','yellow':'[]','green':'[]'}
  payload,error=loader.prepare_row(original,'district_status',2026)
  self.assertIsNone(error);self.assertEqual(payload['_flags']['MAROON'],['Measles','Malaria'])
 def test_metric_year_is_not_overwritten_by_flag_year(self):
  payload,error=loader.prepare_row({'year':'2025','weeks':'2','value':'3'},'facility_metrics',2026)
  self.assertIsNone(error);self.assertEqual(payload['year'],'2025')

if __name__=='__main__':unittest.main()
