import { Close } from "@carbon/react/icons";
import { Button, CodeSnippet, Tag } from "@carbon/react";
import React from "react";

import type { AuditLog } from "@moh-sso/types";

export const AuditLogPanel: React.FC<{
  log: AuditLog | null;
  onClose: () => void;
}> = ({ log, onClose }) => {
  if (!log) return null;

  const success = typeof log.success === "boolean" ? log.success : null;

  return (
    <div style={{ padding: 24, display: "grid", gap: 20 }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between" }}>
        <div>
          <h4 style={{ margin: 0 }}>Audit Log Details</h4>
          <p style={{ marginTop: 4, opacity: 0.7 }}>{new Date(log.createdAt).toLocaleString()}</p>
        </div>

        <Button
          kind="ghost"
          size="sm"
          hasIconOnly
          renderIcon={Close}
          iconDescription="Close"
          onClick={onClose}
        />
      </div>

      {/* Summary */}
      <div>
        <div
          style={{
            display: "flex",
            gap: 8,
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <strong>{log.action}</strong>
          {success !== null && (
            <Tag type={success ? "green" : "red"}>{success ? "success" : "failure"}</Tag>
          )}
        </div>
      </div>

      {/* Core fields */}
      <div style={{ display: "grid", gap: 8 }}>
        <Field label="ID" value={log.id} />
        <Field label="Actor" value={log.username ?? "System"} />
        <Field label="User ID" value={log.userId ?? "—"} />
        <Field label="Client" value={log.clientId ?? "—"} />
        <Field label="IP Address" value={log.ip ?? "—"} />
        <Field label="User Agent" value={log.metadata?.user_agent ?? "—"} />
        <Field
          label="Location"
          value={
            (log.metadata?.country ?? "—") + " / " + (log.metadata?.city ?? "—")
          }
        />
      </div>

      {/* Metadata */}
      <div>
        <strong>Metadata</strong>
        <div style={{ marginTop: 8 }}>
          <CodeSnippet type="multi">{JSON.stringify(log.metadata ?? {}, null, 2)}</CodeSnippet>
        </div>
      </div>
    </div>
  );
};

function Field({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <strong>{label}:</strong> {value}
    </div>
  );
}
