import { API } from "@moh-sso/config";
import type { ClientRole, UserClientRoleAssignment } from "@moh-sso/clients/types";
import type {
  CreateUserPayload,
  User,
} from "../types";

import { baseApi } from "@moh-sso/api";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type UpdateUserPayload = {
  id: string;
  data: Partial<CreateUserPayload>;
};

type CreateUserRequest = {
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  enabled?: boolean;
  email_verified?: boolean;
  realm_roles?: string[];
  client_roles?: Record<string, string[]>;
};

type UpdateUserRequest = Partial<Omit<CreateUserRequest, "username">> & {
  username?: string;
};

const normalizeUser = (user: User): User => {
  const enabled = user.enabled ?? user.isActive ?? false;

  return {
    ...user,
    enabled,
    isActive: user.isActive ?? enabled,
    realmRoles: user.realmRoles ?? [],
    clientRoles: user.clientRoles ?? {},
  };
};

const toCreateUserRequest = (payload: CreateUserPayload): CreateUserRequest => ({
  username: payload.username,
  email: payload.email,
  first_name: payload.firstName,
  last_name: payload.lastName,
  enabled: payload.enabled,
  email_verified: payload.emailVerified,
  realm_roles: payload.realmRoles,
  client_roles: payload.clientRoles,
});

const toUpdateUserRequest = (payload: Partial<CreateUserPayload>): UpdateUserRequest => ({
  email: payload.email,
  first_name: payload.firstName,
  last_name: payload.lastName,
  enabled: payload.enabled,
  email_verified: payload.emailVerified,
  realm_roles: payload.realmRoles,
  client_roles: payload.clientRoles,
});

export const usersApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    /* --------------------------------
     * List users
     * -------------------------------- */
    listUsers: builder.query<User[], void>({
      query: () => ({
        url: API.admin.users.list(),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User[]>) => res.data.map(normalizeUser),
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
        url: API.admin.users.byId(id),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => normalizeUser(res.data),
      providesTags: (_r, _e, id) => [{ type: "User", id }],
    }),

    /* --------------------------------
     * Create user
     * -------------------------------- */
    createUser: builder.mutation<User, CreateUserPayload>({
      query: (body) => ({
        url: API.admin.users.create(),
        method: "POST",
        body: toCreateUserRequest(body),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => normalizeUser(res.data),
      invalidatesTags: [{ type: "User", id: "LIST" }],
    }),

    /* --------------------------------
     * Update user
     * -------------------------------- */
    updateUser: builder.mutation<User, UpdateUserPayload>({
      query: ({ id, data }) => ({
        url: API.admin.users.update(id),
        method: "PUT",
        body: toUpdateUserRequest(data),
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<User>) => normalizeUser(res.data),
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
        url: API.admin.users.passwordResetEmail(id),
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
        url: API.admin.users.delete(id),
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
        url: API.admin.users.toggle(id),
        method: "PATCH",
        body: { enabled },
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
        { type: "UserClientRole", id: `LIST-${userId}` },
        { type: "User", id: userId },
        { type: "User", id: "LIST" },
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
