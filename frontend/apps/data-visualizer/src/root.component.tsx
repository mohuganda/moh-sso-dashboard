import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import DataVisualizer from "./pages/data-visualizer";

export function DataVisualizerRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <DataVisualizer />
    </div>
  );
}
