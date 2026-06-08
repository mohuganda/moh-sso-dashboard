import type { ReactElement } from "react";
import { Navigate, Route } from "react-router-dom";

import { ProtectedRoute } from "./guards/ProtectedRoute";
import { UserRoute } from "./guards/UserRoute";
import UserLayout from "../layouts/user/UserLayout";

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
            basename="/apps/dwh/data-visualizer"
          />,
        )}
      />

      <Route path="dashboards" element={userPage(<div>Dashboards</div>)} />
      <Route path="reports" element={userPage(<div>Reports</div>)} />
      <Route path="exports" element={userPage(<div>Data Exports</div>)} />

      <Route
        path="filesvr/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/documents"
            lifecycles={documentsLifecycles}
            basename="/apps/dwh/filesvr"
          />,
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
        )}
      />
    </Route>

    {/* =========================
        E-SERVICES MICROFRONTEND
        Internal routes are owned by @moh-sso/e-services
       ========================= */}
    <Route
      path="eservices/*"
      element={userPage(
        <SingleSpaApp
          appName="@moh-sso/e-services"
          lifecycles={eServicesLifecycles}
          basename="/apps/eservices"
        />,
      )}
    />

    {/* =========================
        RESEARCH STUDIES
       ========================= */}
    <Route path="research-studies">
      <Route index element={<Navigate to="studies" replace />} />
      <Route path="studies" element={userPage(<div>Studies</div>)} />
      <Route path="datasets" element={userPage(<div>Datasets</div>)} />
      <Route path="ethics" element={userPage(<div>Ethics & Approvals</div>)} />
      <Route path="publications" element={userPage(<div>Publications</div>)} />
    </Route>

    {/* =========================
        CASE REGISTERS
       ========================= */}
    <Route path="case-registers">
      <Route index element={<Navigate to="external-referrals" replace />} />
      <Route path="external-referrals" element={userPage(<div>External Referrals</div>)} />
      <Route path="disease-registers" element={userPage(<div>Disease Registers</div>)} />
    </Route>

    {/* =========================
        OUTBREAK MANAGEMENT
       ========================= */}
    <Route path="outbreak-management">
      <Route index element={<Navigate to="signals-alerts" replace />} />
      <Route path="signals-alerts" element={userPage(<div>Signals & Alerts</div>)} />
      <Route path="poe-management" element={userPage(<div>PoE Management</div>)} />
      <Route path="case-management" element={userPage(<div>Case Management</div>)} />
    </Route>

    {/* =========================
        REFERENCE REGISTERS
       ========================= */}
    <Route path="reference-registers">
      <Route index element={<Navigate to="facility-register" replace />} />
      <Route path="facility-register" element={userPage(<div>Facility Register</div>)} />

      <Route path="terminology">
        <Route index element={<Navigate to="test-menu" replace />} />
        <Route path="test-menu" element={userPage(<div>Test Menu</div>)} />
        <Route path="pharmaceuticals" element={userPage(<div>Pharmaceuticals</div>)} />
        <Route path="procedures" element={userPage(<div>Procedures</div>)} />
        <Route path="equipment" element={userPage(<div>Equipment</div>)} />
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
              basename="/apps/utilities/self-service/timesheet"
            />,
          )}
        />

        <Route
          path="elearning/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/apps/utilities/self-service/elearning"
            />,
          )}
        />

        <Route
          path="leave-plan/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/apps/utilities/self-service/leave-plan"
            />,
          )}
        />

        <Route
          path="absence-requests/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/apps/utilities/self-service/absence-requests"
            />,
          )}
        />

        <Route
          path="absence-dashboard/*"
          element={userPage(
            <SingleSpaApp
              appName="@moh-sso/utilities"
              lifecycles={utilitiesLifecycles}
              basename="/apps/utilities/self-service/absence-dashboard"
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
                basename="/apps/utilities/self-service/eservice/document-upload"
              />,
            )}
          />

          <Route
            path="document-upload/:id/*"
            element={userPage(
              <SingleSpaApp
                appName="@moh-sso/documents"
                lifecycles={documentsLifecycles}
                basename="/apps/utilities/self-service/eservice/document-upload"
              />,
            )}
          />

          <Route path="service-access" element={userPage(<div>Service Access</div>)} />
          <Route path="equipment-request" element={userPage(<div>Equipment Request</div>)} />
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
