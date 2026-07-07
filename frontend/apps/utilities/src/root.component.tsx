import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { ComingSoon, MohThemeProvider } from "@moh-sso/ui";

function normalizeBasename(basename?: string) {
  const normalized = basename?.replace(/\/+$/, "");

  if (!normalized || normalized.endsWith("/apps/utilities")) {
    return "/apps/utilities/self-service";
  }

  return normalized;
}

export function UtilitiesRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveRuntimeBasename(normalizeBasename(props.basename));

  return (
    <div className="moh-microfrontend-root">
      <MohThemeProvider theme="white">
        <BrowserRouter basename={basename}>
          <Routes>
            <Route index element={<Navigate to="timesheet" replace />} />

            <Route path="timesheet" element={<ComingSoon title="Timesheet" />} />

            <Route path="elearning" element={<ComingSoon title="E-learning" />} />

            <Route path="leave-plan" element={<ComingSoon title="Leave Plan" />} />

            <Route path="absence-requests" element={<ComingSoon title="Absence Requests" />} />

            <Route path="absence-dashboard" element={<ComingSoon title="Absence Dashboard" />} />

            <Route path="eservice">
              <Route index element={<Navigate to="service-access" replace />} />
              <Route path="service-access" element={<ComingSoon title="Service Access" />} />
              <Route path="equipment-request" element={<ComingSoon title="Equipment Request" />} />
            </Route>

            <Route path="*" element={<Navigate to="timesheet" replace />} />
          </Routes>
        </BrowserRouter>
      </MohThemeProvider>
    </div>
  );
}
