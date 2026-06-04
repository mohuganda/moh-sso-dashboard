// components/clients/panels/ClientFormPanel.tsx
import {
  Stack,
  TextInput,
  Button,
  Checkbox,
  InlineLoading,
  Form,
  FormGroup,
  TextArea,
} from "@carbon/react";
import { useMemo, useState } from "react";

import { useCreateClientMutation, useUpdateClientMutation } from "@/store/api/clients.api";
import { ClientRolesPanel } from "./roles/client-roles-panel";
import { FormInlineAlert } from "@/shared/components/notifications/in-line-alerts/FormInlineAlert";
import { useToast } from "@/shared/components/notifications/toast/useToast";

export type ClientFormMode = "create" | "edit";

type ClientFormState = {
  id: string;
  clientId: string;
  name: string;
  description?: string;
  publicClient: boolean;
  enabled: boolean;
  rootUrl?: string;
  baseUrl?: string;
  redirectUrisText: string;
  webOriginsText: string;
  icon?: string;
};

type Props = {
  mode: ClientFormMode;
  initialClient?: {
    id: string;
    clientId: string;
    name: string;
    description?: string;
    publicClient: boolean;
    enabled: boolean;
    rootUrl?: string;
    baseUrl?: string;
    redirectUris?: string[];
    webOrigins?: string[];
    icon?: string;
  };
  onSuccess?: () => void;
};

