import { ModalProvider, MohThemeProvider, ToastProvider } from "@moh-sso/ui";
import { PwaLifecycle } from "./pwa/PwaLifecycle";
import AppRouter from "./routes/AppRouter";

function App() {
  return (
    <MohThemeProvider theme="white">
      <ToastProvider>
        <ModalProvider>
          <PwaLifecycle />
          <AppRouter />
        </ModalProvider>
      </ToastProvider>
    </MohThemeProvider>
  );
}

export default App;
