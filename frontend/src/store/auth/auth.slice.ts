import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { AuthUser } from "./auth.types";

type AuthState = {
  accessToken: string | null;
  user: AuthUser | null;
  loading: boolean; // UI loading
  loaded: boolean; // bootstrap completed (FINAL)
};

const initialState: AuthState = {
  accessToken: null,
  user: null,
  loading: true,
  loaded: false,
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
    },

    /* ---------------------------
     * Successful login / refresh
     * --------------------------- */
    loginSuccess(state, action: PayloadAction<{ accessToken?: string; user: AuthUser }>) {
      if (action.payload.accessToken !== undefined) {
        state.accessToken = action.payload.accessToken;
      }

      state.user = action.payload.user;
      state.loading = false;
    },
    /* ---------------------------
     * Token refresh only
     * --------------------------- */
    setAccessToken(state, action: PayloadAction<string>) {
      state.accessToken = action.payload;
      // DO NOT touch loaded here
    },

    /* ---------------------------
     * Logout (still a known state)
     * --------------------------- */
    logout(state) {
      state.accessToken = null;
      state.user = null;
      state.loading = false;
      state.loaded = true;
    },
  },
});

export const { loginSuccess, logout, setAccessToken, authLoaded } = authSlice.actions;

export default authSlice.reducer;
