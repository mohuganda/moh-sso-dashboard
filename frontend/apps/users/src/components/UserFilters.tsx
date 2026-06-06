import { Dropdown, Button, Stack } from "@carbon/react";

import { useHeaderPanel } from "@/shared/components/header-panel/header-panel.context";
import { ImportUsersPanel } from "./import-users-panel";

interface UserFiltersProps {
  status: string;
  roles: string[];
  selectedRole: string;
  neverLoggedIn: boolean;
  onStatusChange: (v: string) => void;
  onRoleChange: (v: string) => void;
  onToggleNeverLoggedIn: () => void;
}

export function UserFilters({
  status,
  roles,
  selectedRole,
  neverLoggedIn,
  onStatusChange,
  onRoleChange,
  onToggleNeverLoggedIn,
}: UserFiltersProps) {
  const { openPanel } = useHeaderPanel();
  return (
    <Stack
      orientation="horizontal"
      gap={5}
      style={{
        width: "100%",
        justifyContent: "flex-end",
        alignItems: "flex-end",
        flexWrap: "nowrap",
      }}
    >
      {/* Status */}
      <Dropdown
        id="user-status-filter"
        label="Status"
        hideLabel
        titleText="Status"
        items={["all", "active", "disabled"]}
        selectedItem={status}
        style={{ width: 160 }}
        onChange={({ selectedItem }) => {
          onStatusChange(selectedItem!);
        }}
      />

      {/* Role */}
      <Dropdown
        id="user-role-filter"
        label="Role"
        hideLabel
        titleText="Role"
        items={roles}
        selectedItem={selectedRole}
        style={{ width: 200 }}
        onChange={({ selectedItem }) => {
          onRoleChange(selectedItem!);
        }}
      />

      {/* Never logged in toggle */}
      <Button
        kind={neverLoggedIn ? "primary" : "secondary"}
        size="md"
        onClick={onToggleNeverLoggedIn}
      >
        Never logged in
      </Button>

      <Button
        size="md"
        kind="primary"
        onClick={() => {
          openPanel({
            title: "Import users",
            content: <ImportUsersPanel />,
            size: "lg",
          });
        }}
      >
        Import users
      </Button>
    </Stack>
  );
}
