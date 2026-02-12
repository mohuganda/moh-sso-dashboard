// src/components/header-panel/ReusableHeaderPanel.tsx
import { HeaderPanel, Button } from "@carbon/react";

import type { PanelSize } from "./header-panel.context";
import "./reusable-header-panel.css";

type Props = {
  isOpen: boolean;
  title?: string;
  content?: React.ReactNode;
  onClose: () => void;
  size: PanelSize;
};

export function ReusableHeaderPanel({ isOpen, title, content, onClose, size }: Props) {
  return (
    <HeaderPanel
      expanded={isOpen}
      aria-label={title ?? "Panel"}
      className={`app-header-panel app-header-panel--${size}`}
    >
      <div className="header-panel__header">
        <h4>{title}</h4>
        <Button kind="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </div>

      <div className="header-panel__content">{content}</div>
    </HeaderPanel>
  );
}
