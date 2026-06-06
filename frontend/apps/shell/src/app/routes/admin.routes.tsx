import { Navigate, Route } from "react-router-dom";

import { AdminRoute } from "./guards/AdminRoute";
import { ProtectedRoute } from "./guards/ProtectedRoute";
import AdminLayout from "../layouts/admin/AdminLayout";
import HomePage from "@/app/home/pages/home.component";
import { SingleSpaApp } from "@/app/microfrontends/SingleSpaApp";
import {
  announcementsLifecycles,
  auditLifecycles,
  clientsLifecycles,
  emailLifecycles,
  usersLifecycles,
} from "@/app/microfrontends/lifecycles";

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
      element={
        <AdminRoute>
          <SingleSpaApp
            appName="@moh-sso/users"
            lifecycles={usersLifecycles}
            basename="/admin/users"
          />
        </AdminRoute>
      }
    />
    <Route
      path="clients"
      element={
        <AdminRoute>
          <SingleSpaApp
            appName="@moh-sso/clients"
            lifecycles={clientsLifecycles}
            basename="/admin/clients"
          />
        </AdminRoute>
      }
    />
    <Route
      path="audit-logs"
      element={
        <AdminRoute>
          <SingleSpaApp
            appName="@moh-sso/audit"
            lifecycles={auditLifecycles}
            basename="/admin/audit-logs"
          />
        </AdminRoute>
      }
    />
    <Route
      path="announcements"
      element={
        <AdminRoute>
          <SingleSpaApp
            appName="@moh-sso/announcements"
            lifecycles={announcementsLifecycles}
            basename="/admin/announcements"
          />
        </AdminRoute>
      }
    />
    <Route
      path="emails"
      element={
        <AdminRoute>
          <SingleSpaApp
            appName="@moh-sso/email"
            lifecycles={emailLifecycles}
            basename="/admin/emails"
          />
        </AdminRoute>
      }
    />
  </Route>
);
