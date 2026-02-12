import { Switcher } from "@carbon/react/icons";
import { HeaderGlobalAction } from "@carbon/react";
import React, { useState, useRef, useEffect } from "react";

import AppGridContent from "./AppGridContent";
import "./AppMenu.css";

const AppMenuAction: React.FC = () => {
  const [expanded, setExpanded] = useState(false);

  // 🔥 Change ref to DIV
  const triggerRef = useRef<HTMLDivElement | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);

  const closeMenu = () => {
    setExpanded(false);
  };

  /* -----------------------------
   * Click outside & ESC to close
   * ----------------------------- */
  useEffect(() => {
    if (!expanded) return;

    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node;

      if (
        panelRef.current &&
        !panelRef.current.contains(target) &&
        triggerRef.current &&
        !triggerRef.current.contains(target)
      ) {
        closeMenu();
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeMenu();
        triggerRef.current?.focus();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [expanded]);

  /* -----------------------------
   * Focus management
   * ----------------------------- */
  useEffect(() => {
    if (expanded) {
      panelRef.current?.focus();
    }
  }, [expanded]);

  return (
    <>
      <div ref={triggerRef}>
        <HeaderGlobalAction
          aria-label="Open application menu"
          aria-haspopup="dialog"
          aria-expanded={expanded}
          isActive={expanded}
          onClick={() => {
            setExpanded((prev) => !prev);
          }}
        >
          <Switcher size={20} />
        </HeaderGlobalAction>
      </div>

      {expanded && (
        <div
          ref={panelRef}
          id="app_menu_container"
          role="dialog"
          aria-label="Application menu"
          tabIndex={-1}
        >
          <AppGridContent />
        </div>
      )}
    </>
  );
};

export default AppMenuAction;
