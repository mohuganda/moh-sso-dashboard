import { createSelector } from "@reduxjs/toolkit";

import type { RootState } from "@moh-sso/state";

export const selectAuth = (s: RootState) => s.auth;

export const selectAccessToken = createSelector(selectAuth, (auth) => auth.accessToken);

export const selectUser = createSelector(selectAuth, (auth) => auth.user);

export const selectAuthenticated = createSelector(selectAuth, (auth) => auth.isAuthenticated);

export const selectIsAdmin = createSelector(selectUser, (user) => user?.isAdmin === true);

export const selectIsUser = createSelector(selectUser, (user) => user?.isUser === true);

export const selectPermissions = createSelector(selectUser, (user) => user?.permissions ?? []);

export const selectHasPermission = (permission: string) =>
  createSelector(selectPermissions, (permissions) => permissions.includes(permission as never));

export const selectHasAnyPermission = (requiredPermissions: string[]) =>
  createSelector(selectPermissions, (permissions) =>
    requiredPermissions.some((permission) => permissions.includes(permission as never)),
  );

export const selectAuthLoading = createSelector(selectAuth, (auth) => auth.loading);

export const selectAuthLoaded = createSelector(selectAuth, (auth) => auth.loaded);
