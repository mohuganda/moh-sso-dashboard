import { OverflowMenu, OverflowMenuItem } from "@carbon/react";
import type { MouseEvent } from "react";

import type { User } from "../types";

type Props = {
  user: User;
  isCurrentUser?: boolean;
  onEdit: () => void;
  onManageRoles: () => void;
  onToggleStatus: () => void;
  onResetPassword: () => void;
};

export function UserActionsMenu({
  user,
  isCurrentUser = false,
  onEdit,
  onManageRoles,
  onToggleStatus,
  onResetPassword,
}: Props) {
  const stopRowSelection = (event: MouseEvent) => {
    event.stopPropagation();
  };

  return (
    <OverflowMenu size="sm" flipped onClick={stopRowSelection}>
      <OverflowMenuItem itemText="Edit user" hasDivider onClick={onEdit} />

      <OverflowMenuItem itemText="Manage roles" hasDivider onClick={onManageRoles} />

      <OverflowMenuItem
        itemText={
          isCurrentUser
            ? "Cannot disable your own account"
            : user.isActive
              ? "Disable user"
              : "Enable user"
        }
        disabled={isCurrentUser}
        isDelete={user.isActive}
        onClick={onToggleStatus}
      />

      <OverflowMenuItem itemText="Reset password" hasDivider onClick={onResetPassword} />
    </OverflowMenu>
  );
}
