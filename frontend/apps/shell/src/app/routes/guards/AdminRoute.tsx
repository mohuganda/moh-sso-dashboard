import { InlineLoading } from "@carbon/react";
import type { JSX } from "react";
import { useSelector } from "react-redux";
import { Navigate } from "react-router-dom";

import { selectAuthenticated, selectAuthLoaded } from "@moh-sso/auth";

export const AdminRoute = ({ children }: { children: JSX.Element }) => {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);

  if (!loaded) {
    return <InlineLoading description="Checking permissions…" />;
  }

  // ❌ Not authenticated → public newsfeed
  if (!authenticated) {
    return <Navigate to="/" replace />;
  }

  return children;
};