export function ClientFormPanel({ mode, initialClient, onSuccess }: Props) {
  const toast = useToast();

  const [createdClientId, setCreatedClientId] = useState<string | null>(null);

  const effectiveClientId = mode === "edit" ? initialClient?.clientId : createdClientId;

  const [form, setForm] = useState<ClientFormState>({
    id: initialClient?.id ?? "",
    clientId: initialClient?.clientId ?? "",
    name: initialClient?.name ?? "",
    description: initialClient?.description ?? "",
    publicClient: initialClient?.publicClient ?? false,
    enabled: initialClient?.enabled ?? true,
    rootUrl: initialClient?.rootUrl ?? "",
    baseUrl: initialClient?.baseUrl ?? "",
    redirectUrisText: (initialClient?.redirectUris ?? []).join("\n"),
    webOriginsText: (initialClient?.webOrigins ?? []).join("\n"),
    icon: initialClient?.icon ?? "",
  });

  const [error, setError] = useState<string | null>(null);

  const [createClient, { isLoading: creating }] = useCreateClientMutation();
  const [updateClient, { isLoading: updating }] = useUpdateClientMutation();

  const submitting = creating || updating;

  /* -----------------------------
   * Helpers
   * ----------------------------- */
  const normalizeList = (value: string): string[] =>
    value
      .split("\n")
      .map((v) => v.trim())
      .filter(Boolean);

  const buildPayload = () => ({
    name: form.name.trim(),
    description: form.description?.trim() || undefined,
    icon: form.icon?.trim() || undefined,
    publicClient: form.publicClient,
    enabled: form.enabled,
    rootUrl: form.rootUrl?.trim() || undefined,
    baseUrl: form.baseUrl?.trim() || undefined,
    redirectUris: normalizeList(form.redirectUrisText),
    webOrigins: normalizeList(form.webOriginsText),
  });

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isClientIdValid = /^[a-z0-9-]+$/.test(form.clientId);

  const isValid = useMemo(() => {
    if (!form.name.trim()) return false;
    if (mode === "create") {
      if (!form.clientId.trim()) return false;
      if (!isClientIdValid) return false;
    }
    return true;
  }, [form, mode, isClientIdValid]);

  const handleChange = <K extends keyof ClientFormState>(field: K, value: ClientFormState[K]) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  /* -----------------------------
   * Submit
   * ----------------------------- */
  const handleSubmit = async () => {
    if (!isValid || submitting) return;

    setError(null);

    try {
      if (mode === "create") {
        const created = await createClient({
          clientId: form.clientId.toLowerCase().trim(),
          ...buildPayload(),
        }).unwrap();

        setCreatedClientId(created.clientId);

        toast.success("Client created", `Client "${form.name}" was created`);
      } else {
        await updateClient({
          id: form.clientId,
          data: buildPayload(),
        }).unwrap();

        toast.success("Client updated", `Changes to "${form.name}" saved`);
      }

      onSuccess?.();
    } catch (err: any) {
      const message =
        err?.data?.message ??
        (mode === "create" ? "Failed to create client" : "Failed to update client");

      setError(message);
      toast.error("Operation failed", "Please review the form and try again");
    }
  };

  return (
    <Form>
      <Stack gap={7}>
        {error && <FormInlineAlert title="Unable to save client" subtitle={error} />}

        {/* -----------------------------
         * Client details
         * ----------------------------- */}
        <FormGroup legendText="Client details">
          <Stack gap={4}>
            <TextInput
              id="clientId"
              labelText="Client ID"
              helperText={
                mode === "create"
                  ? "Lowercase letters, numbers, and dashes only"
                  : "Client ID cannot be changed"
              }
              required
              disabled={mode === "edit"}
              value={form.clientId}
              invalid={mode === "create" && form.clientId.length > 0 && !isClientIdValid}
              invalidText="Only lowercase letters, numbers, and dashes allowed"
              onChange={(e) => {
                handleChange("clientId", e.target.value.toLowerCase());
              }}
            />

            <TextInput
              id="name"
              labelText="Client name"
              required
              value={form.name}
              onChange={(e) => {
                handleChange("name", e.target.value);
              }}
            />

            <TextArea
              id="description"
              labelText="Description"
              value={form.description}
              onChange={(e) => {
                handleChange("description", e.target.value);
              }}
            />

            <TextInput
              id="icon"
              labelText="Icon (optional)"
              value={form.icon}
              onChange={(e) => {
                handleChange("icon", e.target.value);
              }}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * URLs & Redirects
         * ----------------------------- */}
        <FormGroup legendText="URLs & redirects">
          <Stack gap={4}>
            <TextInput
              id="rootUrl"
              labelText="Root URL"
              value={form.rootUrl}
              onChange={(e) => {
                handleChange("rootUrl", e.target.value);
              }}
            />

            <TextInput
              id="baseUrl"
              labelText="Base URL"
              value={form.baseUrl}
              onChange={(e) => {
                handleChange("baseUrl", e.target.value);
              }}
            />

            <TextArea
              id="redirectUris"
              labelText="Redirect URIs"
              helperText="One per line"
              value={form.redirectUrisText}
              onChange={(e) => {
                handleChange("redirectUrisText", e.target.value);
              }}
            />

            <TextArea
              id="webOrigins"
              labelText="Web origins"
              helperText="One per line"
              value={form.webOriginsText}
              onChange={(e) => {
                handleChange("webOriginsText", e.target.value);
              }}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Access & Security
         * ----------------------------- */}
        <FormGroup legendText="Access & security">
          <Stack gap={4}>
            <Checkbox
              id="publicClient"
              labelText="Public client (no client secret)"
              checked={form.publicClient}
              onChange={(_, { checked }) => {
                handleChange("publicClient", checked);
              }}
            />

            <Checkbox
              id="enabled"
              labelText="Client enabled"
              checked={form.enabled}
              onChange={(_, { checked }) => {
                handleChange("enabled", checked);
              }}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Actions
         * ----------------------------- */}
        <Stack orientation="horizontal" gap={3}>
          <Button kind="primary" disabled={!isValid || submitting} onClick={handleSubmit}>
            {submitting ? (
              <InlineLoading
                description={mode === "create" ? "Creating client…" : "Updating client…"}
              />
            ) : mode === "create" ? (
              "Create client"
            ) : (
              "Save changes"
            )}
          </Button>
        </Stack>

        {/* -----------------------------
         * Client roles (post-create / edit)
         * ----------------------------- */}
        {effectiveClientId && (
          <FormGroup legendText="Client roles">
            <ClientRolesPanel clientId={effectiveClientId} id={form.id} />
          </FormGroup>
        )}
      </Stack>
    </Form>
  );
}
