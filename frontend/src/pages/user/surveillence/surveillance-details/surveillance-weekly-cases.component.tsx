import { SimpleBarChart } from "@carbon/charts-react";
import { ScaleTypes } from "@carbon/charts";
import "@carbon/charts/styles.css";
import { useMemo } from "react";
import "./surveillance-details.css";

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
          scaleType: ScaleTypes.LINEAR,
        },
        bottom: {
          mapsTo: "key",
          title: "EPI Weeks",
          scaleType: ScaleTypes.LABELS,
        },
      },
      height,
      toolbar: {
        enabled: false,
      },
      legend: {
        enabled: false,
      },
    }),
    [diseaseName, height],
  );

  if (!chartData.length) {
    return <div className="disease-details-page__placeholder">No weekly case data available.</div>;
  }

  return (
    <div className="disease-details-page__chart">
      <SimpleBarChart data={chartData} options={options} />
    </div>
  );
}
