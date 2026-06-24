ALTER TABLE ihp_systems
    DROP CONSTRAINT IF EXISTS ihp_systems_sidenav_type_check,
    DROP CONSTRAINT IF EXISTS ihp_systems_launch_mode_check,
    DROP CONSTRAINT IF EXISTS ihp_systems_system_type_check,
    DROP COLUMN IF EXISTS launch_mode,
    DROP COLUMN IF EXISTS display_in_sidenav,
    DROP COLUMN IF EXISTS display_in_launcher,
    DROP COLUMN IF EXISTS system_type;
