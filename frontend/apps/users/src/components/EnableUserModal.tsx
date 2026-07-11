import { Button, InlineLoading } from "@carbon/react";
import { useState } from "react";

import "./user-components.scss";

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
    <div className="user-confirm-modal">
      <p>
        Are you sure you want to <strong>{enabled ? "disable" : "enable"}</strong> the user{" "}
        <strong>{username}</strong>?
      </p>

      <p className="user-confirm-modal__description">
        {enabled
          ? "The user will no longer be able to log in."
          : "The user will regain access to the platform."}
      </p>

      {loading && <InlineLoading description={enabled ? "Disabling user…" : "Enabling user…"} />}

      <div className="user-confirm-modal__actions">
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
