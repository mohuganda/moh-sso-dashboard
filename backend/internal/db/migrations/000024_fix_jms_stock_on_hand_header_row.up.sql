-- JMS Stock On Hand sheet has a blank row 1 in the actual export file.
-- Real column headers sit on row 2 and data starts at row 3.
-- Stock Issues is unaffected (its headers are correctly on row 1).

UPDATE document_template_sheets
SET header_row = 2,
    start_row  = 3,
    updated_at = NOW()
WHERE code = 'jms_stock_on_hand'
  AND template_id = (
      SELECT id FROM document_templates WHERE code = 'JMS_STOCK_REPORT'
  );
