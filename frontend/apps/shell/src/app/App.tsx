import { ModalProvider } from "@moh-sso/ui";
import AppRouter from "./routes/AppRouter";

function App() {
  return (
    <ModalProvider>
      <AppRouter />
    </ModalProvider>
  );
}

export default App;
