import type { ReactElement } from "react";
import { Navigate, Route } from "react-router-dom";

import { ProtectedRoute } from "./guards/ProtectedRoute";
import { UserRoute } from "./guards/UserRoute";
import UserLayout from "../layouts/user/UserLayout";
import { ComingSoon } from "@moh-sso/ui";

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

const userPage = (element: ReactElement) => <UserRoute>{element}</UserRoute>;

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
            basename="/portal/apps/dwh/data-visualizer"
          />,
        )}
      />

      <Route path="dashboards" element={userPage(<ComingSoon title="Dashboards" />)} />
      <Route path="reports" element={userPage(<ComingSoon title="Reports" />)} />
      <Route path="exports" element={userPage(<ComingSoon title="Data Exports" />)} />

      <Route
        path="filesvr/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/documents"
            lifecycles={documentsLifecycles}
            basename="/portal/apps/dwh/filesvr"
          />,
        )}
      />

      <Route
        path="surveillance/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/surveillance"
            lifecycles={surveillanceLifecycles}
            basename="/portal/apps/dwh/surveillance"
          />,
        )}
      />

      <Route
        path="issue-tracker/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/issue-tracker"
            lifecycles={issueTrackerLifecycles}
            basename="/portal/apps/dwh/issue-tracker"
          />,
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
          basename="/portal/apps/eservices"
        />,
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
      <Route index element={<Navigate to="self-service/timesheet" replace />} />

      <Route path="self-service">
        <Route
          path="timesheet/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/portal/apps/utilities/self-service/timesheet"
            />,
          )}
        />

        <Route
          path="elearning/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/portal/apps/utilities/self-service/elearning"
            />,
          )}
        />

        <Route
          path="leave-plan/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/portal/apps/utilities/self-service/leave-plan"
            />,
          )}
        />

        <Route
          path="absence-requests/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/portal/apps/utilities/self-service/absence-requests"
            />,
          )}
        />

        <Route
          path="absence-dashboard/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/portal/apps/utilities/self-service/absence-dashboard"
            />,
          )}
        />

        <Route path="eservice">
          <Route
            path="document-upload/*"
            element={userPage(
              <SingleSpaApp
                appName="@moh-sso/documents"
                lifecycles={documentsLifecycles}
                basename="/portal/apps/utilities/self-service/eservice/document-upload"
              />,
            )}
          />

          <Route path="service-access" element={userPage(<ComingSoon title="Service Access" />)} />

          <Route
            path="equipment-request"
            element={userPage(<ComingSoon title="Equipment Request" />)}
          />
        </Route>
      </Route>
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
