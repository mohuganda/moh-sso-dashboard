import { useMemo, useState } from "react";
import { ContentSwitcher, InlineNotification, Switch } from "@carbon/react";

import {
  useApplyRbacImportMutation,
  useCreateChangeRequestMutation,
  useExportRbacSeedQuery,
  useListAccessRequestsQuery,
  useListChangeRequestsQuery,
  useListRbacPermissionsQuery,
  useListRbacRoleTemplatesQuery,
  usePreviewRbacChangeMutation,
  usePreviewRbacImportMutation,
  useSimulateRbacAccessMutation,
  useUpdateRbacPermissionMetadataMutation,
} from "../api";
import type { RbacSimulationResult } from "../types";

import { AccessRequestsPanel } from "./AccessRequestsPanel";
import { AuditTrailPanel } from "./AuditTrailPanel";
import { ChangeRequestsPanel } from "./ChangeRequestsPanel";
import { ImportExportPanel } from "./ImportExportPanel";
import { PermissionCatalogPanel } from "./PermissionCatalogPanel";
import { PolicySimulatorPanel } from "./PolicySimulatorPanel";
import { RoleTemplatesPanel } from "./RoleTemplatesPanel";
import { HealthContextsPanel } from "./HealthContextsPanel";

type GovernanceSection =
  | "permissions"
  | "import-export"
  | "templates"
  | "changes"
  | "access"
  | "audit"
  | "health-contexts"
  | "simulator";

type GovernanceToolsPanelProps = {
  activeClientId: string;
};

