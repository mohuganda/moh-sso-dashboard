import { InlineLoading } from "@carbon/react";
import type { JSX } from "react";
import { useEffect, useRef } from "react";
import { useSelector } from "react-redux";
import { Outlet } from "react-router-dom";

import { API } from "@/lib/constants/api.constants";
import {
  selectAuthenticated,
  selectAuthLoaded,
  selectAuthLoading,
} from "@/store/auth/auth.selectors";

export const ProtectedRoute = ({ children }: { children?: JSX.Element }) => {
  const authenticated = useSelector(selectAuthenticated);
  const loaded = useSelector(selectAuthLoaded);
  const loading = useSelector(selectAuthLoading);

  const redirectStarted = useRef(false);

  useEffect(() => {
    if (!loaded || loading || authenticated) {
      return;
    }

    if (redirectStarted.current) {
      return;
    }

    redirectStarted.current = true;

    window.location.assign(API.auth.login());
  }, [loaded, loading, authenticated]);

  if (!loaded || loading) {
    return <InlineLoading description="Checking session…" />;
  }

  if (!authenticated) {
    return <InlineLoading description="Redirecting to login…" />;
  }

  return children ? children : <Outlet />;
};
