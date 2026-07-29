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
  healthContext?: {
    activeContext?: {
      id: string;
      code: string;
      name: string;
      contextType: string;
    };
    contexts: Array<{
      id: string;
      code: string;
      name: string;
      contextType: string;
      scopeMode: string;
      assignmentType: string;
    }>;
    selectContext?: (contextId: string) => Promise<void>;
  };
};

export type MicrofrontendMountProps = MicrofrontendRuntimeProps & {
  domElement: HTMLElement;
};

declare global {
  interface Window {
    __MOH_SSO_AUTH__?: MicrofrontendRuntimeProps["auth"];
    __MOH_SSO_HEALTH_CONTEXT__?: MicrofrontendRuntimeProps["healthContext"];
  }
}
