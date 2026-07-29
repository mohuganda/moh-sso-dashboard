import { ModalProvider, MohThemeProvider, ToastProvider } from "@moh-sso/ui";
import { HealthContextProvider } from "@moh-sso/auth";
import { PwaLifecycle } from "./pwa/PwaLifecycle";
import AppRouter from "./routes/AppRouter";

function App() {
  return (
    <MohThemeProvider theme="white">
      <ToastProvider>
        <ModalProvider>
          <HealthContextProvider>
            <PwaLifecycle />
            <AppRouter />
          </HealthContextProvider>
        </ModalProvider>
      </ToastProvider>
    </MohThemeProvider>
  );
}

export default App;
