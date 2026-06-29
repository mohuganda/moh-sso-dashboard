import { Tile, InlineLoading } from "@carbon/react";
import type React from "react";

import { useAuditOverviewQuery } from "../api";

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
    return <div style={{ opacity: 0.7 }}>Failed to load audit metrics.</div>;
  }

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
        gap: 12,
      }}
    >
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
      <div
        style={{
          fontSize: "0.8125rem",
          color: "#6f6f6f",
          marginBottom: 8,
        }}
      >
        {label}
      </div>

      <div
        style={{
          fontSize: 24,
          fontWeight: 600,
        }}
      >
        {value ?? "—"}
      </div>
    </Tile>
  );
}
