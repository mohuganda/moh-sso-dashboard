export type AppEvent = {
  type: string;
  source: string;
  payload?: unknown;
};

export type EventBus = {
  publish: (event: AppEvent) => void;
  subscribe: (type: string, handler: (event: AppEvent) => void) => () => void;
};

export type MicrofrontendRuntimeProps = {
  basename: string;
  auth?: {
    isAuthenticated: boolean;
    user?: unknown;
  };
  apiBaseUrl?: string;
  eventBus?: EventBus;
};

export type MicrofrontendMountProps = MicrofrontendRuntimeProps & {
  domElement: HTMLElement;
};
