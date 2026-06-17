type ShellRuntimeConfig = {
  singleSpaOrchestration?: boolean;
  microfrontendMode?: "local" | "remote";
  microfrontendMountMode?: "hybrid" | "orchestrated";
};

export const APP_CONFIG = (
  window as Window & {
    __APP_CONFIG__?: ShellRuntimeConfig;
  }
).__APP_CONFIG__;
