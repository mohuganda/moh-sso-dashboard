import { InlineNotification } from "@carbon/react";
import { useEffect, useRef } from "react";

type RecoverableQuery = { error?: unknown; refetch: () => unknown };

function observatoryUnavailable(error: unknown): boolean {
  if (!error || typeof error !== "object" || !("status" in error) || error.status !== 503) {
    return false;
  }
  const data = "data" in error ? error.data : null;
  return Boolean(
    data &&
    typeof data === "object" &&
    "error" in data &&
    data.error &&
    typeof data.error === "object" &&
    "code" in data.error &&
    data.error.code === "OBSERVATORY_UNAVAILABLE",
  );
}

// Retry failed queries only, so an unavailable database can recover without a
// page reload or continuous polling during normal operation.
export function useSurveillanceRecovery(queries: RecoverableQuery[]): boolean {
  const current = useRef(queries);
  current.current = queries;
  const unavailable = queries.some((query) => observatoryUnavailable(query.error));
  useEffect(() => {
    if (!unavailable) return;
    const timer = window.setInterval(() => {
      for (const query of current.current) {
        if (observatoryUnavailable(query.error)) query.refetch();
      }
    }, 15_000);
    return () => window.clearInterval(timer);
  }, [unavailable]);
  return unavailable;
}

export function SurveillanceUnavailableNotice() {
  return (
    <InlineNotification
      kind="warning"
      title="Surveillance temporarily unavailable"
      subtitle="The database connection is being restored. Other portal services remain available. This page will retry automatically."
      hideCloseButton
      lowContrast
    />
  );
}
