import { API } from "@moh-sso/config";
import type { ClientRole } from "@moh-sso/types";

import { baseApi } from "@moh-sso/api";

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
      query: (clientUuid) => ({
        url: API.clients.roles.list(clientUuid),
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<ClientRole[]>) => res.data,

      providesTags: (result, _, clientUuid) =>
        result
          ? [
              ...result.map((role) => ({
                type: "ClientRole" as const,
                id: `${clientUuid}-${role.name}`,
              })),
              { type: "ClientRole", id: `LIST-${clientUuid}` },
            ]
          : [{ type: "ClientRole", id: `LIST-${clientUuid}` }],
    }),

    /* --------------------------------
     * Create client role (ADMIN)
     * POST /admin/clients/:id/roles
     * -------------------------------- */
    createClientRole: builder.mutation<void, { clientUuid: string; payload: RoleRequest }>({
      query: ({ clientUuid, payload }) => ({
        url: API.clients.roles.create(clientUuid),
        method: "POST",
        body: payload,
        credentials: "include",
      }),

      invalidatesTags: (_r, _e, { clientUuid }) => [
        { type: "ClientRole", id: `LIST-${clientUuid}` },
      ],
    }),

    /* --------------------------------
     * Delete client role (ADMIN)
     * DELETE /admin/clients/:id/roles/:role
     * -------------------------------- */
    deleteClientRole: builder.mutation<void, { clientUuid: string; roleName: string }>({
      query: ({ clientUuid, roleName }) => ({
        url: API.clients.roles.delete(clientUuid, roleName),
        method: "DELETE",
        credentials: "include",
      }),

      invalidatesTags: (_r, _e, { clientUuid, roleName }) => [
        { type: "ClientRole", id: `${clientUuid}-${roleName}` },
        { type: "ClientRole", id: `LIST-${clientUuid}` },
      ],
    }),
  }),
});

export const { useListClientRolesQuery, useCreateClientRoleMutation, useDeleteClientRoleMutation } =
  clientRolesApi;
