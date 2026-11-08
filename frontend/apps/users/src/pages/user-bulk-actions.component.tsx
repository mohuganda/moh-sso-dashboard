import { Button, ButtonSet, InlineLoading } from "@carbon/react";

import type { User } from "../types";
import "../components/user-components.scss";

type BulkAction = "enable" | "disable" | null;

type Props = {
  users: User[];
  currentUserId?: string;
  loadingAction: BulkAction;
  onEnable: () => Promise<void>;
  onDisable: () => Promise<void>;
};

export function UserBulkActions({
  users,
  currentUserId,
  loadingAction,
  onEnable,
  onDisable,
}: Props) {
  if (users.length === 0) return null;

  const allEnabled = users.every((u) => u.isActive);
  const allDisabled = users.every((u) => !u.isActive);
  const busy = loadingAction !== null;
  const includesCurrentUser = users.some((user) => user.id === currentUserId);

  return (
    <>
      <ButtonSet className="user-bulk-actions">
        <Button kind="secondary" disabled={allEnabled || busy} onClick={onEnable}>
          {loadingAction === "enable" ? (
            <InlineLoading description="Enabling…" />
          ) : (
            `Enable (${users.length})`
          )}
        </Button>

        <Button
          kind="danger"
          disabled={allDisabled || busy || includesCurrentUser}
          onClick={onDisable}
        >
          {loadingAction === "disable" ? (
            <InlineLoading description="Disabling…" />
          ) : (
            `Disable (${users.length})`
          )}
        </Button>
      </ButtonSet>
      {includesCurrentUser && (
        <p>Deselect your account to disable other users. You cannot disable your own account.</p>
      )}
    </>
  );
}
