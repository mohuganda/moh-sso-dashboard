import { Tile, InlineLoading } from "@carbon/react";
import type React from "react";

import { useAuditOverviewQuery } from "@moh-sso/api";

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
    return <div style={{ opacity: 0.7 }}>Failed to load audit metrics</div>;
  }

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
        gap: 12,
      }}
    >
      <Tile>
        <strong>Total Events</strong>
        <div style={{ fontSize: 24 }}>{data?.total_events ?? "—"}</div>
      </Tile>

      <Tile>
        <strong>Total Failures</strong>
        <div style={{ fontSize: 24 }}>{data?.total_failures ?? "—"}</div>
      </Tile>

      <Tile>
        <strong>Failed Logins</strong>
        <div style={{ fontSize: 24 }}>{data?.failed_logins ?? "—"}</div>
      </Tile>

      <Tile>
        <strong>Successful Logins</strong>
        <div style={{ fontSize: 24 }}>{data?.successful_logins ?? "—"}</div>
      </Tile>
    </div>
  );
};
