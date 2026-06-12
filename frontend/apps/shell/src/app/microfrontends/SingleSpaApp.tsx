import { useEffect, useRef } from "react";
import { useLocation } from "react-router-dom";

import {
  resolveRuntimeBasename,
  type MicrofrontendLifecycle,
  type MicrofrontendRuntimeProps,
} from "@moh-sso/microfrontend";
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

  const basename = resolveRuntimeBasename(
    runtimeProps.basename || location.pathname.replace(/\/$/, ""),
    import.meta.env.BASE_URL,
  );

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
    let mountedLifecycles: MicrofrontendLifecycle | null = null;

    const loadAndMount = async () => {
      try {
        const resolvedLifecycles =
          typeof lifecycles === "function" ? await lifecycles() : lifecycles;
        mountedLifecycles = resolvedLifecycles;

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
      if (mountedLifecycles) {
        void Promise.resolve(mountedLifecycles.unmount(props));
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
