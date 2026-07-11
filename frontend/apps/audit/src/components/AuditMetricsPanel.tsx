import { Tile, InlineLoading } from "@carbon/react";
import type React from "react";

import { useAuditOverviewQuery } from "../api";
import "./audit-components.scss";

type Props = {
  from: string;
  to: string;
};

export const AuditMetricsPanel: React.FC<Props> = ({ from, to }) => {
  const { data, isLoading, isError } = useAuditOverviewQuery(
    { from, to },
    {
      skip: !from || !to,
    },
  );

  if (isLoading && !data) {
    return <InlineLoading description="Loading metrics..." />;
  }

  if (isError && !data) {
    return <div className="audit-metrics__error">Failed to load audit metrics.</div>;
  }

  return (
    <div className="audit-metrics">
      <MetricCard label="Total Events" value={data?.total_events} />
      <MetricCard label="Total Failures" value={data?.total_failures} />
      <MetricCard label="Failed Logins" value={data?.failed_logins} />
      <MetricCard label="Successful Logins" value={data?.successful_logins} />
    </div>
  );
};

function MetricCard({ label, value }: { label: string; value?: number | null }) {
  return (
    <Tile>
      <div className="audit-metrics__label">{label}</div>

      <div className="audit-metrics__value">{value ?? "—"}</div>
    </Tile>
  );
}
