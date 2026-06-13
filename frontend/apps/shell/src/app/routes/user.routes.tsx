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
          PERMISSIONS.reportBrowserRead,
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
          PERMISSIONS.outbreakAccess,
        )}
      />

      <Route
        path="issue-tracker/*"
        element={userPage(
          <SingleSpaApp
            appName="@moh-sso/issue-tracker"
            lifecycles={issueTrackerLifecyclAbsolutely. Then the RBAC model should become **system-aware**, not just user-role-aware.

Right now we have:

- Realm roles: broad identity roles like `admin`, `manager`, `user`
- Client roles: app/system-specific roles like `integrated-outbreak-system.admin`
- Permissions: app actions like `surveillance:read`, `documents:write`

What we should add is a middle layer:

```text
System -> Roles -> Permissions
```

Example:

```text
integrated-outbreak-system
  super_admin
    -> surveillance:*
    -> data_quality:*
    -> outbreak:*
  viewer
    -> surveillance:read
    -> reports:read

report-browser
  report-browser_access
    -> reports:read

dashboard-web
  dashboard-web_access
    -> portal:access
```

Here is the implementation prompt:

```text
Improve RBAC to be system-aware.

Workspace:
/Users/jaba/Documents/projects/web/moh-sso-dashboard

Goal:
Support users, systems/apps, system roles, and system permissions in authorization.

Conceptual model:
- Realm roles are global identity roles.
- Keycloak clients represent systems/apps.
- Client roles represent roles inside a system.
- Application permissions are derived from both:
  1. global realm roles
  2. system-specific client roles

Target model:
System:
- dashboard-web
- integrated-outbreak-system
- report-browser
- future systems

System role examples:
dashboard-web:
- dashboard-web_access
- integrated-outbreak-system_access

integrated-outbreak-system:
- integrated-outbreak-system_access
- super_admin
- admin
- manager
- viewer
- data_entry
- lab_technician
- surveillance_officer

report-browser:
- report-browser_access

Permission examples:
portal:access
systems:read
systems:launch

report_browser:read

surveillance:read
surveillance:import
surveillance:manage_locations
surveillance:manage_alerts

documents:read
documents:write
data_quality:read
data_quality:write
data_quality:resolve

Implementation tasks:

1. Update backend/internal/authz/permissions.go

Add:
- portal:access
- systems:read
- systems:launch
- report_browser:read
- outbreak:access
- outbreak:manage

Keep existing permissions.

2. Create backend/internal/authz/systems.go

Define system/client constants:
- SystemDashboardWeb = "dashboard-web"
- SystemIntegratedOutbreak = "integrated-outbreak-system"
- SystemReportBrowser = "report-browser"

Define system role constants:
- DashboardWebAccess = "dashboard-web_access"
- DashboardWebIntegratedOutbreakAccess = "integrated-outbreak-system_access"
- IntegratedOutbreakAccess = "integrated-outbreak-system_access"
- IntegratedOutbreakSuperAdmin = "super_admin"
- IntegratedOutbreakAdmin = "admin"
- IntegratedOutbreakManager = "manager"
- IntegratedOutbreakViewer = "viewer"
- IntegratedOutbreakDataEntry = "data_entry"
- IntegratedOutbreakLabTechnician = "lab_technician"
- IntegratedOutbreakSurveillanceOfficer = "surveillance_officer"
- ReportBrowserAccess = "report-browser_access"

3. Update backend/internal/authz/policy.go

Replace simple:
PermissionsForRoles(realmRoles)

With:
PermissionsForContext(realmRoles, clientRoles)

Permission derivation should combine:
- realm role permissions
- system/client role permissions

Rules:
realm admin:
- all permissions

realm manager:
- management/read permissions

realm user:
- basic portal permissions:
  portal:access
  systems:read
  systems:launch

dashboard-web.dashboard-web_access:
- portal:access
- systems:read
- systems:launch

dashboard-web.integrated-outbreak-system_access:
- outbreak:access
- surveillance:read

integrated-outbreak-system.integrated-outbreak-system_access:
- outbreak:access
- surveillance:read

integrated-outbreak-system.super_admin:
- outbreak:access
- outbreak:manage
- surveillance:read
- surveillance:import
- surveillance:manage_locations
- surveillance:manage_alerts
- data_quality:read
- data_quality:write
- data_quality:resolve
- documents:read
- documents:write
- documents:process

integrated-outbreak-system.admin:
- outbreak:access
- outbreak:manage
- surveillance:read
- surveillance:import
- surveillance:manage_locations
- surveillance:manage_alerts
- data_quality:read
- data_quality:write
- data_quality:resolve

integrated-outbreak-system.manager:
- outbreak:access
- surveillance:read
- data_quality:read
- data_quality:write
- documents:read

integrated-outbreak-system.viewer:
- outbreak:access
- surveillance:read
- data_quality:read
- documents:read

integrated-outbreak-system.data_entry:
- outbreak:access
- surveillance:read
- surveillance:import
- data_quality:read
- data_quality:write

integrated-outbreak-system.lab_technician:
- outbreak:access
- surveillance:read
- data_quality:read
- data_quality:write

integrated-outbreak-system.surveillance_officer:
- outbreak:access
- surveillance:read
- surveillance:import
- surveillance:manage_alerts

report-browser.report-browser_access:
- report_browser:read

4. Update backend/internal/authz/context.go

Context should include:
- UserID
- RealmRoles
- ClientRoles
- Permissions
- Systems or AccessibleSystems

Add:
AccessibleSystems []string

Derive accessible systems from client roles:
- dashboard-web role -> dashboard-web
- integrated-outbreak-system role -> integrated-outbreak-system
- report-browser role -> report-browser

Add helpers:
- HasSystem(system string) bool
- HasSystemRole(system string, role string) bool
- HasPermission(permission Permission) bool

5. Update Keycloak AuthUser response

In backend/internal/keycloak/keycloak.go, return:
- permissions: string[]
- systems: string[]

Do not remove:
- isAdmin
- isUser
- realmRoles
- clientRoles

6. Update frontend auth types

In:
frontend/packages/auth/src/auth/auth.types.ts

Add:
systems: string[]

Add new permission literals:
- portal:access
- systems:read
- systems:launch
- report_browser:read
- outbreak:access
- outbreak:manage

7. Update frontend permissions constants

In:
frontend/packages/auth/src/auth/permissions.ts

Add all new permissions.

Create:
frontend/packages/auth/src/auth/systems.ts

Export:
- SYSTEMS.dashboardWeb
- SYSTEMS.integratedOutbreak
- SYSTEMS.reportBrowser

Export role constants grouped by system.

8. Update frontend authorization hook

In:
frontend/packages/auth/src/auth/useAuthorization.ts

Add:
- hasSystem(system)
- hasSystemRole(system, role)

9. Update shell route checks

Use system-aware checks where appropriate:
- dashboard shell requires portal:access
- integrated outbreak / surveillance routes require outbreak:access or surveillance:read
- report browser routes require report_browser:read
- launcher cards should show only systems the user can access

10. Add tests

Backend tests:
- realm admin gets all permissions
- user with dashboard-web_access gets portal access
- user with IOS viewer gets surveillance read but not import
- user with IOS super_admin gets import/manage permissions
- user with report-browser_access gets report_browser:read
- accessible systems are derived correctly

Frontend:
- typecheck selectors/hooks
- optionally test useAuthorization if test setup exists

Verification:
Backend:
GOCACHE=/private/tmp/moh-sso-go-build go test ./...

Frontend:
source ~/.nvm/nvm.sh && nvm use v20.20.1 && npm run typecheck
source ~/.nvm/nvm.sh && nvm use v20.20.1 && npm run lint
source ~/.nvm/nvm.sh && nvm use v20.20.1 && npm run build:shell

Acceptance criteria:
- Authorization derives permissions from realm roles and system/client roles.
- `/auth/me` returns permissions and accessible systems.
- Frontend can check permissions, systems, and system roles.
- Existing admin/user behavior still works.
- Keycloak realm-export roles map cleanly to our application permissions.
- Tests pass.
```es}
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
        PERMISSIONS.systemsLaunch,
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
