UPDATE document_template_sheets
SET
    header_row    = 2,
    start_row     = 3,
    configuration = '{}'::jsonb,
    updated_at    = NOW()
WHERE code = 'lab'
  AND template_id = (SELECT id FROM document_templates WHERE code = 'GHSC_PSM');

UPDATE document_template_sheets
SET
    header_row    = 8,
    start_row     = 9,
    configuration = '{}'::jsonb,
    updated_at    = NOW()
WHERE code IN ('pharma', 'malaria')
  AND template_id = (SELECT id FROM document_templates WHERE code = 'GHSC_PSM');
