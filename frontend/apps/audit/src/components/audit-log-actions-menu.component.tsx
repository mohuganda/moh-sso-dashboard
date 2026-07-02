import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { AuditLog } from "../types";

type AuditLogActionsMenuProps = {
  log: AuditLog;
  onView: () => void;
};

export function AuditLogActionsMenu({ onView }: AuditLogActionsMenuProps) {
  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="View details" onClick={onView} />
    </OverflowMenu>
  );
}
