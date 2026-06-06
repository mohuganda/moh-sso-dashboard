import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import DocumentDetailsPage from "./pages/document-details.component";
import DocumentPage from "./pages/documents.component";
import FileUpload from "./pages/file-upload";

export function DocumentsRoot(props: MicrofrontendRuntimeProps) {
  const documentUploadBase = "/apps/utilities/self-service/eservice/document-upload";
  const basename = window.location.pathname.startsWith(documentUploadBase)
    ? documentUploadBase
    : props.basename || "/apps/dwh/filesvr";

  if (basename === "/apps/dwh/filesvr") {
    return (
      <BrowserRouter basename={basename}>
        <Routes>
          <Route index element={<FileUpload />} />
        </Routes>
      </BrowserRouter>
    );
  }

  return (
    <BrowserRouter basename={basename}>
      <Routes>
        <Route index element={<DocumentPage />} />
        <Route path=":id" element={<DocumentDetailsPage />} />
      </Routes>
    </BrowserRouter>
  );
}
