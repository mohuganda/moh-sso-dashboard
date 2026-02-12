import { useToastContext } from "./ToastProvider";

/* -----------------------------
 * Toast types
 * ----------------------------- */
export type ToastKind = "success" | "error" | "info" | "warning";

export type ToastAction = {
  label: string;
  onClick: () => void;
};

export type ToastOptions = {
  title: string;
  subtitle?: string;
  kind?: ToastKind;
  timeout?: number;
  dismissible?: boolean;
  actions?: ToastAction[];
};

/* -----------------------------
 * Defaults
 * ----------------------------- */
const DEFAULT_TIMEOUT = 5000;

/* -----------------------------
 * Hook
 * ----------------------------- */
export function useToast() {
  const { push } = useToastContext();

  /**
   * Main show function that bridges the options to the context
   */
  const show = (options: ToastOptions) => {
    push({
      kind: options.kind ?? "info",
      title: options.title,
      subtitle: options.subtitle,
      timeout: options.timeout ?? DEFAULT_TIMEOUT,
      dismissible: options.dismissible ?? true,
      actions: options.actions,
    });
  };

  return {
    show,

    success: (options: Omit<ToastOptions, "kind"> | string, subtitle?: string) => {
      if (typeof options === "string") {
        show({ kind: "success", title: options, subtitle });
      } else {
        show({ ...options, kind: "success" });
      }
    },

    error: (options: Omit<ToastOptions, "kind"> | string, subtitle?: string) => {
      if (typeof options === "string") {
        show({ kind: "error", title: options, subtitle });
      } else {
        show({ ...options, kind: "error" });
      }
    },

    info: (options: Omit<ToastOptions, "kind"> | string, subtitle?: string) => {
      if (typeof options === "string") {
        show({ kind: "info", title: options, subtitle });
      } else {
        show({ ...options, kind: "info" });
      }
    },

    warning: (options: Omit<ToastOptions, "kind"> | string, subtitle?: string) => {
      if (typeof options === "string") {
        show({ kind: "warning", title: options, subtitle });
      } else {
        show({ ...options, kind: "warning" });
      }
    },
  };
}
