import { useEffect, useMemo, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Search,
  Select,
  SelectItem,
  Tag,
  TextInput,
  Toggle,
} from "@carbon/react";
import { Add, Save, TrashCan } from "@carbon/react/icons";

import { useListUsersQuery } from "@moh-sso/users";
import { useModal, useToast } from "@moh-sso/ui";

import {
  useCreateHealthContextMutation,
  useApplyHealthContextSyncMutation,
  useDeleteHealthContextAliasMutation,
  useDeleteHealthContextMutation,
  useListGroupHealthContextsQuery,
  useGetHealthContextDriftQuery,
  useListHealthContextAliasesQuery,
  useListHealthContextsQuery,
  useListRbacGroupsQuery,
  useListUserHealthContextsQuery,
  useReplaceGroupHealthContextsMutation,
  useReplaceUserHealthContextsMutation,
  usePreviewHealthContextSyncMutation,
  useUpdateHealthContextMutation,
  useUpsertHealthContextAliasMutation,
} from "../api";
import type {
  HealthContextAssignment,
  HealthContextAssignmentPayload,
  HealthContextNode,
  HealthContextNodePayload,
  HealthContextScopeMode,
  HealthContextSyncMapping,
  HealthContextType,
} from "../types";

function HealthContextDriftPanel({
  contexts,
}: {
  contexts: HealthContextNode[];
}) {
  const toast = useToast();
  const { openModal, closeModal } = useModal();
  const drift = useGetHealthContextDriftQuery();
  const [previewSync, preview] = usePreviewHealthContextSyncMutation();
  const [applySync, applyState] = useApplyHealthContextSyncMutation();
  const [mappings, setMappings] = useState<Record<string, HealthContextSyncMapping>>({});

  const driftedItems = (drift.data?.items ?? []).filter((item) => item.status !== "IN_SYNC");
  const selectedMappings = Object.values(mappings).filter(
    (mapping) => mapping.contextNodeId,
  );
  const request = { mappings: selectedMappings, replaceExisting: false };

  const updateMapping = (
    groupId: string,
    values: Partial<HealthContextSyncMapping>,
  ) => {
    setMappings((current) => ({
      ...current,
      [groupId]: {
        groupId,
        contextNodeId: current[groupId]?.contextNodeId ?? "",
        scopeMode: current[groupId]?.scopeMode ?? "NODE_ONLY",
        ...values,
      },
    }));
    preview.reset();
  };

  const confirmApply = () => {
    if (!preview.data?.valid || selectedMappings.length === 0) return;
    openModal({
      title: "Apply health-context mappings",
      content: (
        <p>
          Apply {selectedMappings.length} explicit group mapping
          {selectedMappings.length === 1 ? "" : "s"}? Existing mappings not listed here
          will be retained.
        </p>
      ),
      primaryAction: {
        label: "Apply mappings",
        onClick: async () => {
          try {
            const result = await applySync(request).unwrap();
            closeModal();
            setMappings({});
            preview.reset();
            toast.success(
              "Health-context mappings applied",
              `${result.appliedMappings} mappings across ${result.affectedGroups} groups.`,
            );
          } catch (error) {
            toast.error(
              "Unable to apply mappings",
              apiError(error, "No mappings were changed."),
            );
          }
        },
      },
      secondaryAction: { label: "Cancel", onClick: closeModal },
      onClose: closeModal,
    });
  };

  return (
    <section className="health-context-drift">
      <div className="rbac-page__panel-header">
        <div>
          <h3>Group mapping drift</h3>
          <p>Compare Keycloak-linked groups with portal health-context mappings.</p>
        </div>
        <Button kind="ghost" size="sm" onClick={() => void drift.refetch()}>
          Refresh
        </Button>
      </div>
      {drift.isLoading ? <InlineLoading description="Checking mappings" /> : null}
      {drift.isError ? (
        <InlineNotification
          kind="error"
          lowContrast
          hideCloseButton
          title="Drift check unavailable"
          subtitle="The portal could not inspect group mappings."
        />
      ) : null}
      {drift.data ? (
        <div className="health-context-drift__summary">
          <Tag type="green">{drift.data.inSync} in sync</Tag>
          <Tag type={drift.data.drifted > 0 ? "red" : "gray"}>
            {drift.data.drifted} require review
          </Tag>
        </div>
      ) : null}
      <div className="health-context-drift__items">
        {driftedItems.map((item) => (
          <div className="health-context-drift__item" key={`${item.groupId}:${item.contextNodeId ?? "none"}`}>
            <div>
              <strong>{item.groupPath}</strong>
              <p>{item.recommendedAction}</p>
              <Tag type={item.status === "MISSING_KEYCLOAK_LINK" ? "red" : "warm-gray"}>
                {item.status.replaceAll("_", " ")}
              </Tag>
            </div>
            <Select
              id={`health-context-drift-context-${item.groupId}`}
              labelText="Health context"
              value={mappings[item.groupId]?.contextNodeId ?? ""}
              disabled={
                item.status === "MISSING_KEYCLOAK_LINK" ||
                item.status === "DISABLED_GROUP"
              }
              onChange={(event) =>
                updateMapping(item.groupId, { contextNodeId: event.target.value })
              }
            >
              <SelectItem value="" text="Select context" />
              {contexts
                .filter((context) => context.enabled)
                .map((context) => (
                  <SelectItem
                    key={context.id}
                    value={context.id}
                    text={`${context.name} (${context.code})`}
                  />
                ))}
            </Select>
            <Select
              id={`health-context-drift-scope-${item.groupId}`}
              labelText="Scope"
              value={mappings[item.groupId]?.scopeMode ?? "NODE_ONLY"}
              disabled={!mappings[item.groupId]?.contextNodeId}
              onChange={(event) =>
                updateMapping(item.groupId, {
                  scopeMode: event.target.value as HealthContextScopeMode,
                })
              }
            >
              <SelectItem value="NODE_ONLY" text="Node only" />
              <SelectItem
                value="NODE_AND_DESCENDANTS"
                text="Node and descendants"
              />
            </Select>
          </div>
        ))}
        {!drift.isLoading && driftedItems.length === 0 ? (
          <p>All linked groups with health-context mappings are in sync.</p>
        ) : null}
      </div>
      <div className="health-context-drift__actions">
        <Button
          kind="tertiary"
          disabled={selectedMappings.length === 0 || preview.isLoading}
          onClick={async () => {
            try {
              await previewSync(request).unwrap();
              toast.success("Sync preview ready");
            } catch (error) {
              toast.error(
                "Unable to preview mappings",
                apiError(error, "Review the selected contexts and try again."),
              );
            }
          }}
        >
          Preview changes
        </Button>
        <Button
          disabled={!preview.data?.valid || applyState.isLoading}
          onClick={confirmApply}
        >
          Apply previewed mappings
        </Button>
      </div>
    </section>
  );
}

