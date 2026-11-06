import { useState } from "react";
import {
  Button,
  Checkbox,
  FormGroup,
  InlineLoading,
  Select,
  SelectItem,
  Stack,
  TextArea,
  TextInput,
  Tag,
} from "@carbon/react";
import type { Client } from "@moh-sso/types";
import { FormInlineAlert, useHeaderPanel, useToast } from "@moh-sso/ui";
import {
  useApplyOnboardingMutation,
  useOnboardingContextQuery,
  usePreviewOnboardingMutation,
} from "../api/onboarding.api";
import type { OnboardingPreview } from "../api/onboarding.api";
import { emptyDraft, exportOnboardingDraft, parseOnboardingDraft } from "./onboarding-draft";
import type { OnboardingDraft } from "./onboarding-draft";
import { ClientRolesPanel } from "./roles/client-roles-panel";
import "./client-onboarding-panel.scss";

function message(error: unknown): string {
  if (typeof error === "object" && error && "data" in error) {
    const data = error.data as { error?: { message?: string } };
    if (data?.error?.message) return data.error.message;
  }
  return error instanceof Error ? error.message : "Unable to complete this step. Please try again.";
}

function initialDraft(client?: Client): OnboardingDraft {
  const attrs = client?.attributes;
  return client
    ? {
        ...emptyDraft(attrs?.["portal.onboarding.environment"] ?? ""),
        clientId: client.clientId,
        name: client.name,
        description: client.description ?? "",
        profile: attrs?.["portal.onboarding.profile"] ?? "public-web",
        owner: attrs?.["portal.owner"] ?? "",
        supportEmail: attrs?.["portal.supportEmail"] ?? "",
        rootUrl: client.rootUrl ?? "",
        redirectUris: client.redirectUris,
        webOrigins: client.webOrigins ?? [],
        logoutUris: (attrs?.["post.logout.redirect.uris"] ?? "").split("##").filter(Boolean),
        launcher: attrs?.["portal.system"] === "true",
      }
    : emptyDraft("");
}

