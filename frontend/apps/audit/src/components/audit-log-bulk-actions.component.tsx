import { Button } from "@carbon/react";

import type { AuditLog } from "../types";

type AuditLogBulkActionsProps = {
  logs: AuditLog[];
  onExportCsv: () => void;
  onExportJson: () => void;
};

export function AuditLogBulkActions({ logs, onExportCsv, onExportJson }: AuditLogBulkActionsProps) {
  if (logs.length === 0) return null;

  return (
    <div
      style={{
        display: "flex",
        gap: 8,
        alignItems: "center",
        justifyContent: "space-between",
        paddingBlockEnd: 12,
      }}
    >
      <strong>{logs.length} selected</strong>

      <div
        style={{
          display: "flex",
          gap: 8,
          alignItems: "center",
          flexWrap: "wrap",
        }}
      >
        <Button size="sm" kind="secondary" onClick={onExportCsv}>
          Export CSV
        </Button>

        <Button size="sm" kind="ghost" onClick={onExportJson}>
          Export JSON
        </Button>
      </div>
    </div>
  );
}
