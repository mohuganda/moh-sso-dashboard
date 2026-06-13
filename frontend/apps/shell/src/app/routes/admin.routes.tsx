import { Navigate, Route } from "react-router-dom";
import type { ReactElement } from "react";

import { AdminRoute } from "./guards/AdminRoute";
import { PermissionRoute } from "./guards/PermissionRoute";
import { ProtectedRoute } from "./guards/ProtectedRoute";
import AdminLayout from "../layouts/admin/admin-layout.component";
import HomePage from "@/app/home/pages/home.component";
import { SingleSpaApp } from "@/app/microfrontends/SingleSpaApp";
import {
  PERMISSIONS,
  type Permission,
} from "@moh-sso/auth";

import {
  announcementsLifecycles,
  auditLifecycles,
  clientsLifecycles,
  emailLifecycles,
  usersLifecycles,
} from "@/app/microfrontends/lifecycles";

const adminPage = (permission: Permission, element: ReactElement) => (
  <AdminRoute>
    <PermissionRoute permission={permission}>{element}</PermissionRoute>
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
          <HomePage />
        </AdminRoute>
      }
    />
    <Route
      path="users"
      element={adminPage(
        PERMISSIONS.usersRead,
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
        PERMISSIONS.clientsRead,
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
        PERMISSIONS.auditRead,
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
        PERMISSIONS.announcementsRead,
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
        PERMISSIONS.emailRead,
        <SingleSpaApp
          appName="@moh-sso/email"
          lifecycles={emailLifecycles}
          basename="/admin/emails"
        />,
      )}
    />
  </Route>
);
