import ReactDOM from "react-dom/client";
import { Provider } from "react-redux";

import App from "./App";
import { store } from "./store";
import "leaflet/dist/leaflet.css";

import "./index.css";
import AuthBootstrap from "./store/auth/AuthBootstrap";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <Provider store={store}>
    <AuthBootstrap />
    <App />
  </Provider>,
);
