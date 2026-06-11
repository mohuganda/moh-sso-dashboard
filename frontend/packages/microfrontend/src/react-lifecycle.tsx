import type { ComponentType, ReactNode } from "react";
import React from "react";
import ReactDOMClient, { type Root } from "react-dom/client";
import { Provider } from "react-redux";

import { store } from "@moh-sso/state";
import { HeaderPanelProvider, ModalProvider, MohThemeProvider, ToastProvider } from "@moh-sso/ui";

import type { MicrofrontendLifecycle } from "./lifecycle";
import type { MicrofrontendMountProps, MicrofrontendRuntimeProps } from "./props";

type ReactLifecycleOptions = {
  withRedux?: boolean;
  withModal?: boolean;
  withToast?: boolean;
  withHeaderPanel?: boolean;
};

const defaultOptions: Required<ReactLifecycleOptions> = {
  withRedux: true,
  withModal: true,
  withToast: true,
  withHeaderPanel: true,
};

type SingleSpaReactFactory = (options: {
  React: typeof React;
  ReactDOMClient: typeof ReactDOMClient;
  rootComponent: ComponentType<MicrofrontendMountProps>;
  domElementGetter: (props: MicrofrontendMountProps) => HTMLElement;
}) => MicrofrontendLifecycle;

function withProviders(children: ReactNode, options: Required<ReactLifecycleOptions>) {
  let tree = children;

  if (options.withModal) {
    tree = <ModalProvider>{tree}</ModalProvider>;
  }

  if (options.withHeaderPanel) {
    tree = <HeaderPanelProvider>{tree}</HeaderPanelProvider>;
  }

  if (options.withToast) {
    tree = <ToastProvider>{tree}</ToastProvider>;
  }

  if (options.withRedux) {
    tree = <Provider store={store}>{tree}</Provider>;
  }

  return <MohThemeProvider theme="white">{tree}</MohThemeProvider>;
}

function renderRoot(
  RootComponent: ComponentType<MicrofrontendRuntimeProps>,
  props: MicrofrontendMountProps,
  options: Required<ReactLifecycleOptions>,
) {
  const { domElement: _domElement, ...runtimeProps } = props;
  return withProviders(<RootComponent {...runtimeProps} />, options);
}

function createManualLifecycle(
  RootComponent: ComponentType<MicrofrontendRuntimeProps>,
  options: Required<ReactLifecycleOptions>,
): MicrofrontendLifecycle {
  let root: Root | null = null;

  return {
    async bootstrap() {
      return undefined;
    },
    async mount(props) {
      root = ReactDOMClient.createRoot(props.domElement);
      root.render(renderRoot(RootComponent, props, options));
    },
    async unmount() {
      root?.unmount();
      root = null;
    },
  };
}

async function loadSingleSpaReact(): Promise<SingleSpaReactFactory | null> {
  try {
    const runtimeImport = new Function("specifier", "return import(specifier)") as (
      specifier: string,
    ) => Promise<{ default?: SingleSpaReactFactory }>;
    const module = await runtimeImport("single-spa-react");
    return module.default ?? null;
  } catch {
    return null;
  }
}

export function createReactMicrofrontendLifecycle(
  RootComponent: ComponentType<MicrofrontendRuntimeProps>,
  lifecycleOptions: ReactLifecycleOptions = {},
): MicrofrontendLifecycle {
  const options = { ...defaultOptions, ...lifecycleOptions };
  const manualLifecycle = createManualLifecycle(RootComponent, options);
  let lifecycle: MicrofrontendLifecycle | null = null;

  async function getLifecycle() {
    if (lifecycle) {
      return lifecycle;
    }

    const singleSpaReact = await loadSingleSpaReact();
    lifecycle = singleSpaReact
      ? singleSpaReact({
          React,
          ReactDOMClient,
          rootComponent: (props) => renderRoot(RootComponent, props, options),
          domElementGetter: ({ domElement }) => domElement,
        })
      : manualLifecycle;

    return lifecycle;
  }

  return {
    async bootstrap(props) {
      const currentLifecycle = await getLifecycle();
      return currentLifecycle.bootstrap(props);
    },
    async mount(props) {
      const currentLifecycle = await getLifecycle();
      return currentLifecycle.mount(props);
    },
    async unmount(props) {
      const currentLifecycle = lifecycle ?? manualLifecycle;
      return currentLifecycle.unmount(props);
    },
  };
}
