import React, { createContext, useContext, useState } from "react";

import { type ReusableModalProps, ReusableModal } from "./ReusableModal";

type ModalState = Omit<ReusableModalProps, "open"> & {
  open: boolean;
};

type ModalContextType = {
  openModal: (config: Omit<ModalState, "open">) => void;
  closeModal: () => void;
};

const ModalContext = createContext<ModalContextType | null>(null);

export function ModalProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<ModalState>({
    open: false,
    content: null,
    onClose: () => {},
  });

  const openModal = (config: Omit<ModalState, "open">) => {
    setState({ ...config, open: true });
  };

  const closeModal = () => {
    setState((prev) => ({ ...prev, open: false }));
  };

  return (
    <ModalContext.Provider value={{ openModal, closeModal }}>
      {children}

      <ReusableModal
        open={state.open}
        title={state.title}
        content={state.content}
        primaryAction={state.primaryAction}
        secondaryAction={state.secondaryAction}
        size={state.size}
        onClose={() => {
          state.onClose?.();
          closeModal();
        }}
      />
    </ModalContext.Provider>
  );
}

export function useModal() {
  const ctx = useContext(ModalContext);
  if (!ctx) {
    throw new Error("useModal must be used inside ModalProvider");
  }
  return ctx;
}
