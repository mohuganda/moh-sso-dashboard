import { ModalProvider } from "@/shared/components/modal/modal.context";
import AppRouter from "./routes/AppRouter";

function App() {
  return (
    <ModalProvider>
      <AppRouter />
    </ModalProvider>
  );
}

export default App;
