import { Button } from "@carbon/react";

import type { AuditLog } from "../types";
import "./audit-components.scss";

type AuditLogBulkActionsProps = {
  logs: AuditLog[];
  onExportCsv: () => void;
  onExportJson: () => void;
};

export function AuditLogBulkActions({ logs, onExportCsv, onExportJson }: AuditLogBulkActionsProps) {
  if (logs.length === 0) return null;

  return (
    <div className="audit-log-bulk-actions">
      <strong>{logs.length} selected</strong>

      <div className="audit-log-bulk-actions__buttons">
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
