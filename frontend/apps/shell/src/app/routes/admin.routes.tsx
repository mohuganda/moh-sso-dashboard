import { Navigate, Route } from "react-router-dom";
import type { ReactElement } from "react";

import { AdminRoute } from "./guards/AdminRoute";
import { PermissionRoute } from "./guards/PermissionRoute";
import { ProtectedRoute } from "./guards/ProtectedRoute";
import AdminLayout from "../layouts/admin/admin-layout.component";
import HomePage from "@/app/home/pages/home.component";
import { SingleSpaApp } from "@/app/microfrontends/SingleSpaApp";
import { announcementsRoute } from "@moh-sso/announcements";
import { auditRoute } from "@moh-sso/audit";
import { clientsRoute } from "@moh-sso/clients";
import { emailRoute } from "@moh-sso/email";
import { PERMISSIONS, type Permission, type System } from "@moh-sso/auth";
import type { MicrofrontendRoute } from "@moh-sso/microfrontend";
import { usersRoute } from "@moh-sso/users";

import {
  announcementsLifecycles,
  auditLifecycles,
  clientsLifecycles,
  emailLifecycles,
  usersLifecycles,
} from "@/app/microfrontends/lifecycles";

const routePermissions = (route: MicrofrontendRoute) =>
  route.requiredPermissions as Permission[] | undefined;

const routeAnyPermissions = (route: MicrofrontendRoute) =>
  route.requiredAnyPermissions as Permission[] | undefined;

const routeSystems = (route: MicrofrontendRoute) => route.requiredSystems as System[] | undefined;

const adminPage = (route: MicrofrontendRoute, element: ReactElement) => (
  <AdminRoute>
    <PermissionRoute
      allOf={routePermissions(route)}
      anyOf={routeAnyPermissions(route)}
      systems={routeSystems(route)}
      systemRoles={route.requiredSystemRoles}
    >
      {element}
    </PermissionRoute>
  </AdminRoute>
);

export const adminRoutes = (
  <Route
    path="/admin"
    element={
      <ProtectedRoute>
        <AdminLayout />
      </ProtectedRoute>
    }
  >
    <Route index element={<Navigate to="home" replace />} />
    <Route
      path="home"
      element={
        <AdminRoute>
          <PermissionRoute permission={PERMISSIONS.portalAccess}>
            <HomePage />
          </PermissionRoute>
        </AdminRoute>
      }
    />
    <Route
      path="users"
      element={adminPage(
        usersRoute,
        <SingleSpaApp
          appName="@moh-sso/users"
          lifecycles={usersLifecycles}
          basename="/admin/users"
        />,
      )}
    />
    <Route
      path="clients"
      element={adminPage(
        clientsRoute,
        <SingleSpaApp
          appName="@moh-sso/clients"
          lifecycles={clientsLifecycles}
          basename="/admin/clients"
        />,
      )}
    />
    <Route
      path="audit-logs"
      element={adminPage(
        auditRoute,
        <SingleSpaApp
          appName="@moh-sso/audit"
          lifecycles={auditLifecycles}
          basename="/admin/audit-logs"
        />,
      )}
    />
    <Route
      path="announcements"
      element={adminPage(
        announcementsRoute,
        <SingleSpaApp
          appName="@moh-sso/announcements"
          lifecycles={announcementsLifecycles}
          basename="/admin/announcements"
        />,
      )}
    />
    <Route
      path="emails"
      element={adminPage(
        emailRoute,
        <SingleSpaApp
          appName="@moh-sso/email"
          lifecycles={emailLifecycles}
          basename="/admin/emails"
        />,
      )}
    />
  </Route>
);
