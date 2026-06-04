import { Button, InlineLoading } from "@carbon/react";
import React, { useState } from "react";

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
    <div style={{ display: "grid", gap: 16 }}>
      <p>
        Are you sure you want to <strong>{enabled ? "disable" : "enable"}</strong> the client
        <strong> {clientName}</strong>?
      </p>

      <p style={{ opacity: 0.75 }}>
        {enabled
          ? "This client will no longer be able to authenticate or access applications."
          : "This client will regain access to authenticate and use assigned applications."}
      </p>

      {loading && (
        <InlineLoading description={enabled ? "Disabling client…" : "Enabling client…"} />
      )}

      <div
        style={{
          display: "flex",
          justifyContent: "flex-end",
          gap: 8,
        }}
      >
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
