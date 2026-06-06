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
import { registeredMicrofrontends } from "./registerApps";

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
  "@moh-sso/announcements": () => Promise.resolve(announcementsLifecycles),
  "@moh-sso/audit": () => Promise.resolve(auditLifecycles),
  "@moh-sso/clients": () => Promise.resolve(clientsLifecycles),
  "@moh-sso/data-visualizer": () => Promise.resolve(dataVisualizerLifecycles),
  "@moh-sso/documents": () => Promise.resolve(documentsLifecycles),
  "@moh-sso/e-services": () => Promise.resolve(eServicesLifecycles),
  "@moh-sso/email": () => Promise.resolve(emailLifecycles),
  "@moh-sso/issue-tracker": () => Promise.resolve(issueTrackerLifecycles),
  "@moh-sso/report-browser": () => Promise.resolve(reportBrowserLifecycles),
  "@moh-sso/surveillance": () => Promise.resolve(surveillanceLifecycles),
  "@moh-sso/users": () => Promise.resolve(usersLifecycles),
  "@moh-sso/utilities": () => Promise.resolve(utilitiesLifecycles),
};

function shouldUseSingleSpaOrchestration() {
  const runtimeConfig = (window as Window & { __APP_CONFIG__?: { singleSpaOrchestration?: boolean } })
    .__APP_CONFIG__;
  return import.meta.env.VITE_SINGLE_SPA_ORCHESTRATION === "true" || runtimeConfig?.singleSpaOrchestration === true;
}

async function loadSingleSpa(): Promise<SingleSpaModule | null> {
  try {
    const runtimeImport = new Function("specifier", "return import(specifier)") as (
      specifier: string,
    ) => Promise<SingleSpaModule>;
    return await runtimeImport("single-spa");
  } catch {
    return null;
  }
}

function registerRoute(singleSpa: SingleSpaModule, route: MicrofrontendRoute) {
  const loader = lifecycleLoaders[route.appName];

  if (!loader) {
    return;
  }

  singleSpa.registerApplication({
    name: route.appName,
    app: loader,
    activeWhen: [route.path],
    customProps: {
      basename: route.path,
    },
  });
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
