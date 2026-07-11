import { Button, InlineLoading } from "@carbon/react";
import { useState } from "react";

import "./user-components.scss";

type ResetPasswordModalProps = {
  username: string;
  email?: string | null;
  onConfirm: () => Promise<void> | void;
  onCancel: () => void;
};

export function ResetPasswordModal({
  username,
  email,
  onConfirm,
  onCancel,
}: ResetPasswordModalProps) {
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
        This will initiate a password reset for <strong>{username}</strong>.
      </p>

      <p className="user-confirm-modal__description">
        {email
          ? `A password reset email will be sent to ${email}.`
          : "The user will be required to set a new password on next login."}
      </p>

      {loading && <InlineLoading description="Sending password reset…" />}

      <div className="user-confirm-modal__actions">
        <Button kind="secondary" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>

        <Button kind="danger" onClick={handleConfirm} disabled={loading}>
          Reset password
        </Button>
      </div>
    </div>
  );
}
