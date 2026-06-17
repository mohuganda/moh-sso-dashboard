import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { MohThemeProvider } from "@moh-sso/ui";

import ActionTrackerPage from "./pages/action-tracker.page";
import ActivityReportingPage from "./pages/activity-reporting.page";
import BudgetTrackerPage from "./pages/budget-tracker.page";
import ClinicianOutputsPage from "./pages/clinician-outputs.page";
import LeaveAbsenceManagementPage from "./pages/leave-absence-management.page";
import MeetingManagerPage from "./pages/meeting-manager.page";
import WorkplansPage from "./pages/workplans.page";

function normalizeBasename(basename?: string) {
  if (!basename || basename === "/apps/eservices") {
    return "/apps/eservices";
  }

  return basename.replace(/\/+$/, "");
}

export function EServicesRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveRuntimeBasename(normalizeBasename(props.basename));

  return (
    <MohThemeProvider theme="white">
      <BrowserRouter basename={basename}>
        <Routes>
          <Route index element={<Navigate to="/ihris/meeting-manager" replace />} />

          <Route path="ihris">
            <Route index element={<Navigate to="/ihris/meeting-manager" replace />} />

            <Route path="meeting-manager" element={<MeetingManagerPage />} />

            <Route path="action-tracker" element={<ActionTrackerPage />} />

            <Route path="clinician-outputs" element={<ClinicianOutputsPage />} />

            <Route path="leave-absence-management" element={<LeaveAbsenceManagementPage />} />

            <Route path="workplans" element={<WorkplansPage />} />

            <Route path="budget-tracker" element={<BudgetTrackerPage />} />

            <Route path="activity-reporting" element={<ActivityReportingPage />} />

            <Route path="*" element={<Navigate to="/ihris/meeting-manager" replace />} />
          </Route>

          <Route path="partner-management" element={<div>Partner Management</div>} />

          <Route path="observatory-uploads" element={<div>Observatory Uploads</div>} />

          <Route path="*" element={<Navigate to="/ihris/meeting-manager" replace />} />
        </Routes>
      </BrowserRouter>
    </MohThemeProvider>
  );
}
