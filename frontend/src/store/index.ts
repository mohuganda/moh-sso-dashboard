import { configureStore } from "@reduxjs/toolkit";

import { baseApi } from "./api/baseApi";
import authReducer from "./auth/auth.slice";
import clientsReducer from "./clients/clients.slice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    clients: clientsReducer,
    [baseApi.reducerPath]: baseApi.reducer,
  },
  middleware: (getDefault) => getDefault().concat(baseApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
