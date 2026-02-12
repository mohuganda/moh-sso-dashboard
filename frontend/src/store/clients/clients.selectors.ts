import { createSelector } from "@reduxjs/toolkit";

import type { RootState } from "..";

import { DEFAULT_CLIENTS } from "./defaultClient";

/* -----------------------------
 * Base selectors
 * ----------------------------- */

export const selectClientsState = (state: RootState) => state.clients;

/**
 * All clients visible to the UI:
 * - System defaults (Utilities, Settings, etc.)
 * - Keycloak-managed clients
 */
export const selectClients = createSelector(selectClientsState, (s) => [
  ...DEFAULT_CLIENTS,
  ...s.items,
]);

export const selectActiveClientId = createSelector(selectClientsState, (s) => s.activeClientId);

/**
 * Resolve active client safely
 * Falls back to first system client if missing
 */
export const selectActiveClient = createSelector(
  [selectClients, selectActiveClientId],
  (clients, activeId) => clients.find((c) => c.clientId === activeId) ?? clients[0] ?? null,
);

/* -----------------------------
 * Attribute helpers
 * ----------------------------- */

const parseJSON = <T>(value?: string): T | null => {
  if (!value) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
};

/* -----------------------------
 * UI selectors
 * ----------------------------- */

export type SideNavItem = {
  id: string;
  label: string;
  path?: string;
  icon?: string;
  permission?: string;
  children?: SideNavItem[];
};

export const selectClientSideNav = createSelector(selectActiveClient, (client): SideNavItem[] => {
  const raw = client?.attributes?.["ui.sidenav"];
  return parseJSON<SideNavItem[]>(raw) ?? [];
});

export const selectClientIcon = createSelector(
  selectActiveClient,
  (client) => client?.attributes?.["ui.icon"] ?? "app",
);

export const selectClientHome = createSelector(
  selectActiveClient,
  (client) => client?.attributes?.["ui.home"] ?? "/",
);
