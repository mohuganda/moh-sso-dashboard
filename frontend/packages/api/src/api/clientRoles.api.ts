import { API } from "@moh-sso/config";
import type { ClientRole } from "@moh-sso/types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export type RoleRequest = {
  role: string;
  description?: string;
};

export const clientRolesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List roles for a client (AUTH)
     * GET /clients/:id/roles
     * -------------------------------- */
    listClientRoles: builder.query<ClientRole[], string>({
      query: (clientId) => ({
        url: API.clients.roles.list(clientId),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<ClientRole[]>) => res.data,

      providesTags: (result, _, clientId) =>
        result
          ? [
              ...result.map((role) => ({
                type: "ClientRole" as const,
                id: `${clientId}-${role.name}`,
              })),
              { type: "ClientRole", id: `LIST-${clientId}` },
            ]
          : [{ type: "ClientRole", id: `LIST-${clientId}` }],
    }),

    /* --------------------------------
     * Create client role (ADMIN)
     * POST /admin/clients/:id/roles
     * -------------------------------- */
    createClientRole: builder.mutation<void, { clientId: string; payload: RoleRequest }>({
      query: ({ clientId, payload }) => ({
        url: API.clients.roles.create(clientId),
        method: "POST",
        body: payload,
        credentials: "include",
      }),

      invalidatesTags: (_r, _e, { clientId }) => [{ type: "ClientRole", id: `LIST-${clientId}` }],
    }),

    /* --------------------------------
     * Delete client role (ADMIN)
     * DELETE /admin/clients/:id/roles/:role
     * -------------------------------- */
    deleteClientRole: builder.mutation<void, { clientId: string; role: string }>({
      query: ({ clientId, role }) => ({
        url: API.clients.roles.delete(clientId, role),
        method: "DELETE",
        credentials: "include",
      }),

      invalidatesTags: (_r, _e, { clientId, role }) => [
        { type: "ClientRole", id: `${clientId}-${role}` },
        { type: "ClientRole", id: `LIST-${clientId}` },
      ],
    }),
  }),
});

export const { useListClientRolesQuery, useCreateClientRoleMutation, useDeleteClientRoleMutation } =
  clientRolesApi;
