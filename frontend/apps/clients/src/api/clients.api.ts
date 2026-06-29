import { API } from "@moh-sso/config";
import type { Client, CreateClientPayload, UpdateClientPayload } from "@moh-sso/types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type UpdateClientRequest = {
  id: string;
  data: UpdateClientPayload;
};

export const clientsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List clients
     * -------------------------------- */
    listClients: builder.query<Client[], void>({
      query: () => ({
        url: API.clients.list(),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Client[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((client) => ({
                type: "Client" as const,
                id: client.id,
              })),
              { type: "Client", id: "LIST" },
            ]
          : [{ type: "Client", id: "LIST" }],
    }),

    /* --------------------------------
     * Get single client
     * -------------------------------- */
    getClient: builder.query<Client, string>({
      query: (id) => ({
        url: API.clients.byId(id),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Client>) => res.data,
      providesTags: (_r, _e, id) => [{ type: "Client", id }],
    }),

    /* --------------------------------
     * Create client
     * -------------------------------- */
    createClient: builder.mutation<Client, CreateClientPayload>({
      query: (body) => ({
        url: API.clients.create(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Client>) => res.data,
      invalidatesTags: [{ type: "Client", id: "LIST" }],
    }),

    /* --------------------------------
     * Update client (general update)
     * -------------------------------- */
    updateClient: builder.mutation<Client, UpdateClientRequest>({
      query: ({ id, data }) => ({
        url: API.clients.update(id),
        method: "PUT",
        body: data,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<Client>) => res.data,
      invalidatesTags: (_r, _e, { id }) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Enable / Disable client
     * PATCH /clients/:id/toggle
     * -------------------------------- */
    toggleClient: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: `${API.clients.byId(id)}/toggle`,
        method: "PATCH",
        body: { enabled },
        credentials: "include",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, { id }) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Delete client
     * -------------------------------- */
    deleteClient: builder.mutation<void, string>({
      query: (id) => ({
        url: API.clients.delete(id),
        method: "DELETE",
        credentials: "include",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, id) => [
        { type: "Client", id },
        { type: "Client", id: "LIST" },
      ],
    }),
  }),
});

export const {
  useListClientsQuery,
  useGetClientQuery,
  useCreateClientMutation,
  useUpdateClientMutation,
  useToggleClientMutation,
  useDeleteClientMutation,
} = clientsApi;
