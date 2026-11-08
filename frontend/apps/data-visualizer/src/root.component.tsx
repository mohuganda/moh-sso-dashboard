import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { PageShell } from "@moh-sso/ui";

import DataVisualizer from "./pages/data-visualizer";

export function DataVisualizerRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <PageShell
        title="Data Visualizer"
        subtitle="Explore data by period and organisation unit, build charts, and export results."
      >
        <DataVisualizer />
      </PageShell>
    </div>
  );
}
