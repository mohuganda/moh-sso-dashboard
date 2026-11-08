ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS health_context_id UUID
        REFERENCES health_context_nodes(id) ON DELETE RESTRICT;

-- Classify records created before contextual authorization at the national
-- root. New uploads from context-aware users are assigned explicitly.
UPDATE documents
SET health_context_id = (
    SELECT id
    FROM health_context_nodes
    WHERE code = 'UG'
    LIMIT 1
)
WHERE health_context_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_documents_health_context_id
    ON documents (health_context_id)
    WHERE deleted_at IS NULL;
