import { useEffect, type RefObject } from "react";

type FocusTrapOptions = {
  active: boolean;
  containerRef: RefObject<HTMLElement | null>;
  restoreFocus?: boolean;
  initialFocusRef?: RefObject<HTMLElement | HTMLButtonElement | null>;
};

const FOCUSABLE_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "textarea:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "[tabindex]:not([tabindex='-1'])",
].join(",");

function getFocusableElements(container: HTMLElement) {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (element) => !element.hasAttribute("disabled") && element.getAttribute("aria-hidden") !== "true",
  );
}

export function useFocusTrap({
  active,
  containerRef,
  restoreFocus = true,
  initialFocusRef,
}: FocusTrapOptions) {
  useEffect(() => {
    if (!active) {
      return;
    }

    const container = containerRef.current;
    const previouslyFocused = document.activeElement instanceof HTMLElement ? document.activeElement : null;

    if (!container) {
      return;
    }

    const focusInitialElement = () => {
      const initialElement = initialFocusRef?.current;

      if (initialElement) {
        initialElement.focus();
        return;
      }

      const focusableElements = getFocusableElements(container);
      (focusableElements[0] ?? container).focus();
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Tab") {
        return;
      }

      const focusableElements = getFocusableElements(container);

      if (focusableElements.length === 0) {
        event.preventDefault();
        container.focus();
        return;
      }

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];
      const activeElement = document.activeElement;

      if (event.shiftKey && activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
        return;
      }

      if (!event.shiftKey && activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    };

    requestAnimationFrame(focusInitialElement);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);

      if (restoreFocus && previouslyFocused?.isConnected) {
        previouslyFocused.focus();
      }
    };
  }, [active, containerRef, initialFocusRef, restoreFocus]);
}
