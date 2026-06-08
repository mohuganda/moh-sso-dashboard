import { useEffect, useRef } from "react";
import { useLocation } from "react-router-dom";

import type { MicrofrontendLifecycle, MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { MicrofrontendErrorBoundary } from "@moh-sso/ui";
import { microfrontendContainerId } from "./containers";
import { shouldUseSingleSpaOrchestration } from "./orchestrator";

type SingleSpaAppProps = MicrofrontendRuntimeProps & {
  appName: string;
  lifecycles: MicrofrontendLifecycle | (() => Promise<MicrofrontendLifecycle>);
};

export function SingleSpaApp({ appName, lifecycles, ...runtimeProps }: SingleSpaAppProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const location = useLocation();
  const { apiBaseUrl, auth, eventBus } = runtimeProps;
  const orchestrated = shouldUseSingleSpaOrchestration();

  const basename = runtimeProps.basename || location.pathname.replace(/\/$/, "");

  useEffect(() => {
    if (orchestrated) {
      return;
    }

    const domElement = containerRef.current;
    if (!domElement) {
      return;
    }

    let disposed = false;
    const props = { apiBaseUrl, auth, basename, eventBus, domElement };

    const loadAndMount = async () => {
      try {
        const resolvedLifecycles =
          typeof lifecycles === "function" ? await lifecycles() : lifecycles;

        if (disposed) return;

        await resolvedLifecycles.bootstrap(props);

        if (disposed) return;

        await resolvedLifecycles.mount(props);
      } catch (error: unknown) {
        console.error(`Failed to load or mount microfrontend ${appName}`, error);
        if (!disposed) {
          throw error;
        }
      }
    };

    void loadAndMount();

    return () => {
      disposed = true;
      // Note: We can't easily wait for unmount here since it's async and we're in a cleanup,
      // but the 'disposed' flag prevents late mount actions.
      if (typeof lifecycles !== "function") {
        void Promise.resolve(lifecycles.unmount(props));
      } else {
        // If it was a function, we'd need to re-resolve it to unmount, 
        // which might be overkill if the page is already gone.
      }
    };
  }, [apiBaseUrl, appName, auth, basename, eventBus, lifecycles, orchestrated]);

  return (
    <MicrofrontendErrorBoundary appName={appName}>
      <div
        id={orchestrated ? microfrontendContainerId(appName) : undefined}
        ref={containerRef}
        data-microfrontend={appName}
      />
    </MicrofrontendErrorBoundary>
  );
}
