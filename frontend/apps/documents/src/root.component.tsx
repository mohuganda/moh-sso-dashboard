import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import DocumentDetailsPage from "./pages/document-management/document-details/document-details.component";
import DocumentPage from "./pages/document-management/documents.component";
import FileUpload from "./pages/file-upload/file-upload.component";

const DEFAULT_FILESVR_BASE = "/portal/apps/dwh/filesvr";
const DEFAULT_DOCUMENT_UPLOAD_BASE = "/portal/apps/utilities/self-service/eservice/document-upload";

function normalizePath(value?: string) {
  if (!value) return "";

  return value.startsWith("/") ? value : `/${value}`;
}

function resolveDocumentsBasename(props: MicrofrontendRuntimeProps) {
  const propBasename = normalizePath(props.basename);
  const pathname = window.location.pathname;

  if (propBasename) {
    return propBasename;
  }

  if (pathname.startsWith(DEFAULT_DOCUMENT_UPLOAD_BASE)) {
    return DEFAULT_DOCUMENT_UPLOAD_BASE;
  }

  return DEFAULT_FILESVR_BASE;
}

export function DocumentsRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveDocumentsBasename(props);

  const isDocumentUpload = basename.endsWith("/utilities/self-service/eservice/document-upload");

  return (
    <BrowserRouter basename={basename}>
      <Routes>
        {!isDocumentUpload ? (
          <>
            <Route index element={<FileUpload />} />
            <Route path=":id" element={<DocumentDetailsPage />} />
          </>
        ) : (
          <>
            <Route index element={<DocumentPage />} />
            <Route path=":id" element={<DocumentDetailsPage />} />
          </>
        )}
      </Routes>
    </BrowserRouter>
  );
}
