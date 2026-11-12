package health_context

import (
	"context"
	"fmt"
	"time"
)

// Observatory owns geography; the primary database owns its context aliases.
// Synchronize through the two connections instead of cross-database triggers.
func (r *postgresRepository) ensureGeographySync(ctx context.Context) error {
	if r.geographyDB == nil || (r.geographyAvailable != nil && !r.geographyAvailable()) {
		return nil
	}
	r.geographyMu.Lock()
	defer r.geographyMu.Unlock()
	if time.Since(r.geographySyncedAt) < time.Minute {
		return nil
	}
	levels := []struct{ namespace, prefix, contextType, query string }{
		{"surveillance-region", "SURV-REGION-", "REGION", `SELECT id, name, NULL::text AS parent_namespace, NULL::uuid AS parent_id FROM regions`},
		{"surveillance-district", "SURV-DISTRICT-", "DISTRICT", `SELECT id, name, 'surveillance-region'::text AS parent_namespace, region_id AS parent_id FROM districts`},
		{"surveillance-sub-county", "SURV-SUB-COUNTY-", "SUB_COUNTY", `SELECT id, name, 'surveillance-district'::text AS parent_namespace, district_id AS parent_id FROM sub_counties`},
		{"surveillance-facility", "SURV-FACILITY-", "FACILITY", `SELECT id, name,
		 CASE WHEN sub_county_id IS NOT NULL THEN 'surveillance-sub-county' WHEN district_id IS NOT NULL THEN 'surveillance-district' WHEN region_id IS NOT NULL THEN 'surveillance-region' END AS parent_namespace,
		 COALESCE(sub_county_id, district_id, region_id) AS parent_id FROM facilities`},
	}
	for _, level := range levels {
		var records []byte
		query := `SELECT COALESCE(jsonb_agg(to_jsonb(geography)), '[]'::jsonb) FROM (` + level.query + `) geography`
		if err := r.geographyDB.QueryRowContext(ctx, query).Scan(&records); err != nil {
			return fmt.Errorf("read observatory geography: %w", err)
		}
		_, err := r.db.ExecContext(ctx, `SELECT sync_surveillance_health_context(
		 $1, record.id, $2 || record.id::text, record.name, $3, record.parent_namespace, record.parent_id
		) FROM jsonb_to_recordset($4::jsonb) AS record(id uuid, name text, parent_namespace text, parent_id uuid)`,
			level.namespace, level.prefix, level.contextType, string(records))
		if err != nil {
			return fmt.Errorf("synchronize primary geography aliases: %w", err)
		}
	}
	r.geographySyncedAt = time.Now()
	return nil
}
