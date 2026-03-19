import { AreaChart } from "@carbon/charts-react";
import "@carbon/charts/styles.css";

const chartData = [
  { group: "Cases", key: "1", value: 2000 },
  { group: "Cases", key: "2", value: 1500 },
  { group: "Cases", key: "3", value: 2800 },
  { group: "Cases", key: "4", value: 1800 },
  { group: "Cases", key: "5", value: 2600 },
  { group: "Cases", key: "6", value: 3000 },
];

const options = {
  title: "Malaria Weekly Cases",
  axes: {
    left: {
      mapsTo: "value",
      title: "Total cases",
      scaleType: "linear",
    },
    bottom: {
      mapsTo: "key",
      title: "EPI Weeks",
      scaleType: "labels",
    },
  },
  height: "320px",
};

export default function WeeklyCasesChart() {
  return <AreaChart data={chartData} options={options} />;
}