const contextTypes: HealthContextType[] = [
  "NATIONAL",
  "REGION",
  "DISTRICT",
  "CITY",
  "DIVISION",
  "MUNICIPALITY",
  "COUNTY",
  "SUB_COUNTY",
  "PARISH",
  "FACILITY",
  "PROGRAM",
  "DEPARTMENT",
  "TEAM",
  "CUSTOM",
];

const emptyDraft: HealthContextNodePayload = {
  code: "",
  name: "",
  contextType: "CUSTOM",
  parentId: null,
  source: "PORTAL",
  metadata: {},
  enabled: true,
};

function apiError(error: unknown, fallback: string) {
  if (typeof error !== "object" || error === null) return fallback;
  const value = error as { data?: { error?: { message?: string }; message?: string } };
  return value.data?.error?.message || value.data?.message || fallback;
}

function assignmentPayload(
  assignment: HealthContextAssignment,
): HealthContextAssignmentPayload {
  return {
    contextNodeId: assignment.contextNodeId,
    scopeMode: assignment.scopeMode,
    isDefault: assignment.isDefault,
    validFrom: assignment.validFrom,
    validUntil: assignment.validUntil,
    source: assignment.source || "PORTAL",
    sourceReference: assignment.sourceReference,
  };
}

function ContextAssignmentEditor({
  kind,
  contexts,
}: {
  kind: "user" | "group";
  contexts: HealthContextNode[];
}) {
  const toast = useToast();
  const { data: users = [] } = useListUsersQuery();
  const { data: groups = [] } = useListRbacGroupsQuery();
  const targets =
    kind === "user"
      ? users.map((user) => ({
          id: user.id,
          label: user.username || user.email || user.id,
        }))
      : groups.map((group) => ({
          id: group.id,
          label: group.displayName || group.name || group.path,
        }));
  const [targetId, setTargetId] = useState("");
  const [contextId, setContextId] = useState("");
  const [scopeMode, setScopeMode] =
    useState<HealthContextScopeMode>("NODE_ONLY");

  const userAssignments = useListUserHealthContextsQuery(targetId, {
    skip: kind !== "user" || !targetId,
  });
  const groupAssignments = useListGroupHealthContextsQuery(targetId, {
    skip: kind !== "group" || !targetId,
  });
  const assignments =
    kind === "user" ? (userAssignments.data ?? []) : (groupAssignments.data ?? []);
  const loading =
    kind === "user" ? userAssignments.isFetching : groupAssignments.isFetching;
  const [replaceUser, userMutation] = useReplaceUserHealthContextsMutation();
  const [replaceGroup, groupMutation] = useReplaceGroupHealthContextsMutation();

  const replace = async (next: HealthContextAssignmentPayload[]) => {
    if (!targetId) return;
    try {
      if (kind === "user") {
        await replaceUser({ targetId, assignments: next }).unwrap();
      } else {
        await replaceGroup({ targetId, assignments: next }).unwrap();
      }
      toast.success("Health-context assignments updated");
    } catch (error) {
      toast.error(
        "Unable to update assignments",
        apiError(error, "The assignment could not be saved."),
      );
    }
  };

  const addAssignment = () => {
    if (!contextId || assignments.some((item) => item.contextNodeId === contextId)) return;
    void replace([
      ...assignments.map(assignmentPayload),
      {
        contextNodeId: contextId,
        scopeMode,
        isDefault: kind === "user" && assignments.length === 0,
        source: "PORTAL",
      },
    ]);
  };

  const removeAssignment = (assignmentId: string) => {
    void replace(
      assignments
        .filter((assignment) => assignment.id !== assignmentId)
        .map(assignmentPayload),
    );
  };

  return (
    <section className="health-context-assignments">
      <h4>{kind === "user" ? "User assignments" : "Group mappings"}</h4>
      <Select
        id={`health-context-${kind}-target`}
        labelText={kind === "user" ? "User" : "Keycloak group"}
        value={targetId}
        onChange={(event) => setTargetId(event.target.value)}
      >
        <SelectItem value="" text={`Select ${kind}`} />
        {targets.map((target) => (
          <SelectItem key={target.id} value={target.id} text={target.label} />
        ))}
      </Select>
      {loading ? <InlineLoading description="Loading assignments" /> : null}
      {targetId ? (
        <>
          <div className="health-context-assignments__add">
            <Select
              id={`health-context-${kind}-node`}
              labelText="Health context"
              value={contextId}
              onChange={(event) => setContextId(event.target.value)}
            >
              <SelectItem value="" text="Select context" />
              {contexts
                .filter((context) => context.enabled)
                .map((context) => (
                  <SelectItem
                    key={context.id}
                    value={context.id}
                    text={`${context.name} (${context.code})`}
                  />
                ))}
            </Select>
            <Select
              id={`health-context-${kind}-scope`}
              labelText="Scope"
              value={scopeMode}
              onChange={(event) =>
                setScopeMode(event.target.value as HealthContextScopeMode)
              }
            >
              <SelectItem value="NODE_ONLY" text="This context only" />
              <SelectItem
                value="NODE_AND_DESCENDANTS"
                text="Context and descendants"
              />
            </Select>
            <Button
              size="md"
              renderIcon={Add}
              disabled={
                !contextId ||
                assignments.some((item) => item.contextNodeId === contextId) ||
                userMutation.isLoading ||
                groupMutation.isLoading
              }
              onClick={addAssignment}
            >
              Assign
            </Button>
          </div>
          <div className="health-context-assignments__list">
            {assignments.length === 0 ? (
              <p>No direct assignments. Group-derived access remains separate.</p>
            ) : (
              assignments.map((assignment) => {
                const context = contexts.find(
                  (item) => item.id === assignment.contextNodeId,
                );
                return (
                  <div key={assignment.id}>
                    <span>
                      {context?.name || assignment.contextNodeId}
                      <small>
                        {assignment.scopeMode === "NODE_AND_DESCENDANTS"
                          ? "Includes descendants"
                          : "Node only"}
                      </small>
                    </span>
                    <Button
                      hasIconOnly
                      kind="ghost"
                      size="sm"
                      renderIcon={TrashCan}
                      iconDescription="Remove assignment"
                      onClick={() => removeAssignment(assignment.id)}
                    />
                  </div>
                );
              })
            )}
          </div>
        </>
      ) : null}
    </section>
  );
}

