ALTER TABLE ihp_systems
    ADD COLUMN system_type TEXT NOT NULL DEFAULT 'platform',
    ADD COLUMN display_in_launcher BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN display_in_sidenav BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN launch_mode TEXT NOT NULL DEFAULT 'internal';

UPDATE ihp_systems
SET display_in_sidenav = COALESCE(NULLIF(metadata->>'navigation', ''), NULLIF(metadata->>'sidenav', '')) IS NOT NULL;

UPDATE ihp_systems
SET system_type = 'external',
    launch_mode = 'new_tab',
    display_in_sidenav = FALSE
WHERE launch_url ~* '^https?://'
  AND COALESCE(NULLIF(metadata->>'navigation', ''), NULLIF(metadata->>'sidenav', '')) IS NULL;

UPDATE ihp_systems
SET system_type = 'platform', launch_mode = 'internal'
WHERE launch_url LIKE '/portal%' OR launch_url LIKE '/apps%';

ALTER TABLE ihp_systems
    ADD CONSTRAINT ihp_systems_system_type_check
        CHECK (system_type IN ('platform', 'external')),
    ADD CONSTRAINT ihp_systems_launch_mode_check
        CHECK (launch_mode IN ('internal', 'new_tab', 'same_tab')),
    ADD CONSTRAINT ihp_systems_sidenav_type_check
        CHECK (NOT display_in_sidenav OR system_type = 'platform');
