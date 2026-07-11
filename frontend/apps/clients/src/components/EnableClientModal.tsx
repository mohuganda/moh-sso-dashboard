import { Button, InlineLoading } from "@carbon/react";
import React, { useState } from "react";

import "./client-components.scss";

type EnableClientModalProps = {
  clientName: string;
  enabled: boolean;
  onConfirm: () => Promise<void> | void;
  onCancel: () => void;
};

export const EnableClientModal: React.FC<EnableClientModalProps> = ({
  clientName,
  enabled,
  onConfirm,
  onCancel,
}) => {
  const [loading, setLoading] = useState(false);

  async function handleConfirm() {
    try {
      setLoading(true);
      await onConfirm();
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="client-confirm-modal">
      <p>
        Are you sure you want to <strong>{enabled ? "disable" : "enable"}</strong> the client
        <strong> {clientName}</strong>?
      </p>

      <p className="client-confirm-modal__description">
        {enabled
          ? "This client will no longer be able to authenticate or access applications."
          : "This client will regain access to authenticate and use assigned applications."}
      </p>

      {loading && (
        <InlineLoading description={enabled ? "Disabling client…" : "Enabling client…"} />
      )}

      <div className="client-confirm-modal__actions">
        <Button kind="secondary" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>

        <Button kind={enabled ? "danger" : "primary"} onClick={handleConfirm} disabled={loading}>
          {enabled ? "Disable client" : "Enable client"}
        </Button>
      </div>
    </div>
  );
};
