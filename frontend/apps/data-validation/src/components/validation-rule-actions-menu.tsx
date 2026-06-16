import { OverflowMenu, OverflowMenuItem } from "@carbon/react";
import type { MouseEvent } from "react";

import type { ValidationRule } from "../types";

type ValidationRuleActionsMenuProps = {
  rule: ValidationRule;
  onView: () => void;
  onEdit: () => void;
  onDelete: () => void;
};

export function ValidationRuleActionsMenu({
  rule,
  onView,
  onEdit,
  onDelete,
}: ValidationRuleActionsMenuProps) {
  const stopRowSelection = (event: MouseEvent) => {
    event.stopPropagation();
  };

  return (
    <OverflowMenu size="sm" flipped onClick={stopRowSelection}>
      <OverflowMenuItem itemText="View rule" hasDivider onClick={onView} />
      <OverflowMenuItem itemText="Edit rule" hasDivider onClick={onEdit} />
      <OverflowMenuItem
        itemText={rule.type === "builtin" ? "Remove built-in rule" : "Delete rule"}
        isDelete
        onClick={onDelete}
      />
    </OverflowMenu>
  );
}
