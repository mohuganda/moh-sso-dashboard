import { createSelector } from "@reduxjs/toolkit";

import type { RootState } from "..";

/* -----------------------------
 * Base selectors
 * ----------------------------- */

export const selectClientsState = (state: RootState) => state.clients;

/**
 * Clients visible to the UI are populated from the authenticated user's
 * RBAC-resolved accessible systems.
 */
export const selectClients = createSelector(selectClientsState, (s) => s.items);

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
