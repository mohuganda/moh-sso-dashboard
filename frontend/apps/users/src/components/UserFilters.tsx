import { Dropdown, Button, Stack } from "@carbon/react";

import "./user-components.scss";

interface UserFiltersProps {
  onReset: () => void;
  status: string;
  roles: string[];
  selectedRole: string;
  neverLoggedIn: boolean;
  onStatusChange: (v: string) => void;
  onRoleChange: (v: string) => void;
  onToggleNeverLoggedIn: () => void;
}

export function UserFilters({
  onReset,
  status,
  roles,
  selectedRole,
  neverLoggedIn,
  onStatusChange,
  onRoleChange,
  onToggleNeverLoggedIn,
}: UserFiltersProps) {
  return (
    <Stack orientation="horizontal" gap={5} className="user-filters">
      {/* Status */}
      <Dropdown
        id="user-status-filter"
        label="Status"
        titleText="Status"
        items={["all", "active", "disabled"]}
        selectedItem={status}
        className="user-filters__field--status"
        onChange={({ selectedItem }) => {
          onStatusChange(selectedItem!);
        }}
      />

      {/* Role */}
      <Dropdown
        id="user-role-filter"
        label="Role"
        titleText="Role"
        items={roles}
        selectedItem={selectedRole}
        className="user-filters__field--role"
        onChange={({ selectedItem }) => {
          onRoleChange(selectedItem!);
        }}
      />

      {/* Never logged in toggle */}
      <Button
        kind={neverLoggedIn ? "primary" : "secondary"}
        size="md"
        aria-pressed={neverLoggedIn}
        onClick={onToggleNeverLoggedIn}
      >
        Never logged in
      </Button>
      <Button kind="ghost" size="md" onClick={onReset}>
        Clear filters
      </Button>
    </Stack>
  );
}
