import React, { useState } from "react";

import DataModal from "./modals/data-model/DataModal.tsx";
import PeriodModal from "./modals/PeriodModal";
import OrgUnitModal from "./modals/orgunit/OrgUnitModal.tsx";

import "./data-visualizer.css";
import { Button, OverflowMenu, OverflowMenuItem } from "@carbon/react";
import { UpdateNow } from "@carbon/react/icons";

import GeneralModal from "./modals/GeneralModal.tsx";
import ChartRenderer from "./chartrenderer/ChartRenderer.tsx";

type VisualizerQuery = {
  dx: string[];
  pe: string[];
  ou: string[];
};

const DataVisualizer = () => {
  const [selectedData, setSelectedData] = useState([]);
  const [selectedPeriods, setSelectedPeriods] = useState([]);
  const [selectedOrgUnits, setSelectedOrgUnits] = useState([]);
  const [loadedChartData, setLoadedChartData] = useState([]);
  const [pivotChartData, setPivotChartData] = useState([]);
  const [showModal, setShowModal] = useState(null);
  const [appliedQuery, setAppliedQuery] = useState<VisualizerQuery | null>(null);

  // Load saved state from localStorage on component mount
  React.useEffect(() => {
    const savedData = localStorage.getItem("selectedData");
    const savedPeriods = localStorage.getItem("selectedPeriods");
    const savedOrgUnits = localStorage.getItem("selectedOrgUnits");
    const savedLoadedData = localStorage.getItem("loadedChartData");
    const savedPivotData = localStorage.getItem("pivotChartData");

    if (savedData) setSelectedData(JSON.parse(savedData));
    if (savedPeriods) setSelectedPeriods(JSON.parse(savedPeriods));
    if (savedOrgUnits) setSelectedOrgUnits(JSON.parse(savedOrgUnits));
    if (savedLoadedData) setLoadedChartData(JSON.parse(savedLoadedData));
    if (savedPivotData) setPivotChartData(JSON.parse(savedPivotData));
  }, []);

  // Save state to localStorage whenever selections change
  React.useEffect(() => {
    localStorage.setItem("selectedData", JSON.stringify(selectedData));
  }, [selectedData]);

  React.useEffect(() => {
    localStorage.setItem("selectedPeriods", JSON.stringify(selectedPeriods));
  }, [selectedPeriods]);

  React.useEffect(() => {
    localStorage.setItem("selectedOrgUnits", JSON.stringify(selectedOrgUnits));
  }, [selectedOrgUnits]);

  React.useEffect(() => {
    localStorage.setItem("loadedChartData", JSON.stringify(loadedChartData));
  }, [loadedChartData]);

  React.useEffect(() => {
    localStorage.setItem("pivotChartData", JSON.stringify(pivotChartData));
  }, [pivotChartData]);

  const open = (which) => {
    setShowModal(which);
  };
  const close = () => {
    setShowModal(null);
  };

  // Function to clear all selections
  const clearAllSelections = () => {
    setSelectedData([]);
    setSelectedPeriods([]);
    setSelectedOrgUnits([]);
    setPivotChartData([]);
    setLoadedChartData([]);
  };

  // Function to check if all required dimensions are selected
  const isDataReady =
    selectedData?.length > 0 && selectedPeriods?.length > 0 && selectedOrgUnits?.length > 0;

  const query: VisualizerQuery | null = isDataReady
    ? {
        dx: selectedData.map((item: any) => item.data_element_id),
        pe: selectedPeriods.map((item: any) => item.id),
        ou: selectedOrgUnits,
      }
    : null;

  // const downloadExcel = async () => {
  //     if (!appliedQuery) return;
  //     try {
  //         // Pull pivot table by period to export a flat table (category + series columns)
  //         const res = await fetch(`/api/analytics?dx=${encodeURIComponent(appliedQuery.dx)}&pe=${encodeURIComponent(appliedQuery.pe)}&ou=${encodeURIComponent(appliedQuery.ou)}&groupBy=period`);
  //         const json = await res.json();
  //         const table = json.table || [];
  //         if (!table.length) return;
  //         const ws = XLSX.utils.json_to_sheet(table);
  //         const wb = XLSX.utils.book_new();
  //         XLSX.utils.book_append_sheet(wb, ws, "Data");
  //         XLSX.writeFile(wb, "visualization_data.xlsx");
  //     } catch (e) {
  //         console.error("Excel download failed", e);
  //     }
  // };

  return (
    <div className="dv-container">
      {/*/!* Left Sidebar *!/*/}
      {/*<div className="dv-sidebar demension-container">*/}
      {/*    <TreeView*/}
      {/*        label="MAIN DIMENSIONS"*/}
      {/*    >*/}
      {/*        {mainDimension.map((item) => (*/}
      {/*            <TreeNode*/}
      {/*                id={item?.id}*/}
      {/*                label={item?.label}*/}
      {/*                renderIcon={item.icon}*/}
      {/*                value={item?.value}*/}
      {/*                onSelect={() => open(item?.value)}*/}
      {/*            />*/}
      {/*        ))}*/}
      {/*    </TreeView>*/}
      {/*</div>*/}

      {/* Main Workspace */}
      <div className="dv-workspace">
        {/* Toolbar */}
        <div className="dv-toolbar">
          <Button
            size="sm"
            kind="tertiary"
            // disabled={!isDataReady}
            renderIcon={UpdateNow}
            className={`dwh-btn-width`}
            // onClick={() => open("general")}
            onClick={() => {
              if (query) setAppliedQuery(query);
            }}
          >
            Update
          </Button>

          {/*{isDataReady && (*/}
          {/*    <Button*/}
          {/*        size="sm"*/}
          {/*        kind="danger--tertiary"*/}
          {/*        renderIcon={FilterRemove}*/}
          {/*        className={`dwh-btn-width`}*/}
          {/*        onClick={clearAllSelections}*/}
          {/*    >*/}
          {/*        Clear All*/}
          {/*    </Button>*/}
          {/*)}*/}

          <OverflowMenu aria-label="overflow-menu" align="bottom" flipped>
            <OverflowMenuItem hasDivider itemText="More Options" />
            <OverflowMenuItem hasDivider itemText="Download" />
            {isDataReady && (
              <OverflowMenuItem
                hasDivider
                isDelete
                onClick={clearAllSelections}
                itemText="Clear All"
              />
            )}
          </OverflowMenu>
        </div>

        {/* Layout Area */}
        <div className="dv-layout">
          {/* Status Indicator */}
          {!isDataReady && (
            <div className="dv-alert dv-alert-warning">
              <i className="fas fa-exclamation-triangle"></i>
              Please select Data, Period, and Organisation Unit to view data
            </div>
          )}

          {/* Dimension Layout */}
          <div className="dv-dimension-layout">
            <div className="dv-dimension-grid">
              <div className="dv-dimension-section">
                <div className="dv-section-label">Data</div>
                <div className="dv-dropzone">
                  <div className="dv-dropzone-content">
                    {selectedData.length >= 0 ? (
                      <div
                        className="dv-badge"
                        onClick={() => {
                          open("data");
                        }}
                        style={{ cursor: "pointer" }}
                      >
                        <i className="fas fa-database dv-badge-icon"></i>
                        <span>Data</span>
                        <span className="dv-badge-count">{selectedData.length}</span>
                      </div>
                    ) : (
                      <span className="dv-placeholder">Select data dimension</span>
                    )}
                  </div>
                </div>
              </div>

              <div className="dv-dimension-section">
                <div className="dv-section-label">Period</div>
                <div className="dv-dropzone">
                  <div className="dv-dropzone-content">
                    {selectedPeriods.length >= 0 ? (
                      <div
                        className="dv-badge"
                        onClick={() => {
                          open("period");
                        }}
                        style={{ cursor: "pointer" }}
                      >
                        <i className="fas fa-clock dv-badge-icon"></i>
                        <span>Period</span>
                        <span className="dv-badge-count">{selectedPeriods.length}</span>
                      </div>
                    ) : (
                      <span className="dv-placeholder">Select period dimension</span>
                    )}
                  </div>
                </div>
              </div>

              <div className="dv-dimension-section">
                <div className="dv-section-label">Organisation Unit</div>
                <div className="dv-dropzone">
                  <div className="dv-dropzone-content">
                    {selectedOrgUnits.length >= 0 ? (
                      <div
                        className="dv-badge"
                        onClick={() => {
                          open("orgunit");
                        }}
                        style={{ cursor: "pointer" }}
                      >
                        <i className="fas fa-sitemap dv-badge-icon"></i>
                        <span>Organisation Unit</span>
                        <span className="dv-badge-count">{selectedOrgUnits.length}</span>
                      </div>
                    ) : (
                      <span className="dv-placeholder">Select organisation unit dimension</span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="dv-canvas">
            {appliedQuery || loadedChartData?.length > 0 ? (
              <ChartRenderer
                queryParams={appliedQuery}
                onSaveLoadedData={setLoadedChartData}
                onSavePivotData={setPivotChartData}
                loadedData={loadedChartData}
                pivotData={pivotChartData}
                periods={selectedPeriods}
              />
            ) : (
              <div className="dv-canvas-placeholder">
                <i className="fas fa-chart-bar"></i>
                <h6>No Data Available</h6>
                <p>
                  Select data elements, periods, and organization units, then click Update to see
                  your visualization
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {showModal === "data" && (
        <DataModal onClose={close} selected={selectedData} onSave={setSelectedData} />
      )}
      {showModal === "period" && (
        <PeriodModal onClose={close} selected={selectedPeriods} onSave={setSelectedPeriods} />
      )}
      {showModal === "orgunit" && (
        <OrgUnitModal onClose={close} selected={selectedOrgUnits} onSave={setSelectedOrgUnits} />
      )}
      {showModal === "general" && (
        <GeneralModal
          onClose={close}
          selectedData={selectedData}
          onSaveData={setSelectedData}
          selectedPeriods={selectedPeriods}
          onSavePeriods={setSelectedPeriods}
          selectedOrgUnits={selectedOrgUnits}
          onSaveOrgUnits={setSelectedOrgUnits}
        />
      )}
    </div>
  );
};

export default DataVisualizer;
