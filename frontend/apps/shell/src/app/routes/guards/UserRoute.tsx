// components/UserRoute.tsx
import { InlineLoading } from "@carbon/react";
import type { JSX } from "react";
import { useSelector } from "react-redux";
import { Navigate } from "react-router-dom";

import {
  PERMISSIONS,
  selectAuthenticated,
  selectAuthLoaded,
  useAuthorization,
} from "@moh-sso/auth";

export const UserRoute = ({ children }: { children: JSX.Element }) => {
  const loaded = useSelector(selectAuthLoaded);
  const authenticated = useSelector(selectAuthenticated);
  const { can } = useAuthorization();

  if (!loaded) {
    return <InlineLoading description="Checking session…" />;
  }

  if (!authenticated || !can(PERMISSIONS.portalAccess)) {
    return <Navigate to="/" replace />;
  }

  return children;
};
