import { useEffect, useState } from "react";
import PivotTableUI from "react-pivottable/PivotTableUI";
import createPlotlyRenderers from "react-pivottable/PlotlyRenderers";
import TableRenderers from "react-pivottable/TableRenderers";
import Plot from "react-plotly.js";

import API from "../../helpers/api";

import "react-pivottable/pivottable.css";

const ChartRenderer = ({
  queryParams,
  onSaveLoadedData,
  onSavePivotData,
  loadedData,
  pivotData,
  periods,
}) => {
  const utils = createPlotlyRenderers(Plot);

  const [loading, setLoading] = useState(false);
  const [chartData, setChartData] = useState(loadedData);
  const [pivotTableData, setPivotTableData] = useState(pivotData);

  useEffect(() => {
    if (!queryParams) return;

    const fetchChartData = async () => {
      setLoading(true);

      try {
        const response = await API.post("/visualizer/datavalues", queryParams);
        const { status, data } = response;

        if (status === 200) {
          const rows = data.rows || [];

          const mappedPivotData =
            rows?.map((item) => ({
              "Age-Sex Disaggregation": item.co,
              District: item.district,
              "Data Element": item.dxName,
              "Facility Name": item.ouName,
              Period: periods.find((period) => period.id === item.pe)?.label ?? "",
              Region: item.region,
              SubCounty: item.subCounty,
            })) ?? [];

          setChartData(rows);
          onSaveLoadedData(rows);

          setPivotTableData(mappedPivotData);
          onSavePivotData(mappedPivotData);
        } else {
          setChartData([]);
        }
      } catch {
        setChartData([]);
      } finally {
        setLoading(false);
      }
    };

    fetchChartData();
  }, [queryParams, onSaveLoadedData, onSavePivotData, periods]);

  return (
    <>
      {loading ? (
        <div> Loading... </div>
      ) : (
        // <StackedBarChart data={[]} options={emptyChartOptions} />
        <>
          {chartData.length === 0 && !loading && (
            <div className="dv-canvas-placeholder">
              <i className="fas fa-chart-bar"></i>
              <h6>No Data Available for the selection</h6>
            </div>
          )}

          {chartData?.length > 0 && (
            <div className={`dwh-chart-container`}>
              <PivotTableUI
                data={pivotTableData}
                cols={["Data Element"]}
                onChange={(s) => {
                  setPivotTableData(s);
                }}
                renderers={{ ...TableRenderers, ...utils }}
                {...pivotTableData}
              />
            </div>
          )}
        </>
      )}
    </>
  );
};

export default ChartRenderer;
