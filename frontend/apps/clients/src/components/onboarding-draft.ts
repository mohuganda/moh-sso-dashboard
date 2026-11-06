export type OnboardingDraft = {
  version: number;
  clientId: string;
  name: string;
  description: string;
  owner: string;
  supportEmail: string;
  profile: string;
  environment: string;
  rootUrl: string;
  redirectUris: string[];
  logoutUris: string[];
  webOrigins: string[];
  launcher: boolean;
};

export const emptyDraft = (environment: string): OnboardingDraft => ({
  version: 1,
  clientId: "",
  name: "",
  description: "",
  owner: "",
  supportEmail: "",
  profile: "public-web",
  environment,
  rootUrl: "",
  redirectUris: [],
  logoutUris: [],
  webOrigins: [],
  launcher: false,
});

// Promotion documents deliberately contain configuration only. Reject unknown keys
// rather than accidentally retaining credentials from a full Keycloak export.
export function parseOnboardingDraft(text: string): OnboardingDraft {
  const value: unknown = JSON.parse(text);
  if (!value || typeof value !== "object" || Array.isArray(value))
    throw new Error("Choose an onboarding draft, not a realm export.");
  const data = value as Record<string, unknown>;
  const fields = emptyDraft("");
  if (Object.keys(data).some((key) => !Object.hasOwn(fields, key)) || data.version !== 1)
    throw new Error(
      "Unsupported document. Import a version 1 onboarding draft without credentials.",
    );
  for (const key of [
    "clientId",
    "name",
    "description",
    "owner",
    "supportEmail",
    "profile",
    "environment",
    "rootUrl",
  ] as const) {
    if (typeof data[key] !== "string") throw new Error(`Invalid ${key} in onboarding draft.`);
  }
  for (const key of ["redirectUris", "logoutUris", "webOrigins"] as const) {
    if (!Array.isArray(data[key]) || !data[key].every((item: unknown) => typeof item === "string"))
      throw new Error(`Invalid ${key} in onboarding draft.`);
  }
  if (typeof data.launcher !== "boolean") throw new Error("Invalid launcher setting.");
  return data as OnboardingDraft;
}

export function exportOnboardingDraft(draft: OnboardingDraft): string {
  // Parsing again enforces the same credential-free contract on export.
  return JSON.stringify(parseOnboardingDraft(JSON.stringify(draft)), null, 2);
}
