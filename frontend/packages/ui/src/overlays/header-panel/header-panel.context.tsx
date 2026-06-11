import React, { createContext, useContext, useState } from "react";

import { ReusableHeaderPanel } from "./ReusableHeaderPanel";

export type PanelSize = "sm" | "md" | "lg";

type HeaderPanelState = {
  isOpen: boolean;
  title?: string;
  content?: React.ReactNode;
  size: PanelSize;
  primaryActionLabel?: string;
  secondaryActionLabel?: string;
  onPrimaryAction?: () => void;
  onSecondaryAction?: () => void;
  disablePrimaryAction?: boolean;
  maximizable?: boolean;
  defaultMaximized?: boolean;
};

type OpenPanelOptions = {
  title?: string;
  content: React.ReactNode;
  size?: PanelSize;
  primaryActionLabel?: string;
  secondaryActionLabel?: string;
  onPrimaryAction?: () => void;
  onSecondaryAction?: () => void;
  disablePrimaryAction?: boolean;
  maximizable?: boolean;
  defaultMaximized?: boolean;
};

type HeaderPanelContextType = {
  openPanel: (opts: OpenPanelOptions) => void;
  closePanel: () => void;
};

const HeaderPanelContext = createContext<HeaderPanelContextType | null>(null);

export function HeaderPanelProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<HeaderPanelState>({
    isOpen: false,
    size: "md",
  });

  const openPanel = ({
    title,
    content,
    size = "md",
    primaryActionLabel,
    secondaryActionLabel,
    onPrimaryAction,
    onSecondaryAction,
    disablePrimaryAction,
  }: OpenPanelOptions) => {
    setState({
      isOpen: true,
      title,
      content,
      size,
      primaryActionLabel,
      secondaryActionLabel,
      onPrimaryAction,
      onSecondaryAction,
      disablePrimaryAction,
    });
  };

  const closePanel = () => {
    setState((prev) => ({
      ...prev,
      isOpen: false,
    }));
  };

  return (
    <HeaderPanelContext.Provider value={{ openPanel, closePanel }}>
      {children}

      <ReusableHeaderPanel
        isOpen={state.isOpen}
        title={state.title}
        content={state.content}
        size={state.size}
        onClose={closePanel}
        primaryActionLabel={state.primaryActionLabel}
        secondaryActionLabel={state.secondaryActionLabel}
        onPrimaryAction={state.onPrimaryAction}
        onSecondaryAction={state.onSecondaryAction}
        disablePrimaryAction={state.disablePrimaryAction}
        maximizable
        defaultMaximized={false}
      />
    </HeaderPanelContext.Provider>
  );
}

export function useHeaderPanel() {
  const ctx = useContext(HeaderPanelContext);
  if (!ctx) {
    throw new Error("useHeaderPanel must be used within HeaderPanelProvider");
  }
  return ctx;
}
