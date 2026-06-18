import {
  Button,
  DataTable,
  Form,
  FormGroup,
  InlineLoading,
  Search,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
  TextArea,
  TextInput,
  Toggle,
} from "@carbon/react";
import { Launch, Save } from "@carbon/react/icons";
import { useMemo, useState } from "react";

import {
  useGetRbacSystemQuery,
  useListRbacSystemsQuery,
  useUpdateRbacSystemMutation,
} from "@moh-sso/api";
import type { RbacSystem, UpsertRbacSystemPayload } from "@moh-sso/types";
import {
  DataTableShell,
  ErrorState,
  RowActionsCell,
  TableStatusTag,
  useHeaderPanel,
  useToast,
} from "@moh-sso/ui";

import "./systems.scss";

const headers = [
  { key: "displayName", header: "System" },
  { key: "clientId", header: "Client ID" },
  { key: "category", header: "Category" },
  { key: "environment", header: "Environment" },
  { key: "criticality", header: "Criticality" },
  { key: "enabled", header: "Status" },
  { key: "launchUrl", header: "Launch URL" },
  { key: "actions", header: "" },
  { key: "raw", header: "" },
];

type SystemDraft = UpsertRbacSystemPayload & {
  enabled: boolean;
};

function toDraft(system: RbacSystem): SystemDraft {
  return {
    displayName: system.displayName || system.clientId,
    description: system.description || "",
    icon: system.icon || "",
    launchUrl: system.launchUrl || "",
    category: system.category || "",
    ownerTeam: system.ownerTeam || "",
    ownerName: system.ownerName || "",
    ownerEmail: system.ownerEmail || "",
    supportUrl: system.supportUrl || "",
    documentationUrl: system.documentationUrl || "",
    environment: system.environment || "",
    criticality: system.criticality || "",
    navigation: system.navigation || "",
    enabled: system.enabled,
    sortOrder: system.sortOrder ?? 0,
  };
}

