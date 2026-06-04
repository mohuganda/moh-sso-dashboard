import { Navigate, Route } from "react-router-dom";

import { AdminRoute } from "./guards/AdminRoute";
import { ProtectedRoute } from "./guards/ProtectedRoute";
import AdminLayout from "../layouts/admin/AdminLayout";
import { AnnouncementsPage } from "@/features/announcements/pages/announcements.components";
import AuditLogsPage from "@/features/audit/pages/audit_component";
import ClientsPage from "@/features/clients/pages/client.component";
import EmailOutbox from "@/features/email/pages/email-outbox.component";
import HomePage from "@/features/home/pages/home.component";
import UsersPage from "@/features/users/pages/user.component";

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
          <UsersPage />
        </AdminRoute>
      }
    />
    <Route
      path="clients"
      element={
        <AdminRoute>
          <ClientsPage />
        </AdminRoute>
      }
    />
    <Route
      path="audit-logs"
      element={
        <AdminRoute>
          <AuditLogsPage />
        </AdminRoute>
      }
    />
    <Route
      path="announcements"
      element={
        <AdminRoute>
          <AnnouncementsPage />
        </AdminRoute>
      }
    />
    <Route
      path="emails"
      element={
        <AdminRoute>
          <EmailOutbox />
        </AdminRoute>
      }
    />
  </Route>
);
