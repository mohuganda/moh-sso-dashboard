-- GHSC-PSM sheets had wrong header_row / start_row values.
--
-- LAB:     title=row2, headers=row3, data starts row4
-- Pharma:  KEY legend rows1-8, title=row9, headers=row10, data starts row11
-- Malaria: same structure as Pharma
--
-- header_anchor lets the processor auto-detect the real header row by scanning
-- for "Commodity Category", so it tolerates minor row shifts between file versions.

UPDATE document_template_sheets
SET
    header_row    = 3,
    start_row     = 4,
    configuration = '{"header_anchor": "Commodity Category"}'::jsonb,
    updated_at    = NOW()
WHERE code = 'lab'
  AND template_id = (SELECT id FROM document_templates WHERE code = 'GHSC_PSM');

UPDATE document_template_sheets
SET
    header_row    = 10,
    start_row     = 11,
    configuration = '{"header_anchor": "Commodity Category"}'::jsonb,
    updated_at    = NOW()
WHERE code IN ('pharma', 'malaria')
  AND template_id = (SELECT id FROM document_templates WHERE code = 'GHSC_PSM');
