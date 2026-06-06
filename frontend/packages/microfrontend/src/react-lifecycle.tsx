import type { ComponentType, ReactNode } from "react";
import ReactDOMClient, { type Root } from "react-dom/client";
import { Provider } from "react-redux";

import { store } from "@moh-sso/state";
import { HeaderPanelProvider , ModalProvider , ToastProvider } from "@moh-sso/ui";

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

  return tree;
}

export function createReactMicrofrontendLifecycle(
  RootComponent: ComponentType<MicrofrontendRuntimeProps>,
  lifecycleOptions: ReactLifecycleOptions = {},
): MicrofrontendLifecycle {
  const options = { ...defaultOptions, ...lifecycleOptions };
  let root: Root | null = null;

  return {
    async bootstrap() {
      return;
    },

    async mount({ domElement, ...props }: MicrofrontendMountProps) {
      root = ReactDOMClient.createRoot(domElement);
      root.render(withProviders(<RootComponent {...props} />, options));
    },

    async unmount() {
      root?.unmount();
      root = null;
    },
  };
}
