import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { AuthUser } from "./auth.types";

type AuthState = {
  accessToken: string | null; // kept for compatibility, but no longer required
  user: AuthUser | null;
  loading: boolean; // UI loading
  loaded: boolean; // bootstrap completed
  isAuthenticated: boolean;
};

const initialState: AuthState = {
  accessToken: null,
  user: null,
  loading: true,
  loaded: false,
  isAuthenticated: false,
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    /* ---------------------------
     * Bootstrap finished
     * --------------------------- */
    authLoaded(state) {
      state.loading = false;
      state.loaded = true;
      state.isAuthenticated = Boolean(state.user);
    },

    /* ---------------------------
     * Successful login / refresh
     *
     * Access token is optional because backend stores it
     * in an HttpOnly cookie.
     * --------------------------- */
    loginSuccess(
      state,
      action: PayloadAction<{
        user: AuthUser;
        accessToken?: string | null;
      }>,
    ) {
      state.user = action.payload.user;
      state.accessToken = action.payload.accessToken ?? null;
      state.isAuthenticated = true;
      state.loading = false;
      state.loaded = true;
    },

    /* ---------------------------
     * Token refresh only
     *
     * Kept for backward compatibility.
     * Prefer not using this with HttpOnly cookies.
     * --------------------------- */
    setAccessToken(state, action: PayloadAction<string | null>) {
      state.accessToken = action.payload;
    },

    /* ---------------------------
     * Logout
     * --------------------------- */
    logout(state) {
      state.accessToken = null;
      state.user = null;
      state.loading = false;
      state.loaded = true;
      state.isAuthenticated = false;
    },
  },
});

export const { loginSuccess, logout, setAccessToken, authLoaded } = authSlice.actions;

export default authSlice.reducer;
