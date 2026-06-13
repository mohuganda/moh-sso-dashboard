import { Button, TextArea } from "@carbon/react";

import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacImportPreview } from "@moh-sso/types";

type ImportExportPanelProps = {
  seedText: string;
  importPreview?: RbacImportPreview;
  hasExportedSeed: boolean;
  onSeedTextChange: (value: string) => void;
  onPreview: () => void;
  onApply: () => void;
};

export function ImportExportPanel({
  seedText,
  importPreview,
  hasExportedSeed,
  onSeedTextChange,
  onPreview,
  onApply,
}: ImportExportPanelProps) {
  return (
    <section>
      <h3>Import / Export</h3>
      <TextArea
        id="rbac-seed-import"
        labelText="Seed payload"
        rows={8}
        value={seedText}
        onChange={(event) => onSeedTextChange(event.target.value)}
      />
      <div className="rbac-sync-actions">
        <Button size="sm" onClick={onPreview}>
          Preview import
        </Button>
        <PermissionGuard permission={PERMISSIONS.rbacWrite}>
          <Button size="sm" kind="primary" onClick={onApply}>
            Apply import
          </Button>
        </PermissionGuard>
      </div>
      {importPreview && (
        <small>
          {importPreview.systemsToCreate} systems to create, {importPreview.systemsToUpdate} to update,{" "}
          {importPreview.rolesToCreate} roles to create
        </small>
      )}
      {hasExportedSeed && <small>Current RBAC seed export is available from the API.</small>}
    </section>
  );
}
