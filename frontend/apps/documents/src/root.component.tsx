import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import DocumentDetailsPage from "./pages/document-management/document-details/document-details.component";
import FileUpload from "./pages/file-upload/file-upload.component";
import { MohThemeProvider } from "@moh-sso/ui";
import DocumentPage from "./pages/document-management/documents.component";

const DEFAULT_FILESVR_BASE = "/apps/dwh/filesvr";
const DEFAULT_DOCUMENT_UPLOAD_BASE = "/apps/utilities/self-service/eservice/document-upload";

function normalizePath(value?: string) {
  if (!value) return "";

  return value.startsWith("/") ? value : `/${value}`;
}

function resolveDocumentsBasename(props: MicrofrontendRuntimeProps) {
  const propBasename = normalizePath(props.basename);
  const pathname = window.location.pathname;

  if (propBasename) {
    return resolveRuntimeBasename(propBasename);
  }

  if (pathname.includes(DEFAULT_DOCUMENT_UPLOAD_BASE)) {
    return resolveRuntimeBasename(DEFAULT_DOCUMENT_UPLOAD_BASE);
  }

  return resolveRuntimeBasename(DEFAULT_FILESVR_BASE);
}

export function DocumentsRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveDocumentsBasename(props);

  const isDocumentUpload = basename.endsWith("/utilities/self-service/eservice/document-upload");

  return (
    <MohThemeProvider theme="white">
      <BrowserRouter basename={basename}>
        <Routes>
          {isDocumentUpload ? (
            <>
              <Route index element={<DocumentPage />} />
              <Route path=":id" element={<DocumentDetailsPage />} />
            </>
          ) : (
            <>
              <Route index element={<FileUpload />} />
              <Route path="*" element={<FileUpload />} />
            </>
          )}
        </Routes>
      </BrowserRouter>
    </MohThemeProvider>
  );
}
