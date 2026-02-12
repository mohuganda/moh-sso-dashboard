// src/components/header-panel/header-panel.context.tsx
import React, { createContext, useContext, useState } from "react";

import { ReusableHeaderPanel } from "./ReusableHeaderPanel";

/* ---------------------------------
 * Types
 * --------------------------------- */
export type PanelSize = "sm" | "md" | "lg";

type HeaderPanelState = {
  isOpen: boolean;
  title?: string;
  content?: React.ReactNode;
  size: PanelSize;
};

type OpenPanelOptions = {
  title?: string;
  content: React.ReactNode;
  size?: PanelSize;
};

type HeaderPanelContextType = {
  openPanel: (opts: OpenPanelOptions) => void;
  closePanel: () => void;
};

/* ---------------------------------
 * Context
 * --------------------------------- */
const HeaderPanelContext = createContext<HeaderPanelContextType | null>(null);

export function HeaderPanelProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<HeaderPanelState>({
    isOpen: false,
    size: "sm",
  });

  const openPanel = ({ title, content, size = "sm" }: OpenPanelOptions) => {
    setState({
      isOpen: true,
      title,
      content,
      size,
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
