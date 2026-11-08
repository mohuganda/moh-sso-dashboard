import React, { useEffect, useId, useRef, useState, type ReactNode } from "react";
import { Button, HeaderGlobalAction } from "@carbon/react";
import { Close, Switcher } from "@carbon/react/icons";

import { useFocusTrap } from "../../accessibility/useFocusTrap";
import "./AppMenu.scss";

type AppMenuActionProps = {
  children?: ReactNode | ((closeMenu: () => void) => ReactNode);
};

const AppMenuAction: React.FC<AppMenuActionProps> = ({ children }) => {
  const [expanded, setExpanded] = useState(false);

  const panelId = useId();
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);

  useFocusTrap({
    active: expanded,
    containerRef: panelRef,
    restoreFocus: true,
    initialFocusSelector: "input",
  });

  const closeMenu = () => {
    setExpanded(false);
  };

  const toggleMenu = () => {
    setExpanded((prev) => !prev);
  };

  useEffect(() => {
    if (!expanded) {
      return;
    }

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;

      const clickedInsidePanel = panelRef.current?.contains(target);
      const clickedTrigger = triggerRef.current?.contains(target);

      if (!clickedInsidePanel && !clickedTrigger) {
        closeMenu();
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMenu();
        triggerRef.current?.focus();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    // Search handles Escape itself; capture it first so the dialog still closes.
    document.addEventListener("keydown", handleKeyDown, true);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown, true);
    };
  }, [expanded]);

  return (
    <div className="app-menu">
      <HeaderGlobalAction
        ref={triggerRef}
        aria-label={expanded ? "Close application menu" : "Open application menu"}
        aria-haspopup="dialog"
        aria-expanded={expanded}
        aria-controls={panelId}
        isActive={expanded}
        onClick={toggleMenu}
      >
        <Switcher size={20} />
      </HeaderGlobalAction>

      {expanded && (
        <div
          ref={panelRef}
          id={panelId}
          className="app-menu-panel"
          role="dialog"
          aria-modal="false"
          aria-label="Application menu"
          tabIndex={-1}
        >
          <div className="app-menu-panel__header">
            <div>
              <h4 className="app-menu-panel__title">Applications</h4>
              <p className="app-menu-panel__subtitle">Find and open your available services.</p>
            </div>
            <Button
              kind="ghost"
              size="sm"
              hasIconOnly
              renderIcon={Close}
              iconDescription="Close application menu"
              onClick={() => {
                closeMenu();
                triggerRef.current?.focus();
              }}
              tooltipAlignment="end"
            />
          </div>

          {typeof children === "function" ? children(closeMenu) : children}
        </div>
      )}
    </div>
  );
};

export default AppMenuAction;
