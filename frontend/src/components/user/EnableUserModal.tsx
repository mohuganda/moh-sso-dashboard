import { Button, InlineLoading } from "@carbon/react";
import { useState } from "react";

type EnableUserModalProps = {
  username: string;
  enabled: boolean;
  onConfirm: () => Promise<void> | void;
  onCancel: () => void;
};

export function EnableUserModal({ username, enabled, onConfirm, onCancel }: EnableUserModalProps) {
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
        Are you sure you want to <strong>{enabled ? "disable" : "enable"}</strong> the user{" "}
        <strong>{username}</strong>?
      </p>

      <p style={{ opacity: 0.75 }}>
        {enabled
          ? "The user will no longer be able to log in."
          : "The user will regain access to the platform."}
      </p>

      {loading && <InlineLoading description={enabled ? "Disabling user…" : "Enabling user…"} />}

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
          {enabled ? "Disable user" : "Enable user"}
        </Button>
      </div>
    </div>
  );
}
