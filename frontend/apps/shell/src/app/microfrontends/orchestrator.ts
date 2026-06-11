import type { MicrofrontendLifecycle, MicrofrontendRoute } from "@moh-sso/microfrontend";

import {
  announcementsLifecycles,
  auditLifecycles,
  clientsLifecycles,
  dataVisualizerLifecycles,
  documentsLifecycles,
  eServicesLifecycles,
  emailLifecycles,
  issueTrackerLifecycles,
  reportBrowserLifecycles,
  surveillanceLifecycles,
  usersLifecycles,
  utilitiesLifecycles,
} from "./lifecycles";
import { getMicrofrontendContainer } from "./containers";
import { registeredMicrofrontends } from "./registerApps";

type MicrofrontendMountMode = "hybrid" | "orchestrated";

type RuntimeMicrofrontendConfig = {
  singleSpaOrchestration?: boolean;
  microfrontendMode?: "local" | "remote";
  microfrontendMountMode?: MicrofrontendMountMode;
};

type SingleSpaModule = {
  registerApplication: (config: {
    name: string;
    app: () => Promise<MicrofrontendLifecycle>;
    activeWhen: string[];
    customProps: Record<string, unknown>;
  }) => void;
  start: () => void;
};

const lifecycleLoaders: Record<string, () => Promise<MicrofrontendLifecycle>> = {
  "@moh-sso/announcements": announcementsLifecycles,
  "@moh-sso/audit": auditLifecycles,
  "@moh-sso/clients": clientsLifecycles,
  "@moh-sso/data-visualizer": dataVisualizerLifecycles,
  "@moh-sso/documents": documentsLifecycles,
  "@moh-sso/e-services": eServicesLifecycles,
  "@moh-sso/email": emailLifecycles,
  "@moh-sso/issue-tracker": issueTrackerLifecycles,
  "@moh-sso/report-browser": reportBrowserLifecycles,
  "@moh-sso/surveillance": surveillanceLifecycles,
  "@moh-sso/users": usersLifecycles,
  "@moh-sso/utilities": utilitiesLifecycles,
};

export function shouldUseSingleSpaOrchestration() {
  const runtimeConfig = (
    window as Window & {
      __APP_CONFIG__?: RuntimeMicrofrontendConfig;
    }
  ).__APP_CONFIG__;
  const orchestrationEnabled =
    import.meta.env.VITE_SINGLE_SPA_ORCHESTRATION === "true" || runtimeConfig?.singleSpaOrchestration === true;

  return orchestrationEnabled && getMicrofrontendMountMode() === "orchestrated";
}

function getMicrofrontendMode() {
  const runtimeConfig = (
    window as Window & {
      __APP_CONFIG__?: RuntimeMicrofrontendConfig;
    }
  ).__APP_CONFIG__;
  return runtimeConfig?.microfrontendMode ?? import.meta.env.VITE_MICROFRONTEND_MODE ?? "local";
}

function getMicrofrontendMountMode(): MicrofrontendMountMode {
  const runtimeConfig = (
    window as Window & {
      __APP_CONFIG__?: RuntimeMicrofrontendConfig;
    }
  ).__APP_CONFIG__;
  return runtimeConfig?.microfrontendMountMode ?? import.meta.env.VITE_MICROFRONTEND_MOUNT_MODE ?? "hybrid";
}

async function loadSingleSpa(): Promise<SingleSpaModule | null> {
  try {
    return await runtimeImport("single-spa");
  } catch {
    return null;
  }
}

function registerRoute(singleSpa: SingleSpaModule, route: MicrofrontendRoute) {
  const mode = getMicrofrontendMode();
  const loader =
    mode === "remote"
      ? () => runtimeImport<MicrofrontendLifecycle>(route.appName)
      : lifecycleLoaders[route.appName];

  if (!loader) {
    return;
  }

  singleSpa.registerApplication({
    name: route.appName,
    app: async () => {
      const lifecycle = await loader();
      const propsWithContainer = (props: Record<string, unknown>) => ({
        ...props,
        basename: route.path,
        domElement: getMicrofrontendContainer(route.appName),
      });

      return {
        bootstrap: (props) => lifecycle.bootstrap(propsWithContainer(props)),
        mount: (props) => lifecycle.mount(propsWithContainer(props)),
        unmount: (props) => lifecycle.unmount(propsWithContainer(props)),
      };
    },
    activeWhen: route.paths ?? [route.path],
    customProps: {
      basename: route.path,
    },
  });
}

function runtimeImport<T>(specifier: string): Promise<T> {
  const load = new Function("specifier", "return import(specifier)") as (
    specifier: string,
  ) => Promise<T>;
  return load(specifier);
}

export async function startMicrofrontendOrchestration() {
  if (!shouldUseSingleSpaOrchestration()) {
    return false;
  }

  const singleSpa = await loadSingleSpa();

  if (!singleSpa) {
    return false;
  }

  registeredMicrofrontends.forEach((route) => registerRoute(singleSpa, route));
  singleSpa.start();
  return true;
}
