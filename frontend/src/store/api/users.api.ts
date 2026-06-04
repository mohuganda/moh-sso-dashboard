import { API } from "@/lib/constants/api.constants";
import type { ClientRole } from "@/store/types/client-role.types";
import type { UserClientRoleAssignment } from "@/store/types/user-role.types";
import type { CreateUserPayload, User } from "@/store/types/user.types";

import { baseApi } from "./baseApi";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type UpdateUserPayload = {
  id: string;
  data: Partial<CreateUserPayload>;
};

export const usersApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List users
     * -------------------------------- */
    listUsers: builder.query<User[], void>({
      query: () => ({
        url: API.users.list(),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User[]>) => res.data,
      providesTags: (result) =>
        result
          ? [
              ...result.map((u) => ({ type: "User" as const, id: u.id })),
              { type: "User", id: "LIST" },
            ]
          : [{ type: "User", id: "LIST" }],
    }),

    /* --------------------------------
     * Get single user
     * -------------------------------- */
    getUser: builder.query<User, string>({
      query: (id) => ({
        url: API.users.byId(id),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => res.data,
      providesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    /* --------------------------------
     * Create user
     * -------------------------------- */
    createUser: builder.mutation<User, CreateUserPayload>({
      query: (body) => ({
        url: API.users.create(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => res.data,
      invalidatesTags: [{ type: "User", id: "LIST" }],
    }),

    /* --------------------------------
     * Update user
     * -------------------------------- */
    updateUser: builder.mutation<User, UpdateUserPayload>({
      query: ({ id, data }) => ({
        url: API.users.update(id),
        method: "PATCH",
        body: data,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => res.data,
      invalidatesTags: (_r, _e, { id }) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Reset password
     * -------------------------------- */
    resetUserPassword: builder.mutation<void, string>({
      query: (id) => ({
        url: API.users.resetPassword(id),
        method: "POST",
        credentials: "include",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    /* --------------------------------
     * Delete user
     * -------------------------------- */
    deleteUser: builder.mutation<void, string>({
      query: (id) => ({
        url: API.users.delete(id),
        method: "DELETE",
        credentials: "include",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, id) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Enable / Disable user
     * -------------------------------- */
    toggleUser: builder.mutation<void, { id: string; enabled: boolean }>({
      query: ({ id, enabled }) => ({
        url: `${API.users.byId(id)}/toggle`,
        method: enabled ? "POST" : "DELETE",
        credentials: "include",
      }),
      transformResponse: () => undefined,
      invalidatesTags: (_r, _e, { id }) => [
        { type: "User", id },
        { type: "User", id: "LIST" },
      ],
    }),

    /* --------------------------------
     * Get user client roles (ADMIN)
     * GET /admin/users/:id/client-roles
     * -------------------------------- */
    getUserClientRoles: builder.query<UserClientRoleAssignment[], string>({
      query: (userId) => ({
        url: `${API.admin.base}/users/${userId}/client-roles`,
        credentials: "include",
      }),

      transformResponse: (res: ApiEnvelope<UserClientRoleAssignment[]>) => res.data,

      providesTags: (result, _, userId) =>
        result
          ? [
              ...result.map((r) => ({
                type: "UserClientRole" as const,
                id: `${userId}-${r.id}`,
              })),
              { type: "UserClientRole", id: `LIST-${userId}` },
            ]
          : [{ type: "UserClientRole", id: `LIST-${userId}` }],
    }),

    /* --------------------------------
     * Get user client roles for a client (ADMIN)
     * -------------------------------- */
    getUserClientRolesForClient: builder.query<ClientRole[], { userId: string; clientId: string }>({
      query: ({ userId, clientId }) => ({
        url: API.admin.users.clientRoles.list(userId, clientId),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<ClientRole[]>) => res.data,
      providesTags: (_r, _e, { userId, clientId }) => [
        { type: "UserClientRole", id: `${userId}-${clientId}` },
      ],
    }),

    /* --------------------------------
     * Update user client roles (ADMIN – bulk)
     * -------------------------------- */
    updateUserClientRoles: builder.mutation<
      void,
      { userId: string; clientId: string; clientUuid: string; roles: string[] }
    >({
      query: ({ userId, clientId, clientUuid, roles }) => ({
        url: API.admin.users.clientRoles.update(userId),
        method: "PUT",
        body: {
          clientId,
          clientUuid,
          roles,
        },
        credentials: "include",
      }),

      invalidatesTags: (_r, _e, { userId, clientUuid }) => [
        {
          type: "UserClientRole",
          id: `${userId}-${clientUuid}`,
        },
      ],
    }),
  }),
});

export const {
  useListUsersQuery,
  useGetUserQuery,
  useCreateUserMutation,
  useUpdateUserMutation,
  useResetUserPasswordMutation,
  useDeleteUserMutation,
  useToggleUserMutation,
  useGetUserClientRolesQuery,
  useGetUserClientRolesForClientQuery,
  useUpdateUserClientRolesMutation,
} = usersApi;
