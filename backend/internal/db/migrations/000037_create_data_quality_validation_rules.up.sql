CREATE TABLE IF NOT EXISTS data_quality_validation_rules (
    id BIGSERIAL PRIMARY KEY,
    table_id TEXT,
    program TEXT,
    category TEXT,
    code TEXT NOT NULL,
    severity TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    column_name TEXT NOT NULL,
    operator TEXT NOT NULL,
    value TEXT,
    value_column TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by TEXT,
    updated_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT data_quality_validation_rules_severity_check
        CHECK (severity IN ('error', 'warning', 'info')),
    CONSTRAINT data_quality_validation_rules_operator_check
        CHECK (operator IN ('contains', 'eq', 'gt', 'gte', 'isnull', 'lt', 'lte', 'ne', 'notnull')),
    CONSTRAINT data_quality_validation_rules_compare_target_check
        CHECK (
            operator IN ('isnull', 'notnull')
            OR (
                (value IS NOT NULL AND value_column IS NULL)
                OR (value IS NULL AND value_column IS NOT NULL)
            )
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_data_quality_validation_rules_identity
    ON data_quality_validation_rules (
        COALESCE(table_id, ''),
        COALESCE(program, ''),
        COALESCE(category, ''),
        code
    )
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_data_quality_validation_rules_lookup
    ON data_quality_validation_rules (table_id, program, category, code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_data_quality_validation_rules_updated_at
    ON data_quality_validation_rules (updated_at DESC, id DESC)
    WHERE deleted_at IS NULL;
