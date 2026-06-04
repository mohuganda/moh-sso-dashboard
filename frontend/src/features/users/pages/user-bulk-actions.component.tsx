import { Button, ButtonSet, InlineLoading } from "@carbon/react";

import type { User } from "@/store/types/user.types";

type BulkAction = "enable" | "disable" | "roles" | null;

type Props = {
  users: User[];
  loadingAction: BulkAction;
  onEnable: () => Promise<void>;
  onDisable: () => Promise<void>;
  onAssignRoles: () => Promise<void>;
};

export function UserBulkActions({
  users,
  loadingAction,
  onEnable,
  onDisable,
  onAssignRoles,
}: Props) {
  if (users.length === 0) return null;

  const allEnabled = users.every((u) => u.isActive);
  const allDisabled = users.every((u) => !u.isActive);
  const busy = loadingAction !== null;

  return (
    <ButtonSet style={{ marginBottom: 16 }}>
      <Button kind="secondary" disabled={allEnabled || busy} onClick={onEnable}>
        {loadingAction === "enable" ? (
          <InlineLoading description="Enabling…" />
        ) : (
          `Enable (${users.length})`
        )}
      </Button>

      <Button kind="danger" disabled={allDisabled || busy} onClick={onDisable}>
        {loadingAction === "disable" ? (
          <InlineLoading description="Disabling…" />
        ) : (
          `Disable (${users.length})`
        )}
      </Button>

      <Button kind="primary" disabled={busy} onClick={onAssignRoles}>
        {loadingAction === "roles" ? (
          <InlineLoading description="Opening…" />
        ) : (
          `Assign roles (${users.length})`
        )}
      </Button>
    </ButtonSet>
  );
}
