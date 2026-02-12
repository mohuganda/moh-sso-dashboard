import { Button, InlineLoading } from "@carbon/react";
import { useState } from "react";

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
    <div style={{ display: "grid", gap: 16 }}>
      <p>
        This will initiate a password reset for <strong>{username}</strong>.
      </p>

      <p style={{ opacity: 0.75 }}>
        {email
          ? `A password reset email will be sent to ${email}.`
          : "The user will be required to set a new password on next login."}
      </p>

      {loading && <InlineLoading description="Sending password reset…" />}

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

        <Button kind="danger" onClick={handleConfirm} disabled={loading}>
          Reset password
        </Button>
      </div>
    </div>
  );
}
