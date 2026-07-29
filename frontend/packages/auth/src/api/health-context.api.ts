import { API } from "@moh-sso/config";
import { baseApi } from "@moh-sso/api";

import type { EffectiveHealthContext } from "../auth/auth.types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const healthContextApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    myHealthContexts: builder.query<EffectiveHealthContext[], void>({
      query: () => ({
        url: API.healthContexts.mine(),
        credentials: "include",
      }),
      transformResponse: (
        response: ApiEnvelope<{ healthContexts: EffectiveHealthContext[] | null }>,
      ) => response.data.healthContexts ?? [],
      providesTags: ["HealthContext"],
    }),
    selectHealthContext: builder.mutation<{ contextId: string }, string>({
      query: (contextId) => ({
        url: API.healthContexts.select(),
        method: "POST",
        credentials: "include",
        body: { contextId },
      }),
      transformResponse: (response: ApiEnvelope<{ contextId: string }>) => response.data,
      invalidatesTags: ["HealthContext"],
    }),
  }),
});

export const { useMyHealthContextsQuery, useSelectHealthContextMutation } = healthContextApi;
