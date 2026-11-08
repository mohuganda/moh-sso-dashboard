DROP INDEX IF EXISTS idx_documents_health_context_id;

ALTER TABLE documents
    DROP COLUMN IF EXISTS health_context_id;
