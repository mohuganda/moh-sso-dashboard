import { configureStore } from "@reduxjs/toolkit";

import { baseApi } from "@moh-sso/api";
import authReducer from "@/store/auth/auth.slice";
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
