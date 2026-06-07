import { useEffect, useRef } from "react";

import type { MicrofrontendLifecycle, MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";
import { microfrontendContainerId } from "./containers";
import { shouldUseSingleSpaOrchestration } from "./orchestrator";

type SingleSpaAppProps = MicrofrontendRuntimeProps & {
  appName: string;
  lifecycles: MicrofrontendLifecycle;
};

export function SingleSpaApp({ appName, lifecycles, ...runtimeProps }: SingleSpaAppProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const { apiBaseUrl, auth, basename, eventBus } = runtimeProps;
  const orchestrated = shouldUseSingleSpaOrchestration();

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

    void Promise.resolve(lifecycles.bootstrap(props))
      .then(() => {
        if (!disposed) {
          return lifecycles.mount(props);
        }
      })
      .catch((error: unknown) => {
        domElement.textContent = `Unable to load ${appName}`;
        console.error(`Failed to mount microfrontend ${appName}`, error);
      });

    return () => {
      disposed = true;
      void Promise.resolve(lifecycles.unmount(props));
    };
  }, [apiBaseUrl, appName, auth, basename, eventBus, lifecycles, orchestrated]);

  return (
    <div
      id={orchestrated ? microfrontendContainerId(appName) : undefined}
      ref={containerRef}
      data-microfrontend={appName}
    />
  );
}
