import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";

import {
  resolveRuntimeBasename,
  type MicrofrontendLifecycle,
  type MicrofrontendRuntimeProps,
} from "@moh-sso/microfrontend";
import { MicrofrontendErrorBoundary } from "@moh-sso/ui";
import { microfrontendContainerId } from "./containers";
import {
  isSingleSpaOrchestrationStarted,
  isSingleSpaOrchestrationUnavailable,
  MICROFRONTEND_ORCHESTRATION_EVENT,
  shouldUseSingleSpaOrchestration,
  startMicrofrontendOrchestration,
} from "./orchestrator";

type SingleSpaAppProps = MicrofrontendRuntimeProps & {
  appName: string;
  lifecycles: MicrofrontendLifecycle | (() => Promise<MicrofrontendLifecycle>);
};

export function SingleSpaApp({ appName, lifecycles, ...runtimeProps }: SingleSpaAppProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const location = useLocation();
  const { apiBaseUrl, auth, eventBus } = runtimeProps;
  const orchestrationRequested = shouldUseSingleSpaOrchestration();
  const [orchestrationState, setOrchestrationState] = useState(() => ({
    started: isSingleSpaOrchestrationStarted(),
    unavailable: isSingleSpaOrchestrationUnavailable(),
  }));
  const [mountError, setMountError] = useState<Error | null>(null);
  const orchestrated =
    orchestrationRequested && orchestrationState.started && !orchestrationState.unavailable;
  const shouldMountLocally = !orchestrationRequested || orchestrationState.unavailable;

  const basename = resolveRuntimeBasename(
    runtimeProps.basename || location.pathname.replace(/\/$/, ""),
    import.meta.env.BASE_URL,
  );

  useEffect(() => {
    const handleOrchestrationChange = () => {
      setOrchestrationState({
        started: isSingleSpaOrchestrationStarted(),
        unavailable: isSingleSpaOrchestrationUnavailable(),
      });
    };

    window.addEventListener(MICROFRONTEND_ORCHESTRATION_EVENT, handleOrchestrationChange);
    handleOrchestrationChange();

    return () => {
      window.removeEventListener(MICROFRONTEND_ORCHESTRATION_EVENT, handleOrchestrationChange);
    };
  }, []);

  useEffect(() => {
    if (
      orchestrationRequested &&
      !orchestrationState.started &&
      !orchestrationState.unavailable
    ) {
      void startMicrofrontendOrchestration();
    }
  }, [orchestrationRequested, orchestrationState.started, orchestrationState.unavailable]);

  useEffect(() => {
    if (!shouldMountLocally) {
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
        setMountError(null);
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
          setMountError(error instanceof Error ? error : new Error(String(error)));
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
  }, [apiBaseUrl, appName, auth, basename, eventBus, lifecycles, shouldMountLocally]);

  return (
    <MicrofrontendErrorBoundary appName={appName}>
      {mountError ? (
        <div role="alert" className="microfrontend-mount-error">
          Unable to load {appName}. {mountError.message}
        </div>
      ) : null}
      <div
        id={orchestrated || orchestrationRequested ? microfrontendContainerId(appName) : undefined}
        ref={containerRef}
        data-microfrontend={appName}
      />
    </MicrofrontendErrorBoundary>
  );
}