export function GovernanceToolsPanel({ activeClientId }: GovernanceToolsPanelProps) {
  const { data: permissions = [] } = useListRbacPermissionsQuery();
  const { data: templates = [] } = useListRbacRoleTemplatesQuery();
  const { data: accessRequests = [] } = useListAccessRequestsQuery();
  const { data: changeRequests = [] } = useListChangeRequestsQuery();
  const { data: exportedSeed } = useExportRbacSeedQuery();

  const [selectedPermission, setSelectedPermission] = useState("");
  const [permissionDescription, setPermissionDescription] = useState("");
  const [seedText, setSeedText] = useState("");
  const [simulationUser, setSimulationUser] = useState("");
  const [simulationRoles, setSimulationRoles] = useState("admin");
  const [simulationClientRoles, setSimulationClientRoles] = useState("outbreak-management:viewer");
  const [addRealmRoles, setAddRealmRoles] = useState("");
  const [removeRealmRoles, setRemoveRealmRoles] = useState("");
  const [addClientRoles, setAddClientRoles] = useState("");
  const [removeClientRoles, setRemoveClientRoles] = useState("");
  const [addPermissions, setAddPermissions] = useState("");
  const [removePermissions, setRemovePermissions] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [simulation, setSimulation] = useState<RbacSimulationResult | null>(null);
  const [activeSection, setActiveSection] = useState<GovernanceSection>("permissions");

  const [updatePermission] = useUpdateRbacPermissionMetadataMutation();
  const [previewImport, { data: importPreview }] = usePreviewRbacImportMutation();
  const [applyImport] = useApplyRbacImportMutation();
  const [previewChange, { data: changePreview }] = usePreviewRbacChangeMutation();
  const [createChangeRequest] = useCreateChangeRequestMutation();
  const [simulateAccess] = useSimulateRbacAccessMutation();

  const selectedPermissionData = useMemo(
    () => permissions.find((permission) => permission.key === selectedPermission),
    [permissions, selectedPermission],
  );

  const handleUpdatePermission = async () => {
    if (!selectedPermission) return;
    await updatePermission({
      permissionKey: selectedPermission,
      data: {
        displayName: selectedPermissionData?.displayName,
        category: selectedPermissionData?.category,
        status: selectedPermissionData?.status || "active",
        description: permissionDescription,
      },
    }).unwrap();
    setMessage("Permission metadata updated.");
  };

  const parseSeedPayload = () => {
    try {
      return JSON.parse(seedText);
    } catch {
      return seedText;
    }
  };

  const handlePreviewImport = async () => {
    await previewImport({ payload: parseSeedPayload() }).unwrap();
  };

  const handleApplyImport = async () => {
    if (!confirm("Apply this RBAC seed import? This will create/update mappings but will not prune stale records.")) {
      return;
    }
    await applyImport({ payload: parseSeedPayload() }).unwrap();
    setMessage("RBAC seed import applied.");
  };

  const handlePreviewRisk = async () => {
    await previewChange({
      action: "delete-role",
      resourceType: "system-role",
      resourceId: "preview",
    }).unwrap();
  };

  const handleCreateChangeRequest = async () => {
    await createChangeRequest({
      action: "delete-role",
      resourceType: "system-role",
      resourceId: "manual-review",
      riskLevel: "high",
      reason: "Manual RBAC governance review",
    }).unwrap();
    setMessage("Change request created.");
  };

  const handleSimulate = async () => {
    const toList = (value: string) =>
      value
      .split(",")
      .map((role) => role.trim())
      .filter(Boolean);

    const toClientRoles = (value: string) =>
      toList(value).reduce<Record<string, string[]>>((acc, entry) => {
        const [clientId, role] = entry.split(":");
        if (clientId && role) {
          acc[clientId] = [...(acc[clientId] ?? []), role];
        }
        return acc;
      }, {});

    const userValue = simulationUser.trim();
    const userIdentity = userValue.includes("@")
      ? { email: userValue }
      : userValue.length > 30
        ? { userId: userValue }
        : userValue
          ? { username: userValue }
          : {};

    setSimulation(
      await simulateAccess({
        ...userIdentity,
        realmRoles: toList(simulationRoles),
        clientRoles: toClientRoles(simulationClientRoles),
        addRealmRoles: toList(addRealmRoles),
        removeRealmRoles: toList(removeRealmRoles),
        addClientRoles: toClientRoles(addClientRoles),
        removeClientRoles: toClientRoles(removeClientRoles),
        addPermissions: toList(addPermissions),
        removePermissions: toList(removePermissions),
      }).unwrap(),
    );
  };

  return (
    <section className="rbac-card">
      <div className="rbac-page__panel-header">
        <div>
          <h2>Governance Tools</h2>
          <p>Catalog, contextual access, audit, requests, approvals, templates, and simulation.</p>
        </div>
      </div>

      {message && (
        <InlineNotification
          kind="success"
          lowContrast
          title="RBAC governance"
          subtitle={message}
          onClose={() => setMessage(null)}
        />
      )}

      <div className="rbac-governance-tabs">
        <ContentSwitcher
          onChange={(item) => setActiveSection((item?.name as GovernanceSection) || "permissions")}
          selectedIndex={0}
          size="md"
        >
          <Switch name="permissions" text="Permissions" />
          <Switch name="import-export" text="Import / Export" />
          <Switch name="templates" text="Templates" />
          <Switch name="changes" text="Changes" />
          <Switch name="access" text="Access" />
          <Switch name="audit" text="Audit" />
          <Switch name="health-contexts" text="Health contexts" />
          <Switch name="simulator" text="Simulator" />
        </ContentSwitcher>
      </div>

      <div className="rbac-governance-grid">
        {activeSection === "permissions" && (
          <PermissionCatalogPanel
            permissions={permissions}
            selectedPermission={selectedPermission}
            selectedPermissionData={selectedPermissionData}
            permissionDescription={permissionDescription}
            onSelectedPermissionChange={setSelectedPermission}
            onPermissionDescriptionChange={setPermissionDescription}
            onSave={handleUpdatePermission}
          />
        )}
        {activeSection === "import-export" && (
          <ImportExportPanel
            seedText={seedText}
            importPreview={importPreview}
            exportedSeed={exportedSeed}
            onSeedTextChange={setSeedText}
            onPreview={handlePreviewImport}
            onApply={handleApplyImport}
          />
        )}
        {activeSection === "templates" && (
          <RoleTemplatesPanel
            templates={templates}
            activeClientId={activeClientId}
            onMessage={setMessage}
          />
        )}
        {activeSection === "changes" && (
          <ChangeRequestsPanel
            changeRequests={changeRequests}
            changePreview={changePreview}
            onPreviewRisk={handlePreviewRisk}
            onCreateRequest={handleCreateChangeRequest}
          />
        )}
        {activeSection === "access" && <AccessRequestsPanel accessRequests={accessRequests} />}
        {activeSection === "audit" && <AuditTrailPanel />}
        {activeSection === "health-contexts" && <HealthContextsPanel />}
        {activeSection === "simulator" && (
          <PolicySimulatorPanel
            simulationRoles={simulationRoles}
            simulationClientRoles={simulationClientRoles}
            simulationUser={simulationUser}
            addRealmRoles={addRealmRoles}
            removeRealmRoles={removeRealmRoles}
            addClientRoles={addClientRoles}
            removeClientRoles={removeClientRoles}
            addPermissions={addPermissions}
            removePermissions={removePermissions}
            simulation={simulation}
            onSimulationRolesChange={setSimulationRoles}
            onSimulationClientRolesChange={setSimulationClientRoles}
            onSimulationUserChange={setSimulationUser}
            onAddRealmRolesChange={setAddRealmRoles}
            onRemoveRealmRolesChange={setRemoveRealmRoles}
            onAddClientRolesChange={setAddClientRoles}
            onRemoveClientRolesChange={setRemoveClientRoles}
            onAddPermissionsChange={setAddPermissions}
            onRemovePermissionsChange={setRemovePermissions}
            onSimulate={handleSimulate}
          />
        )}
      </div>
    </section>
  );
}
