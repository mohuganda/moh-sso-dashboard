import type { ReactElement } from "react";
import { Navigate, Outlet, Route } from "react-router-dom";

import { ProtectedRoute } from "./guards/ProtectedRoute";
import { UserRoute } from "./guards/UserRoute";
import UserLayout from "../layouts/user/UserLayout";
import DataVisualizer from "@/features/data-visualizer/pages/data-visualizer";
import FileUpload from "@/features/documents/pages/file-upload";
import IssueTracker from "@/features/issue-tracker/pages/issue-tracker";
import NewsFeedPage from "@/app/newsfeed/pages/news_feed.component";
import DocumentDetailsPage from "@/features/documents/pages/document-details.component";
import DocumentPage from "@/features/documents/pages/documents.component";
import MyProfilePage from "@/app/settings/pages/Profile/profile.component";
import SecurityPage from "@/app/settings/pages/security/security.component";
import ActiveSessionsPage from "@/app/settings/pages/sessions/active-sesssions.component";
import DiseaseDetailsPage from "@/features/surveillance/pages/surveillance-details/surveillance-details.component";
import SurveillancePage from "@/features/surveillance/pages/surveillance.component";
import AbsenceRequests from "@/features/utilities/pages/absence-requests.component";
import Elearning from "@/features/utilities/pages/elearning.component";
import LeavePlan from "@/features/utilities/pages/leave-plan.component";
import MyAbsenceDashboard from "@/features/utilities/pages/my-absence-dashboard.component";
import MyTimeSheet from "@/features/utilities/pages/my-timesheet.component";

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

    <Route path="dwh">
      <Route index element={<Navigate to="data-visualizer" replace />} />
      <Route path="data-visualizer" element={userPage(<DataVisualizer />)} />
      <Route path="dashboards" element={userPage(<div>Dashboards</div>)} />
      <Route path="reports" element={userPage(<div>Reports</div>)} />
      <Route path="exports" element={userPage(<div>Data Exports</div>)} />
      <Route path="filesvr" element={userPage(<FileUpload />)} />
      <Route path="surveillance" element={userPage(<Outlet />)}>
        <Route index element={<SurveillancePage />} />
        <Route path=":diseaseName" element={<DiseaseDetailsPage />} />
      </Route>
      <Route path="issue-tracker" element={userPage(<IssueTracker />)} />
    </Route>

    <Route path="eservices">
      <Route index element={<Navigate to="ihris" replace />} />
      <Route path="ihris" element={userPage(<div>iHRIS</div>)} />
      <Route path="meeting-manager" element={userPage(<div>Meeting Manager</div>)} />
      <Route path="action-tracker" element={userPage(<div>Action Tracker</div>)} />
      <Route path="clinician-outputs" element={userPage(<div>Clinician Outputs</div>)} />
      <Route path="leave-absence" element={userPage(<div>Leave & Absence</div>)} />
      <Route path="workplans" element={userPage(<div>Workplans</div>)} />
      <Route path="budget-tracker" element={userPage(<div>Budget Tracker</div>)} />
      <Route path="activity-reporting" element={userPage(<div>Activity Reporting</div>)} />
      <Route path="partner-management" element={userPage(<div>Partner Management</div>)} />
      <Route path="observatory-uploads" element={userPage(<div>Observatory Uploads</div>)} />
    </Route>

    <Route path="research-studies">
      <Route index element={<Navigate to="studies" replace />} />
      <Route path="studies" element={userPage(<div>Studies</div>)} />
      <Route path="datasets" element={userPage(<div>Datasets</div>)} />
      <Route path="ethics" element={userPage(<div>Ethics & Approvals</div>)} />
      <Route path="publications" element={userPage(<div>Publications</div>)} />
    </Route>

    <Route path="case-registers">
      <Route index element={<Navigate to="external-referrals" replace />} />
      <Route path="external-referrals" element={userPage(<div>External Referrals</div>)} />
      <Route path="disease-registers" element={userPage(<div>Disease Registers</div>)} />
    </Route>

    <Route path="outbreak-management">
      <Route index element={<Navigate to="signals-alerts" replace />} />
      <Route path="signals-alerts" element={userPage(<div>Signals & Alerts</div>)} />
      <Route path="poe-management" element={userPage(<div>PoE Management</div>)} />
      <Route path="case-management" element={userPage(<div>Case Management</div>)} />
    </Route>

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

    <Route path="utilities">
      <Route index element={<Navigate to="self-service/timesheet" replace />} />
      <Route path="self-service">
        <Route path="timesheet" element={userPage(<MyTimeSheet />)} />
        <Route path="elearning" element={userPage(<Elearning />)} />
        <Route path="leave-plan" element={userPage(<LeavePlan />)} />
        <Route path="absence-requests" element={userPage(<AbsenceRequests />)} />
        <Route path="absence-dashboard" element={userPage(<MyAbsenceDashboard />)} />
        <Route path="eservice">
          <Route path="document-upload" element={userPage(<DocumentPage />)} />
          <Route path="document-upload/:id" element={userPage(<DocumentDetailsPage />)} />
          <Route path="service-access" element={<div>Service Access</div>} />
          <Route path="equipment-request" element={<div>Equipment Request</div>} />
        </Route>
      </Route>
    </Route>

    <Route path="settings">
      <Route index element={<Navigate to="profile" replace />} />
      <Route path="profile" element={<MyProfilePage />} />
      <Route path="sessions" element={<ActiveSessionsPage />} />
      <Route path="security" element={<SecurityPage />} />
    </Route>
  </Route>
);