export default function SystemsPage() {
  const { data: systems = [], isLoading, isError, error, refetch } = useListRbacSystemsQuery();
  const [search, setSearch] = useState("");
  const { openPanel, closePanel } = useHeaderPanel();

  const rows = useMemo(
    () =>
      systems
        .filter((system) => {
          const query = search.trim().toLowerCase();
          if (!query) return true;

          return [
            system.displayName,
            system.clientId,
            system.category,
            system.environment,
            system.criticality,
            system.ownerTeam,
          ]
            .filter(Boolean)
            .some((value) => String(value).toLowerCase().includes(query));
        })
        .map((system) => ({
          id: system.clientId,
          displayName: system.displayName || system.clientId,
          clientId: system.clientId,
          category: system.category || "-",
          environment: system.environment || "-",
          criticality: system.criticality || "-",
          enabled: system.enabled ? "Enabled" : "Disabled",
          launchUrl: system.launchUrl || "-",
          actions: "",
          raw: system,
        })),
    [search, systems],
  );

  const enabledCount = systems.filter((system) => system.enabled).length;

  return (
    <DataTableShell
      title="Systems"
      description="Connected Keycloak clients registered as portal systems."
      rows={rows}
      headers={headers}
      getRowId={(row) => row.id}
      isLoading={isLoading}
      loadingDescription="Loading systems..."
      emptyTitle="No systems found"
      emptyDescription="No connected systems match the selected filters."
      tableState={
        isError ? (
          <ErrorState
            title="Failed to load systems"
            description={getApiErrorMessage(error, "Unable to fetch system registry.")}
            primaryAction={{ label: "Retry", onClick: refetch }}
          />
        ) : undefined
      }
      topContent={
        <div className="systems-summary">
          <SummaryTile label="Registered" value={systems.length} />
          <SummaryTile label="Enabled" value={enabledCount} />
          <SummaryTile label="Disabled" value={systems.length - enabledCount} />
        </div>
      }
      filters={
        <Search
          id="systems-search"
          labelText="Search systems"
          placeholder="Search by system, client ID, category..."
          value={search}
          onChange={(event) => setSearch(event.target.value)}
        />
      }
    >
      {({ rows, headers }) => (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps }) => (
            <Table size="lg">
              <TableHead>
                <TableRow>
                  {headers
                    .filter((header) => header.key !== "raw")
                    .map((header) => (
                      <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                    ))}
                </TableRow>
              </TableHead>

              <TableBody>
                {rows.map((row) => {
                  const system = row.cells.find((cell) => cell.info.header === "raw")
                    ?.value as RbacSystem;

                  return (
                    <TableRow {...getRowProps({ row })}>
                      {row.cells.map((cell) => {
                        if (cell.info.header === "raw") return null;

                        if (cell.info.header === "enabled") {
                          return (
                            <TableCell key={cell.id}>
                              <TableStatusTag
                                status={String(cell.value)}
                                kind={system.enabled ? "green" : "red"}
                              />
                            </TableCell>
                          );
                        }

                        if (cell.info.header === "criticality") {
                          return (
                            <TableCell key={cell.id}>
                              <Tag size="sm" type={getCriticalityTag(system.criticality)}>
                                {system.criticality || "Unclassified"}
                              </Tag>
                            </TableCell>
                          );
                        }

                        if (cell.info.header === "actions") {
                          return (
                            <RowActionsCell key={cell.id}>
                              <Button
                                size="sm"
                                kind="ghost"
                                renderIcon={Launch}
                                onClick={() => {
                                  openPanel({
                                    title: `System: ${system.displayName || system.clientId}`,
                                    size: "lg",
                                    content: (
                                      <SystemPanel
                                        clientId={system.clientId}
                                        onClose={closePanel}
                                      />
                                    ),
                                  });
                                }}
                              >
                                Manage
                              </Button>
                            </RowActionsCell>
                          );
                        }

                        return <TableCell key={cell.id}>{cell.value || "-"}</TableCell>;
                      })}
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </DataTable>
      )}
    </DataTableShell>
  );
}

function SummaryTile({ label, value }: { label: string; value: number }) {
  return (
    <div className="systems-summary__item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function SystemPanel({ clientId, onClose }: { clientId: string; onClose: () => void }) {
  const toast = useToast();
  const { data: system, isLoading, isError, error } = useGetRbacSystemQuery(clientId);
  const [updateSystem, { isLoading: saving }] = useUpdateRbacSystemMutation();
  const [draft, setDraft] = useState<SystemDraft | null>(null);

  const currentDraft = draft ?? (system ? toDraft(system) : null);

  if (isLoading) {
    return <InlineLoading description="Loading system..." />;
  }

  if (isError || !system || !currentDraft) {
    return (
      <ErrorState
        title="Unable to load system"
        description={getApiErrorMessage(error, "Check RBAC sync and try again.")}
      />
    );
  }

  const updateField = <K extends keyof SystemDraft>(key: K, value: SystemDraft[K]) => {
    setDraft({ ...currentDraft, [key]: value });
  };

  const handleSave = async () => {
    try {
      await updateSystem({ clientId, data: currentDraft }).unwrap();
      toast.success("System updated", `${currentDraft.displayName} was saved.`);
      onClose();
    } catch (err) {
      toast.error("Save failed", getApiErrorMessage(err, "Unable to update system."));
    }
  };

  return (
    <Form className="system-panel">
      <Stack gap={6}>
        <FormGroup legendText="System metadata">
          <Stack gap={4}>
            <TextInput id="system-client-id" labelText="Client ID" value={system.clientId} disabled />
            <TextInput
              id="system-display-name"
              labelText="Display name"
              value={currentDraft.displayName}
              onChange={(event) => updateField("displayName", event.target.value)}
            />
            <TextArea
              id="system-description"
              labelText="Description"
              value={currentDraft.description}
              onChange={(event) => updateField("description", event.target.value)}
            />
            <TextInput
              id="system-launch-url"
              labelText="Launch URL"
              value={currentDraft.launchUrl}
              onChange={(event) => updateField("launchUrl", event.target.value)}
            />
            <Stack orientation="horizontal" gap={4}>
              <TextInput
                id="system-category"
                labelText="Category"
                value={currentDraft.category}
                onChange={(event) => updateField("category", event.target.value)}
              />
              <TextInput
                id="system-icon"
                labelText="Icon"
                value={currentDraft.icon}
                onChange={(event) => updateField("icon", event.target.value)}
              />
            </Stack>
            <TextArea
              id="system-navigation"
              labelText="Navigation JSON"
              value={currentDraft.navigation}
              onChange={(event) => updateField("navigation", event.target.value)}
            />
            <Stack orientation="horizontal" gap={4}>
              <TextInput
                id="system-environment"
                labelText="Environment"
                value={currentDraft.environment}
                onChange={(event) => updateField("environment", event.target.value)}
              />
              <TextInput
                id="system-criticality"
                labelText="Criticality"
                value={currentDraft.criticality}
                onChange={(event) => updateField("criticality", event.target.value)}
              />
            </Stack>
            <Toggle
              id="system-enabled"
              labelText="Enabled"
              toggled={currentDraft.enabled}
              onToggle={(enabled) => updateField("enabled", enabled)}
            />
          </Stack>
        </FormGroup>

        <FormGroup legendText="Ownership">
          <Stack gap={4}>
            <TextInput
              id="system-owner-team"
              labelText="Owner team"
              value={currentDraft.ownerTeam}
              onChange={(event) => updateField("ownerTeam", event.target.value)}
            />
            <Stack orientation="horizontal" gap={4}>
              <TextInput
                id="system-owner-name"
                labelText="Owner name"
                value={currentDraft.ownerName}
                onChange={(event) => updateField("ownerName", event.target.value)}
              />
              <TextInput
                id="system-owner-email"
                labelText="Owner email"
                value={currentDraft.ownerEmail}
                onChange={(event) => updateField("ownerEmail", event.target.value)}
              />
            </Stack>
            <TextInput
              id="system-support-url"
              labelText="Support URL"
              value={currentDraft.supportUrl}
              onChange={(event) => updateField("supportUrl", event.target.value)}
            />
            <TextInput
              id="system-documentation-url"
              labelText="Documentation URL"
              value={currentDraft.documentationUrl}
              onChange={(event) => updateField("documentationUrl", event.target.value)}
            />
          </Stack>
        </FormGroup>

        <FormGroup legendText="RBAC mapping">
          <div className="system-panel__mapping">
            <SummaryTile label="System roles" value={system.roles.length} />
            <SummaryTile label="Access roles" value={system.accessRoles.length} />
          </div>

          <div className="system-panel__tags">
            {system.accessRoles.map((role) => (
              <Tag key={role} type="cyan">
                {role}
              </Tag>
            ))}
            {system.accessRoles.length === 0 && <span>No access roles configured.</span>}
          </div>
        </FormGroup>

        <Button renderIcon={Save} disabled={saving} onClick={handleSave}>
          {saving ? "Saving..." : "Save system"}
        </Button>
      </Stack>
    </Form>
  );
}

function getCriticalityTag(criticality?: string) {
  const value = criticality?.toLowerCase();
  if (value === "critical" || value === "high") return "red";
  if (value === "medium") return "purple";
  if (value === "low") return "green";
  return "gray";
}

function getApiErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== "object") {
    return fallback;
  }

  const maybeError = error as {
    data?: {
      message?: unknown;
      error?: {
        message?: unknown;
      };
    };
    error?: unknown;
  };

  if (typeof maybeError.data?.message === "string") return maybeError.data.message;
  if (typeof maybeError.data?.error?.message === "string") return maybeError.data.error.message;
  if (typeof maybeError.error === "string") return maybeError.error;
  return fallback;
}
