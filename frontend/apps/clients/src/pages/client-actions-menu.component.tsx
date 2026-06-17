import { OverflowMenu, OverflowMenuItem } from "@carbon/react";
import type { MouseEvent } from "react";

import type { Client } from "@moh-sso/types";

type Props = {
  client: Client;
  onEdit: () => void;
  onToggleStatus: () => void;
};

export function ClientActionsMenu({ client, onEdit, onToggleStatus }: Props) {
  const handleEdit = (event: MouseEvent<HTMLElement>) => {
    event.stopPropagation();
    onEdit();
  };

  const handleToggleStatus = (event: MouseEvent<HTMLElement>) => {
    event.stopPropagation();
    onToggleStatus();
  };

  return (
    <OverflowMenu
      size="sm"
      flipped
      onClick={(event) => {
        event.stopPropagation();
      }}
    >
      <OverflowMenuItem itemText="Edit client" hasDivider onClick={handleEdit} />

      <OverflowMenuItem
        itemText={client.enabled ? "Disable client" : "Enable client"}
        isDelete={client.enabled}
        onClick={handleToggleStatus}
      />
    </OverflowMenu>
  );
}
