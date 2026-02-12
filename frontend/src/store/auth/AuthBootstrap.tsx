import { useEffect, useRef } from "react";

import { useRefreshMutation } from "../api/auth.api";

export default function AuthBootstrap() {
  const [refresh] = useRefreshMutation();
  const didRun = useRef(false);

  useEffect(() => {
    if (didRun.current) return;
    didRun.current = true;

    refresh()
      .unwrap()
      .catch(() => {
        // refresh failure is handled inside the mutation
        // authLoaded() will still be dispatched there
      });
  }, [refresh]);

  return null;
}
