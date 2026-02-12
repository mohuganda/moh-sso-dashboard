import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { Client, ClientsState } from "../types/client.types";

const initialState: ClientsState = {
  items: [],
  activeClientId: null,
};

const clientsSlice = createSlice({
  name: "clients",
  initialState,
  reducers: {
    setClients(state, action: PayloadAction<Client[]>) {
      state.items = action.payload;
    },

    setActiveClient(state, action: PayloadAction<string | null>) {
      state.activeClientId = action.payload;
    },

    clearActiveClient(state) {
      state.activeClientId = null;
    },
  },
});

export const { setClients, setActiveClient, clearActiveClient } = clientsSlice.actions;

export default clientsSlice.reducer;
