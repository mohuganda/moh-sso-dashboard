import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { Client } from "@moh-sso/types";

type Props = {
  client: Client;
  onEdit: () => void;
  onToggleStatus: () => void;
};

export function ClientActionsMenu({ client, onEdit, onToggleStatus }: Props) {
  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="Edit client" hasDivider onClick={onEdit} />

      <OverflowMenuItem
        itemText={client.enabled ? "Disable client" : "Enable client"}
        isDelete={client.enabled}
        onClick={onToggleStatus}
      />
    </OverflowMenu>
  );
}
