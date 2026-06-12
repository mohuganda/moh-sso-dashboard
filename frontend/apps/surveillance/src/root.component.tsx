import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import DiseaseDetailsPage from "./pages/surveillance-details/surveillance-details.component";
import SurveillanceDashboardPage from "./pages/surveillance.component";
import { MohThemeProvider } from "@moh-sso/ui";

export function SurveillanceRoot(props: MicrofrontendRuntimeProps) {
  return (
    <MohThemeProvider theme="white">
      <BrowserRouter basename={resolveRuntimeBasename(props.basename || "/apps/dwh/surveillance")}>
        <Routes>
          <Route index element={<SurveillanceDashboardPage />} />
          <Route path=":diseaseName" element={<DiseaseDetailsPage />} />
          <Route path="*" element={<SurveillanceDashboardPage />} />
        </Routes>
      </BrowserRouter>
    </MohThemeProvider>
  );
}
