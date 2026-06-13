import { Button, FileUploaderButton, TextArea } from "@carbon/react";

import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacImportPreview } from "@moh-sso/types";

type ImportExportPanelProps = {
  seedText: string;
  importPreview?: RbacImportPreview;
  exportedSeed?: unknown;
  onSeedTextChange: (value: string) => void;
  onPreview: () => void;
  onApply: () => void;
};

export function ImportExportPanel({
  seedText,
  importPreview,
  exportedSeed,
  onSeedTextChange,
  onPreview,
  onApply,
}: ImportExportPanelProps) {
  const handleDownload = () => {
    if (!exportedSeed) return;
    const blob = new Blob([JSON.stringify(exportedSeed, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "system-rbac.seed.json";
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const handleUpload = async (_event: unknown, details?: { addedFiles?: Array<File> }) => {
    const file = details?.addedFiles?.[0];
    if (!file) return;
    onSeedTextChange(await file.text());
  };

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
        <Button size="sm" kind="secondary" disabled={!exportedSeed} onClick={handleDownload}>
          Download seed
        </Button>
        <FileUploaderButton
          accept={[".json", ".yaml", ".yml"]}
          buttonKind="tertiary"
          labelText="Upload seed"
          multiple={false}
          onChange={handleUpload}
          size="sm"
        />
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
      {exportedSeed ? <small>Current RBAC seed export is available from the API.</small> : null}
    </section>
  );
}
