import { ToastNotification } from "@carbon/react";
import { createContext, useContext, useState, type ReactNode } from "react";
import type { ToastAction } from "./useToast";

import "./ToastProvider.scss";

export type ToastKind = "success" | "error" | "info" | "warning";

export type Toast = {
  id: string;
  kind: ToastKind;
  title: string;
  subtitle?: string;
  timeout: number; // Added this
  dismissible: boolean; // Added this
  actions?: ToastAction[];
};

type ToastContextType = {
  push: (toast: Omit<Toast, "id">) => void;
};

const ToastContext = createContext<ToastContextType | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = (toast: Omit<Toast, "id">) => {
    const id = crypto.randomUUID();
    setToasts((t) => [...t, { ...toast, id }]);

    // Uses the custom timeout passed from the hook
    setTimeout(() => {
      setToasts((t) => t.filter((x) => x.id !== id));
    }, toast.timeout);
  };

  const remove = (id: string) => {
    setToasts((t) => t.filter((x) => x.id !== id));
  };

  return (
    <ToastContext.Provider value={{ push }}>
      {children}

      <div className="moh-toast-stack">
        {toasts.map((t) => (
          <ToastNotification
            key={t.id}
            kind={t.kind}
            title={t.title}
            subtitle={t.subtitle}
            // Allow manual dismissal if dismissible is true
            onCloseButtonClick={() => remove(t.id)}
            hideCloseButton={!t.dismissible}
            lowContrast
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToastContext() {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error("useToastContext must be used inside ToastProvider");
  }
  return ctx;
}
