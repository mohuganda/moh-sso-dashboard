import { createSelector } from "@reduxjs/toolkit";

import type { RootState } from "../index";

export const selectAuth = (s: RootState) => s.auth;

export const selectAccessToken = createSelector(selectAuth, (auth) => auth.accessToken);

export const selectUser = createSelector(selectAuth, (auth) => auth.user);

export const selectAuthenticated = createSelector(selectAuth, (auth) => auth.isAuthenticated);

export const selectIsAdmin = createSelector(selectUser, (user) => user?.isAdmin === true);

export const selectIsUser = createSelector(selectUser, (user) => user?.isUser === true);

export const selectAuthLoading = createSelector(selectAuth, (auth) => auth.loading);

export const selectAuthLoaded = createSelector(selectAuth, (auth) => auth.loaded);
