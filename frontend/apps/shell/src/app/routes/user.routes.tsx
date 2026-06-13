import type { ReactElement } from "react";
import { Navigate, Route } from "react-router-dom";

import { ProtectedRoute } from "./guards/ProtectedRoute";
import { PermissionRoute } from "./guards/PermissionRoute";
import { UserRoute } from "./guards/UserRoute";
import UserLayout from "../layouts/user/user-layout.component";
import { ComingSoon } from "@moh-sso/ui";
import { PERMISSIONS, type Permission } from "@moh-sso/auth";

import NewsFeedPage from "@/app/newsfeed/pages/news_feed.component";
import MyProfilePage from "@/app/settings/pages/Profile/profile.component";
import SecurityPage from "@/app/settings/pages/security/security.component";
import ActiveSessionsPage from "@/app/settings/pages/sessions/active-sesssions.component";

import { SingleSpaApp } from "@/app/microfrontends/SingleSpaApp";

import {
  dataVisualizerLifecycles,
  documentsLifecycles,
  issueTrackerLifecycles,
  surveillanceLifecycles,
  utilitiesLifecycles,
  eServicesLifecycles,
} from "@/app/microfrontends/lifecycles";

const userPage = (element: ReactElement, permission?: Permission) => {
  const guarded = permission ? (
    <PermissionRoute permission={permission}>{element}</PermissionRoute>
  ) : (
    element
  );

  return <UserRoute>{guarded}</UserRoute>;
};

export const userRoutes = (
  <Route
    path="/apps"
    element={
      <ProtectedRoute>
        <UserLayout />
      </ProtectedRoute>
    }
  >
    <Route index element={<Navigate to="news" replace />} />

    <Route path="news" element={userPage(<NewsFeedPage />)} />

    {/* =========================
        DWH
       ========================= */}
    <Route path="dwh">
      <Route index element={<Navigate to="data-visualizer" replace />} />

      <Route
        path="data-visualizer/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/data-visualizer"
            lifecycles={dataVisualizerLifecycles}
            basename="/apps/dwh/data-visualizer"
          />,
          PERMISSIONS.documentsRead,
        )}
      />

      <Route path="dashboards" element={userPage(<ComingSoon title="Dashboards" />)} />
      <Route
        path="reports/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/reports"
            lifecycles={utilitiesLifecycles}
            basename="/apps/dwh/reports"
          />,
        )}
      />
      <Route path="exports" element={userPage(<ComingSoon title="Data Exports" />)} />

      <Route
        path="filesvr/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/documents"
            lifecycles={documentsLifecycles}
            basename="/apps/dwh/filesvr"
          />,
          PERMISSIONS.documentsRead,
        )}
      />

      <Route
        path="surveillance/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/surveillance"
            lifecycles={surveillanceLifecycles}
            basename="/apps/dwh/surveillance"
          />,
          PERMISSIONS.surveillanceRead,
        )}
      />

      <Route
        path="issue-tracker/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/issue-tracker"
            lifecycles={issueTrackerLifecycles}
            basename="/apps/dwh/issue-tracker"
          />,
          PERMISSIONS.dataQualityRead,
        )}
      />
    </Route>

    {/* =========================
        E-SERVICES MICROFRONTEND
       ========================= */}
    <Route
      path="eservices/*"
      element={userPage(
        <SingleSpaApp
          appName="@moh-sso/e-services"
          lifecycles={eServicesLifecycles}
          basename="/apps/eservices"
        />,
        PERMISSIONS.documentsRead,
      )}
    />

    {/* =========================
        RESEARCH STUDIES
       ========================= */}
    <Route path="research-studies">
      <Route index element={<Navigate to="studies" replace />} />
      <Route path="studies" element={userPage(<ComingSoon title="Studies" />)} />
      <Route path="datasets" element={userPage(<ComingSoon title="Datasets" />)} />
      <Route path="ethics" element={userPage(<ComingSoon title="Ethics & Approvals" />)} />
      <Route path="publications" element={userPage(<ComingSoon title="Publications" />)} />
    </Route>

    {/* =========================
        CASE REGISTERS
       ========================= */}
    <Route path="case-registers">
      <Route index element={<Navigate to="external-referrals" replace />} />

      <Route
        path="external-referrals"
        element={userPage(<ComingSoon title="External Referrals" />)}
      />

      <Route
        path="disease-registers"
        element={userPage(<ComingSoon title="Disease Registers" />)}
      />
    </Route>

    {/* =========================
        OUTBREAK MANAGEMENT
       ========================= */}
    <Route path="outbreak-management">
      <Route index element={<Navigate to="signals-alerts" replace />} />

      <Route path="signals-alerts" element={userPage(<ComingSoon title="Signals & Alerts" />)} />

      <Route path="poe-management" element={userPage(<ComingSoon title="PoE Management" />)} />

      <Route path="case-management" element={userPage(<ComingSoon title="Case Management" />)} />
    </Route>

    {/* =========================
        REFERENCE REGISTERS
       ========================= */}
    <Route path="reference-registers">
      <Route index element={<Navigate to="facility-register" replace />} />

      <Route
        path="facility-register"
        element={userPage(<ComingSoon title="Facility Register" />)}
      />

      <Route path="terminology">
        <Route index element={<Navigate to="test-menu" replace />} />

        <Route path="test-menu" element={userPage(<ComingSoon title="Test Menu" />)} />

        <Route path="pharmaceuticals" element={userPage(<ComingSoon title="Pharmaceuticals" />)} />

        <Route path="procedures" element={userPage(<ComingSoon title="Procedures" />)} />

        <Route path="equipment" element={userPage(<ComingSoon title="Equipment" />)} />
      </Route>
    </Route>

    {/* =========================
    UTILITIES MICROFRONTEND
   ========================= */}
    <Route path="utilities">
      <Route index element={<Navigate to="self-service" replace />} />

      {/* E-Service routes should stay in the host because they mount other apps */}
      <Route path="self-service/eservice">
        <Route index element={<Navigate to="document-upload" replace />} />

        <Route
          path="document-upload/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/documents"
              lifecycles={documentsLifecycles}
              basename="/apps/utilities/self-service/eservice/document-upload"
            />,
            PERMISSIONS.documentsWrite,
          )}
        />

        <Route path="service-access" element={userPage(<ComingSoon title="Service Access" />)} />

        <Route
          path="equipment-request"
          element={userPage(<ComingSoon title="Equipment Request" />)}
        />
      </Route>

      {/* Mount utilities once here */}
      <Route
        path="self-service/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/utilities"
            lifecycles={utilitiesLifecycles}
            basename="/apps/utilities/self-service"
          />,
        )}
      />
    </Route>

    {/* =========================
        SETTINGS
       ========================= */}
    <Route path="settings">
      <Route index element={<Navigate to="profile" replace />} />
      <Route path="profile" element={userPage(<MyProfilePage />)} />
      <Route path="sessions" element={userPage(<ActiveSessionsPage />)} />
      <Route path="security" element={userPage(<SecurityPage />)} />
    </Route>

    {/* =========================
        FALLBACK
       ========================= */}
    <Route path="*" element={<Navigate to="news" replace />} />
  </Route>
);
