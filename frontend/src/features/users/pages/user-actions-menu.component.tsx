import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { User } from "@/store/types/user.types";

type Props = {
  user: User;
  onEdit: () => void;
  onManageRoles: () => void;
  onToggleStatus: () => void;
  onResetPassword: () => void;
};

export function UserActionsMenu({
  user,
  onEdit,
  onManageRoles,
  onToggleStatus,
  onResetPassword,
}: Props) {
  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="Edit user" hasDivider onClick={onEdit} />

      <OverflowMenuItem itemText="Manage roles" hasDivider onClick={onManageRoles} />

      <OverflowMenuItem
        itemText={user.isActive ? "Disable user" : "Enable user"}
        isDelete={user.isActive}
        onClick={onToggleStatus}
      />

      <OverflowMenuItem itemText="Reset password" hasDivider onClick={onResetPassword} />
    </OverflowMenu>
  );
}
