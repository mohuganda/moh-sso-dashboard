import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import {
  notifyMicrofrontendNavigation,
  resolveRuntimeBasename,
  type MicrofrontendRuntimeProps,
} from "@moh-sso/microfrontend";
import { MohThemeProvider } from "@moh-sso/ui";
import { useEffect, type ReactNode } from "react";
import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";

import DiseaseDetailsPage from "./pages/surveillance-details/surveillance-details.component";
import SurveillanceDashboardPage from "./pages/surveillance.component";

function SurveillanceAccessDenied() {
  return (
    <div className="surveillance-dashboard-page">
      <div className="surveillance-dashboard-page__header">
        <div className="surveillance-dashboard-page__hero">
          <div className="surveillance-dashboard-page__hero-main">
            <div className="surveillance-dashboard-page__hero-copy">
              <h1 className="surveillance-dashboard-page__title">Access denied</h1>
              <p className="surveillance-dashboard-page__subtitle">
                You need surveillance read access to view this workspace.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function requireSurveillanceRead(element: ReactNode) {
  return (
    <PermissionGuard
      permission={PERMISSIONS.surveillanceRead}
      fallback={<SurveillanceAccessDenied />}
    >
      {element}
    </PermissionGuard>
  );
}

function SurveillanceNavigationSync() {
  const { pathname, search } = useLocation();
  useEffect(() => notifyMicrofrontendNavigation(), [pathname, search]);
  return null;
}

function SurveillanceRouteFallback() {
  const { search } = useLocation();
  return <Navigate to={{ pathname: "/", search }} replace />;
}

export function SurveillanceRoot(props: MicrofrontendRuntimeProps) {
  return (
    <div className="moh-microfrontend-root">
      <MohThemeProvider theme="white">
        <BrowserRouter
          basename={resolveRuntimeBasename(props.basename || "/apps/dwh/surveillance")}
        >
          <SurveillanceNavigationSync />
          <Routes>
            <Route index element={requireSurveillanceRead(<SurveillanceDashboardPage />)} />
            <Route path=":diseaseName" element={requireSurveillanceRead(<DiseaseDetailsPage />)} />
            <Route path="*" element={<SurveillanceRouteFallback />} />
          </Routes>
        </BrowserRouter>
      </MohThemeProvider>
    </div>
  );
}
