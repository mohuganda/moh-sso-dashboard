import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import FileUpload from "./pages/file-upload";

export function DocumentsRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return <FileUpload />;
}
