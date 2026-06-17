import { ModalProvider, MohThemeProvider } from "@moh-sso/ui";
import AppRouter from "./routes/AppRouter";

function App() {
  return (
    <MohThemeProvider theme="white">
      <ModalProvider>
        <AppRouter />
      </ModalProvider>
    </MohThemeProvider>
  );
}

export default App;