export function ClientOnboardingPanel({
  initialClient,
  onSuccess,
}: {
  initialClient?: Client;
  onSuccess?: () => void;
}) {
  const { data: context, error: contextError, isLoading } = useOnboardingContextQuery();
  const [draft, setDraft] = useState(() => initialDraft(initialClient));
  const [step, setStep] = useState(0);
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<OnboardingPreview | null>(null);
  const [result, setResult] = useState<{ client: Client; issuer: string; message: string } | null>(
    null,
  );
  const [reviewed, setReviewed] = useState(false);
  const [previewDraft, { isLoading: previewing }] = usePreviewOnboardingMutation();
  const [apply, { isLoading: applying }] = useApplyOnboardingMutation();
  const toast = useToast();
  const { closePanel } = useHeaderPanel();
  const environment = draft.environment || context?.environment || "";
  const spec = { ...draft, environment };
  const change = <K extends keyof OnboardingDraft>(key: K, value: OnboardingDraft[K]) => {
    setDraft((current) => ({ ...current, [key]: value }));
    setPreview(null);
    setReviewed(false);
    setError("");
  };
  const download = () => {
    const blob = new Blob([exportOnboardingDraft(spec)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${spec.clientId || "application"}-onboarding.json`;
    link.click();
    URL.revokeObjectURL(url);
  };
  const review = async () => {
    try {
      setError("");
      setPreview(await previewDraft(spec).unwrap());
      setReviewed(false);
      setStep(2);
    } catch (failure) {
      setError(message(failure));
    }
  };
  const provision = async () => {
    if (!preview || !reviewed) return;
    try {
      setError("");
      const saved = await apply({ spec, reviewToken: preview.reviewToken }).unwrap();
      setResult(saved);
      toast.success(
        "Client provisioned",
        "Configuration verified. Complete integration checks before enabling.",
      );
    } catch (failure) {
      setError(message(failure));
      setPreview(null);
      setReviewed(false);
      setStep(1);
    }
  };
  if (isLoading) return <InlineLoading description="Loading this environment…" />;
  if (contextError || !context)
    return (
      <FormInlineAlert title="Client onboarding unavailable" subtitle={message(contextError)} />
    );
  if (result)
    return (
      <Stack gap={5}>
        <h3>Application configured</h3>
        <Tag type="gray">Disabled pending integration checks</Tag>
        <p>{result.message}</p>
        <dl>
          <dt>Client ID</dt>
          <dd>{result.client.clientId}</dd>
          <dt>Issuer</dt>
          <dd>{result.issuer}</dd>
          <dt>Environment</dt>
          <dd>
            {context.environment} · {context.realm}
          </dd>
        </dl>
        <p>
          {spec.profile === "public-web"
            ? "This public client does not use a secret."
            : "Retrieve or rotate this client's secret privately through this realm's Keycloak administration. Secrets are never included in onboarding drafts."}
        </p>
        <p>
          Test sign-in, token validation and logout. Configure portal access roles in RBAC if this
          application has a launcher entry, then enable the client from Clients.
        </p>
        <details>
          <summary>Configure client roles (optional)</summary>
          <ClientRolesPanel clientUuid={result.client.id} clientId={result.client.clientId} />
        </details>
        <Stack orientation="horizontal" gap={3}>
          <Button kind="secondary" onClick={download}>
            Export onboarding draft
          </Button>
          <Button onClick={onSuccess ?? closePanel}>Done</Button>
        </Stack>
      </Stack>
    );
  return (
    <div className="client-onboarding">
      <Stack gap={5}>
        <div>
          <Tag type={context.environment === "production" ? "red" : "blue"}>
            {context.environment}
          </Tag>
          <span>{context.realm}</span>
        </div>
        <ol className="client-onboarding__steps" aria-label="Onboarding progress">
          {["Application", "URLs & access", "Review"].map((label, index) => (
            <li key={label} aria-current={step === index ? "step" : undefined}>
              {index + 1}. {label}
            </li>
          ))}
        </ol>
        {error && <FormInlineAlert title="Review required" subtitle={error} />}
        {initialClient?.enabled && (
          <FormInlineAlert
            title="Disable this client first"
            subtitle="Active integrations cannot be changed through onboarding. Disable the client from Clients before reviewing an update."
          />
        )}
        {step === 0 && (
          <>
            <FormGroup legendText="Register an application">
              <Stack gap={4}>
                <label htmlFor="onboarding-import">Import a draft from another environment</label>
                <input
                  id="onboarding-import"
                  type="file"
                  accept=".json,application/json"
                  onChange={async (event) => {
                    const file = event.target.files?.[0];
                    if (!file) return;
                    try {
                      if (file.size > 65536) throw new Error("Choose a draft smaller than 64 KB.");
                      setDraft(parseOnboardingDraft(await file.text()));
                      setPreview(null);
                      setReviewed(false);
                      setError("");
                    } catch (failure) {
                      setError(message(failure));
                    }
                    event.target.value = "";
                  }}
                />
                <TextInput
                  id="onboarding-client-id"
                  labelText="Client ID"
                  helperText="Use the same ID across environments."
                  value={draft.clientId}
                  disabled={Boolean(initialClient)}
                  onChange={(event) => change("clientId", event.target.value)}
                />
                <TextInput
                  id="onboarding-name"
                  labelText="Application name"
                  value={draft.name}
                  onChange={(event) => change("name", event.target.value)}
                />
                <TextArea
                  id="onboarding-description"
                  labelText="Description"
                  value={draft.description}
                  onChange={(event) => change("description", event.target.value)}
                />
                <TextInput
                  id="onboarding-owner"
                  labelText="Application owner or team"
                  value={draft.owner}
                  onChange={(event) => change("owner", event.target.value)}
                />
                <TextInput
                  id="onboarding-support"
                  type="email"
                  labelText="Support email"
                  value={draft.supportEmail}
                  onChange={(event) => change("supportEmail", event.target.value)}
                />
                <Select
                  id="onboarding-profile"
                  labelText="Client profile"
                  value={draft.profile}
                  disabled={Boolean(initialClient)}
                  onChange={(event) => {
                    const profile = event.target.value;
                    change("profile", profile);
                    if (profile === "service")
                      setDraft((current) => ({
                        ...current,
                        profile,
                        rootUrl: "",
                        redirectUris: [],
                        logoutUris: [],
                        webOrigins: [],
                        launcher: false,
                      }));
                  }}
                >
                  {context.profiles.map((profile) => (
                    <SelectItem key={profile.id} value={profile.id} text={profile.name} />
                  ))}
                </Select>
              </Stack>
            </FormGroup>
            <Button
              disabled={!draft.clientId || !draft.name || !draft.owner || !draft.supportEmail}
              onClick={() => setStep(1)}
            >
              Continue
            </Button>
          </>
        )}
        {step === 1 && (
          <>
            {environment !== context.environment ? (
              <FormInlineAlert
                title={`Draft belongs to ${environment}`}
                subtitle={`Review every URL for ${context.environment}, then select “Use this environment” before previewing.`}
              />
            ) : (
              <p>
                Review all addresses for <strong>{context.environment}</strong>. Callbacks and
                logout URLs must be exact. HTTPS is required outside development/test loopback.
              </p>
            )}
            {environment !== context.environment && (
              <Button kind="secondary" onClick={() => change("environment", context.environment)}>
                Use this environment
              </Button>
            )}
            {draft.profile !== "service" && (
              <Stack gap={4}>
                <TextInput
                  id="onboarding-root"
                  labelText="Application URL"
                  value={draft.rootUrl}
                  onChange={(event) => change("rootUrl", event.target.value)}
                />
                {(
                  [
                    ["redirectUris", "Sign-in callback URLs"],
                    ["logoutUris", "Post-logout URLs"],
                    ["webOrigins", "Allowed browser origins"],
                  ] as const
                ).map(([key, label]) => (
                  <TextArea
                    key={key}
                    id={`onboarding-${key}`}
                    labelText={label}
                    helperText="One exact address per line. No wildcards."
                    value={draft[key].join("\n")}
                    onChange={(event) =>
                      change(
                        key,
                        event.target.value
                          .split("\n")
                          .map((value) => value.trim())
                          .filter(Boolean),
                      )
                    }
                  />
                ))}
                <Checkbox
                  id="onboarding-launcher"
                  labelText="Register an external portal launcher entry"
                  checked={draft.launcher}
                  onChange={(_, { checked }) => change("launcher", checked)}
                />
                <p>
                  Launcher registration does not grant access. Configure access roles, approve and
                  activate the portal application separately in RBAC.
                </p>
              </Stack>
            )}
            {draft.profile === "service" && (
              <p>
                This profile enables client credentials only. It has no browser redirects, user
                scopes or automatic application-role grants.
              </p>
            )}
            <Stack orientation="horizontal" gap={3}>
              <Button kind="secondary" onClick={() => setStep(0)}>
                Back
              </Button>
              <Button
                disabled={
                  previewing ||
                  applying ||
                  environment !== context.environment ||
                  initialClient?.enabled
                }
                onClick={review}
              >
                {previewing ? "Validating…" : "Validate & preview"}
              </Button>
            </Stack>
          </>
        )}
        {step === 2 && preview && (
          <>
            <h3>
              {preview.operation === "update"
                ? "Review configuration changes"
                : "Review new client"}
            </h3>
            <dl>
              <dt>Target</dt>
              <dd>
                {context.environment} · {preview.realm}
              </dd>
              <dt>Issuer</dt>
              <dd>{preview.issuer}</dd>
              <dt>Client ID</dt>
              <dd>{draft.clientId}</dd>
              <dt>Profile</dt>
              <dd>{draft.profile}</dd>
              <dt>Sign-in method</dt>
              <dd>
                {draft.profile === "service"
                  ? "Client credentials"
                  : "Authorization code with S256 PKCE"}
              </dd>
              <dt>Application owner</dt>
              <dd>
                {draft.owner} · {draft.supportEmail}
              </dd>
              {draft.profile !== "service" && (
                <>
                  <dt>Sign-in callbacks</dt>
                  <dd>
                    {draft.redirectUris.map((url) => (
                      <div key={url}>{url}</div>
                    ))}
                  </dd>
                  <dt>Post-logout addresses</dt>
                  <dd>
                    {draft.logoutUris.map((url) => (
                      <div key={url}>{url}</div>
                    ))}
                  </dd>
                  <dt>Browser origins</dt>
                  <dd>
                    {draft.webOrigins.map((url) => (
                      <div key={url}>{url}</div>
                    ))}
                  </dd>
                </>
              )}
              <dt>Portal launcher</dt>
              <dd>
                {draft.launcher
                  ? "Register for separate access mapping and activation"
                  : "Authentication only"}
              </dd>
              <dt>State after provisioning</dt>
              <dd>Disabled</dd>
            </dl>
            {preview.existing && (
              <details>
                <summary>Current configuration</summary>
                <pre>{JSON.stringify(preview.existing, null, 2)}</pre>
              </details>
            )}
            <details>
              <summary>Advanced configuration to apply</summary>
              <pre>{JSON.stringify(preview.client, null, 2)}</pre>
            </details>
            <Checkbox
              id="onboarding-reviewed"
              labelText={`I reviewed these URLs and settings for ${context.environment}.`}
              checked={reviewed}
              onChange={(_, { checked }) => setReviewed(checked)}
            />
            <Stack orientation="horizontal" gap={3}>
              <Button kind="secondary" disabled={applying} onClick={() => setStep(1)}>
                Back
              </Button>
              <Button disabled={!reviewed || applying} onClick={provision}>
                {applying ? "Provisioning…" : "Provision disabled client"}
              </Button>
            </Stack>
          </>
        )}
        {!result && (
          <Button kind="ghost" disabled={applying} onClick={download}>
            Export onboarding draft
          </Button>
        )}
      </Stack>
    </div>
  );
}
