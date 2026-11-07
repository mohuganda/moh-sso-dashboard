import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";
import type { Client } from "@moh-sso/types";
import type { OnboardingDraft } from "../components/onboarding-draft";

type Context = {
  environment: string;
  realm: string;
  issuer: string;
  profiles: { id: string; name: string }[];
};
export type OnboardingPreview = {
  spec: OnboardingDraft;
  realm: string;
  issuer: string;
  operation: string;
  reviewToken: string;
  client: Client & {
    standardFlowEnabled: boolean;
    serviceAccountsEnabled: boolean;
    defaultClientScopes: string[];
  };
  existing?: Client;
};
type Result = { client: Client; issuer: string; message: string };
type Envelope<T> = { data: T };
const path = (suffix: string) => `${API.clients.list()}/onboarding/${suffix}`;
const onboardingApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    onboardingContext: builder.query<Context, void>({
      query: () => ({ url: path("context"), credentials: "include" }),
      transformResponse: (res: Envelope<Context>) => res.data,
    }),
    previewOnboarding: builder.mutation<OnboardingPreview, OnboardingDraft>({
      query: (spec) => ({
        url: path("preview"),
        method: "POST",
        body: { spec },
        credentials: "include",
      }),
      transformResponse: (res: Envelope<OnboardingPreview>) => res.data,
    }),
    applyOnboarding: builder.mutation<Result, { spec: OnboardingDraft; reviewToken: string }>({
      query: (body) => ({ url: path("apply"), method: "POST", body, credentials: "include" }),
      transformResponse: (res: Envelope<Result>) => res.data,
      invalidatesTags: [{ type: "Client", id: "LIST" }],
    }),
  }),
});
export const {
  useOnboardingContextQuery,
  usePreviewOnboardingMutation,
  useApplyOnboardingMutation,
} = onboardingApi;