export function HealthContextsPanel() {
  const toast = useToast();
  const { openModal, closeModal } = useModal();
  const { data: contexts = [], isLoading } = useListHealthContextsQuery({
    includeDisabled: true,
  });
  const [selectedId, setSelectedId] = useState("");
  const [search, setSearch] = useState("");
  const [creating, setCreating] = useState(false);
  const selected = contexts.find((context) => context.id === selectedId);
  const [draft, setDraft] = useState<HealthContextNodePayload>(emptyDraft);
  const [aliasNamespace, setAliasNamespace] = useState("");
  const [aliasExternalId, setAliasExternalId] = useState("");
  const aliases = useListHealthContextAliasesQuery(selectedId, {
    skip: !selectedId || creating,
  });
  const [createContext, createState] = useCreateHealthContextMutation();
  const [updateContext, updateState] = useUpdateHealthContextMutation();
  const [deleteContext] = useDeleteHealthContextMutation();
  const [upsertAlias] = useUpsertHealthContextAliasMutation();
  const [deleteAlias] = useDeleteHealthContextAliasMutation();

  useEffect(() => {
    if (!selected || creating) return;
    setDraft({
      code: selected.code,
      name: selected.name,
      contextType: selected.contextType,
      parentId: selected.parentId ?? null,
      source: selected.source,
      metadata: selected.metadata ?? {},
      enabled: selected.enabled,
      version: selected.version,
    });
  }, [creating, selected]);

  const visibleContexts = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return contexts;
    return contexts.filter((context) =>
      [context.code, context.name, context.contextType]
        .join(" ")
        .toLowerCase()
        .includes(term),
    );
  }, [contexts, search]);

  const beginCreate = () => {
    setCreating(true);
    setSelectedId("");
    setDraft({ ...emptyDraft });
  };

  const save = async () => {
    try {
      if (creating) {
        const result = await createContext(draft).unwrap();
        setSelectedId(result.id);
        setCreating(false);
        toast.success("Health context created");
      } else if (selected) {
        await updateContext({
          contextId: selected.id,
          body: { ...draft, version: selected.version },
        }).unwrap();
        toast.success("Health context updated");
      }
    } catch (error) {
      toast.error(
        "Unable to save health context",
        apiError(error, "Check the hierarchy values and try again."),
      );
    }
  };

  const confirmDelete = () => {
    if (!selected) return;
    openModal({
      title: "Delete health context",
      content: (
        <p>
          Delete <strong>{selected.name}</strong>? Contexts with children or assignments
          cannot be deleted.
        </p>
      ),
      primaryAction: {
        label: "Delete",
        kind: "danger",
        onClick: async () => {
          try {
            await deleteContext(selected.id).unwrap();
            closeModal();
            setSelectedId("");
            toast.success("Health context deleted");
          } catch (error) {
            toast.error(
              "Unable to delete health context",
              apiError(error, "Remove child contexts and assignments first."),
            );
          }
        },
      },
      secondaryAction: { label: "Cancel", onClick: closeModal },
      onClose: closeModal,
    });
  };

  const saveAlias = async () => {
    if (!selected || !aliasNamespace.trim() || !aliasExternalId.trim()) return;
    try {
      await upsertAlias({
        contextId: selected.id,
        namespace: aliasNamespace.trim(),
        externalId: aliasExternalId.trim(),
      }).unwrap();
      setAliasNamespace("");
      setAliasExternalId("");
      toast.success("External identifier mapped");
    } catch (error) {
      toast.error("Unable to map identifier", apiError(error, "The alias was not saved."));
    }
  };

  return (
    <div className="health-context-admin">
      <div className="health-context-admin__browser">
        <div className="rbac-page__panel-header">
          <div>
            <h3>Health hierarchy</h3>
            <p>Authorization scopes managed by the portal database.</p>
          </div>
          <Button hasIconOnly kind="ghost" renderIcon={Add} iconDescription="Create context" onClick={beginCreate} />
        </div>
        <Search
          id="health-context-search"
          labelText="Search health contexts"
          placeholder="Search code, name, or type"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
        />
        {isLoading ? <InlineLoading description="Loading hierarchy" /> : null}
        <div className="health-context-admin__nodes">
          {visibleContexts.map((context) => (
            <button
              type="button"
              key={context.id}
              className={context.id === selectedId ? "is-active" : ""}
              onClick={() => {
                setCreating(false);
                setSelectedId(context.id);
              }}
            >
              <span>{context.name}</span>
              <small>
                {context.code} · {context.contextType}
              </small>
              {!context.enabled ? <Tag type="gray">Disabled</Tag> : null}
            </button>
          ))}
        </div>
      </div>

      <div className="health-context-admin__detail">
        {selected || creating ? (
          <>
            <div className="rbac-page__panel-header">
              <div>
                <h3>{creating ? "Create health context" : "Context metadata"}</h3>
                <p>Parent relationships are cycle-checked by the backend.</p>
              </div>
              {!creating && selected ? (
                <Button
                  hasIconOnly
                  kind="danger--ghost"
                  renderIcon={TrashCan}
                  iconDescription="Delete context"
                  onClick={confirmDelete}
                />
              ) : null}
            </div>
            <div className="rbac-form-grid">
              <TextInput
                id="health-context-code"
                labelText="Code"
                value={draft.code}
                onChange={(event) => setDraft((value) => ({ ...value, code: event.target.value }))}
              />
              <TextInput
                id="health-context-name"
                labelText="Name"
                value={draft.name}
                onChange={(event) => setDraft((value) => ({ ...value, name: event.target.value }))}
              />
              <Select
                id="health-context-type"
                labelText="Type"
                value={draft.contextType}
                onChange={(event) =>
                  setDraft((value) => ({
                    ...value,
                    contextType: event.target.value as HealthContextType,
                  }))
                }
              >
                {contextTypes.map((type) => (
                  <SelectItem key={type} value={type} text={type.replaceAll("_", " ")} />
                ))}
              </Select>
              <Select
                id="health-context-parent"
                labelText="Parent"
                value={draft.parentId ?? ""}
                onChange={(event) =>
                  setDraft((value) => ({ ...value, parentId: event.target.value || null }))
                }
              >
                <SelectItem value="" text="No parent" />
                {contexts
                  .filter((context) => context.id !== selectedId)
                  .map((context) => (
                    <SelectItem key={context.id} value={context.id} text={`${context.name} (${context.code})`} />
                  ))}
              </Select>
              <TextInput
                id="health-context-source"
                labelText="Source"
                value={draft.source}
                onChange={(event) => setDraft((value) => ({ ...value, source: event.target.value }))}
              />
              <Toggle
                id="health-context-enabled"
                labelText="Enabled"
                labelA="Off"
                labelB="On"
                toggled={draft.enabled}
                onToggle={(enabled) => setDraft((value) => ({ ...value, enabled }))}
              />
              <Button
                renderIcon={Save}
                disabled={!draft.code.trim() || !draft.name.trim() || createState.isLoading || updateState.isLoading}
                onClick={() => void save()}
              >
                Save context
              </Button>
            </div>

            {!creating && selected ? (
              <>
                <section className="health-context-aliases">
                  <h4>External identifiers</h4>
                  <div className="health-context-aliases__form">
                    <TextInput id="health-context-alias-namespace" labelText="Namespace" value={aliasNamespace} onChange={(event) => setAliasNamespace(event.target.value)} />
                    <TextInput id="health-context-alias-id" labelText="External ID" value={aliasExternalId} onChange={(event) => setAliasExternalId(event.target.value)} />
                    <Button kind="tertiary" onClick={() => void saveAlias()}>Map identifier</Button>
                  </div>
                  <div className="health-context-aliases__list">
                    {(aliases.data ?? []).map((alias) => (
                      <Tag
                        key={alias.id}
                        type="blue"
                        filter
                        onClose={() => void deleteAlias({ contextId: selected.id, aliasId: alias.id })}
                      >
                        {alias.namespace}: {alias.externalId}
                      </Tag>
                    ))}
                  </div>
                </section>
                <div className="health-context-admin__assignment-grid">
                  <ContextAssignmentEditor kind="user" contexts={contexts} />
                  <ContextAssignmentEditor kind="group" contexts={contexts} />
                </div>
              </>
            ) : null}
          </>
        ) : (
          <div className="health-context-admin__empty">
            Select a health context to inspect it, or create a new one.
          </div>
        )}
      </div>
      <HealthContextDriftPanel contexts={contexts} />
    </div>
  );
}
