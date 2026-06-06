import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import AbsenceRequests from "./pages/absence-requests.component";
import Elearning from "./pages/elearning.component";
import LeavePlan from "./pages/leave-plan.component";
import MyAbsenceDashboard from "./pages/my-absence-dashboard.component";
import MyTimeSheet from "./pages/my-timesheet.component";

export function UtilitiesRoot(props: MicrofrontendRuntimeProps) {
  const basename =
    props.basename === "/apps/utilities"
      ? "/apps/utilities/self-service"
      : props.basename || "/apps/utilities/self-service";

  return (
    <BrowserRouter basename={basename}>
      <Routes>
        <Route index element={<Navigate to="timesheet" replace />} />
        <Route path="timesheet" element={<MyTimeSheet />} />
        <Route path="elearning" element={<Elearning />} />
        <Route path="leave-plan" element={<LeavePlan />} />
        <Route path="absence-requests" element={<AbsenceRequests />} />
        <Route path="absence-dashboard" element={<MyAbsenceDashboard />} />
      </Routes>
    </BrowserRouter>
  );
}
