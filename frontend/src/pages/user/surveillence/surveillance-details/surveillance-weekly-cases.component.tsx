import { AreaChart } from "@carbon/charts-react";
import "@carbon/charts/styles.css";
import { useMemo } from "react";

type WeeklyCasesPoint = {
  week: string | number;
  value: number;
};

type WeeklyCasesChartProps = {
  diseaseName?: string;
  data?: WeeklyCasesPoint[];
  height?: string;
};

export default function WeeklyCasesChart({
  diseaseName = "Disease",
  data = [],
  height = "320px",
}: WeeklyCasesChartProps) {
  const chartData = useMemo(() => {
    return data.map((item) => ({
      group: "Cases",
      key: String(item.week),
      value: Number(item.value ?? 0),
    }));
  }, [data]);

  const options = useMemo(
    () => ({
      title: `${diseaseName} Weekly Cases`,
      axes: {
        left: {
          mapsTo: "value",
          title: "Total cases",
          scaleType: "linear" as const,
        },
        bottom: {
          mapsTo: "key",
          title: "EPI Weeks",
          scaleType: "labels" as const,
        },
      },
      height,
      curve: "curveMonotoneX",
      toolbar: {
        enabled: false,
      },
      points: {
        enabled: true,
      },
    }),
    [diseaseName, height],
  );

  if (!chartData.length) {
    return (
      <div
        style={{
          minHeight: height,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: "1rem",
          border: "1px solid #e0e0e0",
          borderRadius: "0.25rem",
        }}
      >
        No weekly case data available.
      </div>
    );
  }

  return <AreaChart data={chartData} options={options} />;
}
