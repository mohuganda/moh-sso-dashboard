UPDATE document_template_sheets
SET header_row = 1,
    start_row  = 2,
    updated_at = NOW()
WHERE code = 'jms_stock_on_hand'
  AND template_id = (
      SELECT id FROM document_templates WHERE code = 'JMS_STOCK_REPORT'
  );
