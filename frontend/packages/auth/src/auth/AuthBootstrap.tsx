import { useEffect } from "react";
import { useDispatch } from "react-redux";

import { authApi } from "../api/auth.api";
import { authLoaded, logout as logoutAction } from "./auth.slice";

let bootstrapStarted = false;

export default function AuthBootstrap() {
  const dispatch = useDispatch();

  useEffect(() => {
    if (bootstrapStarted) {
      return;
    }

    bootstrapStarted = true;

    dispatch(
      authApi.endpoints.me.initiate(undefined, {
        forceRefetch: true,
      }) as any,
    )
      .unwrap()
      .catch(() => {
        dispatch(logoutAction());
      })
      .finally(() => {
        dispatch(authLoaded());
      });
  }, [dispatch]);

  return null;
}
