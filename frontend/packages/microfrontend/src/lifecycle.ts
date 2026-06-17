import type { MicrofrontendMountProps } from "./props";

export type MicrofrontendLifecycle = {
  bootstrap: (props: MicrofrontendMountProps) => Promise<void> | void;
  mount: (props: MicrofrontendMountProps) => Promise<void> | void;
  unmount: (props: MicrofrontendMountProps) => Promise<void> | void;
};
