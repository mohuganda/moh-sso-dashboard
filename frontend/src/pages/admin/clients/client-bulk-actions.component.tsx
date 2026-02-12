import { Button, ButtonSet, InlineLoading } from "@carbon/react";

import type { Client } from "../../../store/types/client.types";

type Props = {
  clients: Client[];
  loadingAction?: "enable" | "disable" | null;
  onEnable: () => Promise<void>;
  onDisable: () => Promise<void>;
};

export function ClientBulkActions({ clients, loadingAction, onEnable, onDisable }: Props) {
  if (clients.length === 0) return null;

  const allEnabled = clients.every((c) => c.enabled);
  const allDisabled = clients.every((c) => !c.enabled);

  return (
    <ButtonSet style={{ marginBottom: 16 }}>
      <Button kind="secondary" disabled={allEnabled || loadingAction !== null} onClick={onEnable}>
        {loadingAction === "enable" ? (
          <InlineLoading description="Enabling…" />
        ) : (
          `Enable (${clients.length})`
        )}
      </Button>

      <Button kind="danger" disabled={allDisabled || loadingAction !== null} onClick={onDisable}>
        {loadingAction === "disable" ? (
          <InlineLoading description="Disabling…" />
        ) : (
          `Disable (${clients.length})`
        )}
      </Button>
    </ButtonSet>
  );
}
