import { useEffect, useState } from "react";
import PivotTableUI from "react-pivottable/PivotTableUI";
import createPlotlyRenderers from "react-pivottable/PlotlyRenderers";
import TableRenderers from "react-pivottable/TableRenderers";
import { aggregatorTemplates } from 'react-pivottable/Utilities';
import Plot from "react-plotly.js";

import "react-pivottable/pivottable.css";
import {useLazyGetDataValuesQuery} from "./ChartOptions.ts";

const ChartRenderer = ({
  queryParams,
  onSaveLoadedData,
  onSavePivotData,
  loadedData,
  pivotData,
  periods,
}) => {
  const utils = createPlotlyRenderers(Plot);
  const intFormat = (val) => Math.round(val).toLocaleString();
  const [loading, setLoading] = useState(false);
  const [chartData, setChartData] = useState(loadedData);
  const [pivotTableData, setPivotTableData] = useState(pivotData);
  const [ triggerGetDataValues ] = useLazyGetDataValuesQuery();

  const customAggregators = {
    "Sum": aggregatorTemplates.sum(intFormat),
    "Average": aggregatorTemplates.average(intFormat),
    "Count": aggregatorTemplates.count(intFormat)
  };
  useEffect(() => {
    if (!queryParams) return;

    const fetchChartData = async () => {
      setLoading(true);
      setChartData([]);
      try {
        const data = await triggerGetDataValues(queryParams).unwrap();
        const rows = data['rows'] || [];

        const mappedPivotData =
          rows?.map((item) => ({
            "Age-Sex Disaggregation": item.category_combo,
            District: item.district,
            "Data Element": item.dataelement,
            "Facility Name": item.facility,
            Period: periods.find((period) => period.id === item.period)?.label ?? "",
            Region: item.region,
            SubCounty: item.sub_county,
            Value: item?.value
          })) ?? [];

        setChartData(rows);
        onSaveLoadedData(rows);

        setPivotTableData(mappedPivotData);
        onSavePivotData(mappedPivotData);

      } catch (error) {
        setChartData([]);
        console.error("Error Encountered while fetching data elements:: " + error);
      }  finally {
        setLoading(false);
      }
    };

    fetchChartData();
  }, [queryParams, onSaveLoadedData, onSavePivotData, periods, triggerGetDataValues]);

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
                rows={["Facility Name"]}
                aggregators={customAggregators}
                aggregatorName="Sum"
                vals={["Value"]}
                hiddenFromDragDrop={["Value"]}
                hiddenFromAggregators={[
                    "Data Element",
                    "Facility Name",
                    "Age-Sex Disaggregation",
                    "District",
                    "Period",
                    "Region",
                    "SubCounty"
                ]}
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
