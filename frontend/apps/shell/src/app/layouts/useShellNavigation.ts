import { useCallback, useEffect, useState } from "react";
import { useLocation } from "react-router-dom";

const COMPACT_QUERY = "(max-width: 1056px)";

function readDesktopPreference(key: string) {
  try {
    return window.localStorage.getItem(key) !== "false";
  } catch {
    return true;
  }
}

export function useShellNavigation(storageKey: string, enabled = true) {
  const location = useLocation();
  const [state, setState] = useState(() => {
    const compact = window.matchMedia(COMPACT_QUERY).matches;
    return { compact, visible: !compact && readDesktopPreference(storageKey) };
  });

  useEffect(() => {
    const media = window.matchMedia(COMPACT_QUERY);
    const updateViewport = () =>
      setState({
        compact: media.matches,
        visible: !media.matches && readDesktopPreference(storageKey),
      });
    media.addEventListener("change", updateViewport);
    return () => media.removeEventListener("change", updateViewport);
  }, [storageKey]);

  useEffect(() => {
    if (state.compact) return;
    try {
      window.localStorage.setItem(storageKey, String(state.visible));
    } catch {
      // Navigation still works when browser storage is unavailable.
    }
  }, [state.compact, state.visible, storageKey]);

  useEffect(() => {
    if (state.compact) setState((previous) => ({ ...previous, visible: false }));
  }, [location.pathname, state.compact]);

  useEffect(() => {
    if (!enabled || !state.compact || !state.visible) return;
    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setState((previous) => ({ ...previous, visible: false }));
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = originalOverflow;
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [enabled, state.compact, state.visible]);

  const setVisible = useCallback((visible: boolean) => {
    setState((previous) => ({ ...previous, visible }));
  }, []);
  const toggle = useCallback(() => {
    setState((previous) => ({ ...previous, visible: !previous.visible }));
  }, []);

  return { visible: state.visible, compact: state.compact, setVisible, toggle };
}
