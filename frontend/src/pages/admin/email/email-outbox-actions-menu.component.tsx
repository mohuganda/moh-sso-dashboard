import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

export type EmailOutboxAction = {
  id: string;
  label: string;
  onClick: () => void;
  disabled?: boolean;
  danger?: boolean;
  hasDivider?: boolean;
};

type EmailOutboxActionsMenuProps = {
  actions: EmailOutboxAction[];
  flipped?: boolean;
  size?: "sm" | "md" | "lg";
};

export function EmailOutboxActionsMenu({
  actions,
  flipped = true,
  size = "sm",
}: EmailOutboxActionsMenuProps) {
  return (
    <OverflowMenu size={size} flipped={flipped}>
      {actions.map((action) => (
        <OverflowMenuItem
          key={action.id}
          itemText={action.label}
          disabled={action.disabled}
          isDelete={action.danger}
          hasDivider={action.hasDivider}
          onClick={action.onClick}
        />
      ))}
    </OverflowMenu>
  );
}
