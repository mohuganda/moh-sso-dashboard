import { createSelector } from "@reduxjs/toolkit";

import type { RootState } from "../index";

export const selectAuth = (s: RootState) => s.auth;

export const selectAccessToken = createSelector(selectAuth, (a) => a.accessToken);

export const selectUser = createSelector(selectAuth, (a) => a.user);

export const selectAuthenticated = createSelector(selectAccessToken, (t) => !!t);

export const selectIsAdmin = createSelector(selectUser, (u) => u?.isAdmin === true);

export const selectIsUser = createSelector(selectUser, (u) => u?.isUser === true);

export const selectAuthLoading = createSelector(selectAuth, (a) => a.loading);

export const selectAuthLoaded = createSelector(selectAuth, (a) => a.loaded);
