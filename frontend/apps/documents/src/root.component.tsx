import { resolveRuntimeBasename, type MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import DocumentDetailsPage from "./pages/document-management/document-details/document-details.component";
import { MohThemeProvider } from "@moh-sso/ui";
import DocumentPage from "./pages/document-management/documents.component";

const DEFAULT_DOCUMENTS_BASE = "/apps/dwh/documents";
const LEGACY_FILESVR_BASE = "/apps/dwh/filesvr";
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
  if (pathname.includes(LEGACY_FILESVR_BASE)) {
    return resolveRuntimeBasename(LEGACY_FILESVR_BASE);
  }
  return resolveRuntimeBasename(DEFAULT_DOCUMENTS_BASE);
}
export function DocumentsRoot(props: MicrofrontendRuntimeProps) {
  const basename = resolveDocumentsBasename(props);
  return (
    <div className="moh-microfrontend-root">
      <MohThemeProvider theme="white">
        <BrowserRouter basename={basename}>
          <Routes>
            <Route index element={<DocumentPage />} />
            <Route path=":id" element={<DocumentDetailsPage />} />
          </Routes>
        </BrowserRouter>
      </MohThemeProvider>
    </div>
  );
}
