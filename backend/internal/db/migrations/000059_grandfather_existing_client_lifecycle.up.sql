-- Existing/discovered integrations are already operational. Only an explicit new
-- registration opts into the onboarding state machine after this migration.
ALTER TABLE ihp_systems ADD COLUMN lifecycle_managed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE ihp_systems ALTER COLUMN lifecycle_status SET DEFAULT 'active';

INSERT INTO ihp_rbac_audit_events (action, resource_type, resource_id, system_client_id, details)
SELECT 'system.lifecycle.grandfathered', 'system', id::text, client_id,
       jsonb_build_object('from', lifecycle_status, 'to', 'active',
                          'reason', 'existing integration exempted from new-client onboarding')
FROM ihp_systems
WHERE lifecycle_status IN ('registered', 'configured', 'validated', 'approved');

UPDATE ihp_systems
SET lifecycle_status = 'active', enabled = TRUE, updated_at = NOW()
WHERE lifecycle_status IN ('registered', 'configured', 'validated', 'approved');

CREATE OR REPLACE FUNCTION ihp_enforce_application_lifecycle() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND OLD.lifecycle_managed THEN
        -- Ordinary edits cannot opt a newly onboarded integration out of review.
        NEW.lifecycle_managed := TRUE;
        IF OLD.lifecycle_status IN ('validated', 'approved', 'active') AND
           ROW(NEW.launch_url, NEW.owner_team, NEW.owner_name, NEW.owner_email,
               NEW.support_url, NEW.documentation_url, NEW.environment, NEW.system_type,
               NEW.launch_mode, NEW.integration_type, NEW.contract_version, NEW.metadata)
           IS DISTINCT FROM
           ROW(OLD.launch_url, OLD.owner_team, OLD.owner_name, OLD.owner_email,
               OLD.support_url, OLD.documentation_url, OLD.environment, OLD.system_type,
               OLD.launch_mode, OLD.integration_type, OLD.contract_version, OLD.metadata) THEN
            NEW.lifecycle_status := 'configured';
            INSERT INTO ihp_rbac_audit_events (actor_user_id, action, resource_type, resource_id, system_client_id, details)
            VALUES (NULLIF(current_setting('ihp.actor_user_id', true), ''), 'system.lifecycle.invalidated', 'system', NEW.id::text, NEW.client_id,
                jsonb_build_object('from', OLD.lifecycle_status, 'to', 'configured', 'reason', 'application configuration changed'));
        END IF;
    END IF;
    IF NEW.lifecycle_status <> 'active' THEN
        NEW.enabled := FALSE;
    END IF;
    RETURN NEW;
END;
$$;
