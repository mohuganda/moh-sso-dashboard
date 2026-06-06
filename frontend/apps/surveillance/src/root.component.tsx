import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import DiseaseDetailsPage from "./pages/surveillance-details/surveillance-details.component";
import SurveillanceDashboardPage from "./pages/surveillance.component";

export function SurveillanceRoot(props: MicrofrontendRuntimeProps) {
  return (
    <BrowserRouter basename={props.basename || "/apps/dwh/surveillance"}>
      <Routes>
        <Route index element={<SurveillanceDashboardPage />} />
        <Route path=":diseaseName" element={<DiseaseDetailsPage />} />
      </Routes>
    </BrowserRouter>
  );
}
