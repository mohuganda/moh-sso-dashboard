import { Button } from "@carbon/react";
import { Close, Maximize, Minimize } from "@carbon/react/icons";
import { useEffect, useState } from "react";

import type { PanelSize } from "./header-panel.context";
import "./reusable-header-panel.css";

type Props = {
  isOpen: boolean;
  title?: string;
  content?: React.ReactNode;
  contentKey?: number;
  onClose: () => void;
  size: PanelSize;
  primaryActionLabel?: string;
  secondaryActionLabel?: string;
  onPrimaryAction?: () => void;
  onSecondaryAction?: () => void;
  disablePrimaryAction?: boolean;
  maximizable?: boolean;
  defaultMaximized?: boolean;
};

export function ReusableHeaderPanel({
  isOpen,
  title,
  content,
  contentKey,
  onClose,
  size,
  primaryActionLabel,
  secondaryActionLabel,
  onPrimaryAction,
  onSecondaryAction,
  disablePrimaryAction = false,
  maximizable = true,
  defaultMaximized = false,
}: Props) {
  const [isMaximized, setIsMaximized] = useState(defaultMaximized);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = originalOverflow;
    };
  }, [isOpen, onClose]);

  useEffect(() => {
    if (isOpen) {
      setIsMaximized(defaultMaximized);
    }
  }, [isOpen, defaultMaximized]);

  const handleToggleMaximize = () => {
    setIsMaximized((prev) => !prev);
  };

  return (
    <div className={`app-side-panel ${isOpen ? "app-side-panel--open" : ""}`} aria-hidden={!isOpen}>
      <button
        type="button"
        className="app-side-panel__backdrop"
        onClick={onClose}
        aria-label="Close panel backdrop"
      />

      <aside
        className={[
          "app-side-panel__drawer",
          `app-side-panel__drawer--${size}`,
          isMaximized ? "app-side-panel__drawer--maximized" : "",
        ].join(" ")}
        role="dialog"
        aria-modal="true"
        aria-label={title ?? "Panel"}
      >
        <div className="app-side-panel__header">
          <h4 className="app-side-panel__title">{title}</h4>

          <div className="app-side-panel__header-actions">
            {maximizable && (
              <Button
                kind="ghost"
                hasIconOnly
                renderIcon={isMaximized ? Minimize : Maximize}
                iconDescription={isMaximized ? "Restore panel size" : "Maximize panel"}
                size="md"
                onClick={handleToggleMaximize}
                tooltipAlignment="end"
              />
            )}

            <Button
              kind="ghost"
              hasIconOnly
              renderIcon={Close}
              iconDescription="Close panel"
              size="md"
              onClick={onClose}
              tooltipAlignment="end"
            />
          </div>
        </div>

        <div key={contentKey} className="app-side-panel__body">{content}</div>

        {(primaryActionLabel || secondaryActionLabel) && (
          <div className="app-side-panel__footer">
            {secondaryActionLabel && (
              <Button
                kind="secondary"
                onClick={onSecondaryAction ?? onClose}
                className="app-side-panel__footer-button"
              >
                {secondaryActionLabel}
              </Button>
            )}

            {primaryActionLabel && (
              <Button
                kind="primary"
                onClick={onPrimaryAction}
                disabled={disablePrimaryAction}
                className="app-side-panel__footer-button"
              >
                {primaryActionLabel}
              </Button>
            )}
          </div>
        )}
      </aside>
    </div>
  );
}
