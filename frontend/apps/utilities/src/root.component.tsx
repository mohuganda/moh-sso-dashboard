import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { ComingSoon, MohThemeProvider } from "@moh-sso/ui";

function normalizeBasename(basename?: string) {
  if (!basename || basename === "/apps/utilities") {
    return "/apps/utilities/self-service";
  }

  return basename.replace(/\/+$/, "");
}

export function UtilitiesRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveRuntimeBasename(normalizeBasename(props.basename));

  return (
    <MohThemeProvider theme="white">
      <BrowserRouter basename={basename}>
        <Routes>
          <Route index element={<Navigate to="/timesheet" replace />} />

          <Route path="timesheet" element={<ComingSoon title="Timesheet" />} />

          <Route path="elearning" element={<ComingSoon title="E-learning" />} />

          <Route path="leave-plan" element={<ComingSoon title="Leave Plan" />} />

          <Route path="absence-requests" element={<ComingSoon title="Absence Requests" />} />

          <Route path="absence-dashboard" element={<ComingSoon title="Absence Dashboard" />} />

          <Route path="*" element={<Navigate to="timesheet" replace />} />
        </Routes>
      </BrowserRouter>
    </MohThemeProvider>
  );
}
