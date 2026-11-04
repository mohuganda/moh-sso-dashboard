import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";

import { ComingSoon, MohThemeProvider, PageShell } from "@moh-sso/ui";

function UtilityPage({ title }: { title: string }) {
  return (
    <PageShell title={title} subtitle="Access your self-service tools and requests.">
      <ComingSoon title="Coming soon" />
    </PageShell>
  );
}

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

            <Route path="timesheet" element={<UtilityPage title="Timesheet" />} />

            <Route path="elearning" element={<UtilityPage title="E-learning" />} />

            <Route path="leave-plan" element={<UtilityPage title="Leave Plan" />} />

            <Route path="absence-requests" element={<UtilityPage title="Absence Requests" />} />

            <Route path="absence-dashboard" element={<UtilityPage title="Absence Dashboard" />} />

            <Route path="eservice">
              <Route index element={<Navigate to="service-access" replace />} />
              <Route path="service-access" element={<UtilityPage title="Service Access" />} />
              <Route path="equipment-request" element={<UtilityPage title="Equipment Request" />} />
            </Route>

            <Route path="*" element={<Navigate to="timesheet" replace />} />
          </Routes>
        </BrowserRouter>
      </MohThemeProvider>
    </div>
  );
}
