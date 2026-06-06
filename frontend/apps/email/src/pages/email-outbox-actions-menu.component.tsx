import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { EmailOutboxItem } from "@/store/types/email.types";

type Props = {
  email: EmailOutboxItem;
  onView: () => void;
  onRetry: () => void;
  onDelete: () => void;
  isMutating?: boolean;
};

export function EmailOutboxActionsMenu({
  email,
  onView,
  onRetry,
  onDelete,
  isMutating = false,
}: Props) {
  const isSent = email.status === "SENT";
  const isProcessing = email.status === "PROCESSING";

  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="View email" hasDivider onClick={onView} />

      <OverflowMenuItem
        itemText="Retry email"
        hasDivider
        disabled={isMutating || isSent || isProcessing}
        onClick={onRetry}
      />

      <OverflowMenuItem
        itemText="Delete email"
        isDelete
        disabled={isMutating || isProcessing}
        onClick={onDelete}
      />
    </OverflowMenu>
  );
}
