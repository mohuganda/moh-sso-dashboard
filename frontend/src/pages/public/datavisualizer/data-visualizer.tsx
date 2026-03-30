import React, { useState } from "react";

import DataModal from "./modals/data-model/DataModal.tsx";
import PeriodModal from "./modals/period/PeriodModal.tsx";
import OrgUnitModal from "./modals/orgunit/OrgUnitModal.tsx";

import "./data-visualizer.css";
import { Button } from "@carbon/react";
import { Download, FilterRemove, UpdateNow } from "@carbon/react/icons";

import GeneralModal from "./modals/GeneralModal.tsx";
import ChartRenderer from "./chartrenderer/ChartRenderer.tsx";
import ClearDataConfirmationModal from "./modals/ClearDataConfirmationModal.tsx";

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
  const [isClearModalOpen, setIsClearModalOpen] = useState(false);

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
    setAppliedQuery(null);
    setIsClearModalOpen(false);
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

  const getDynamicFileName = (baseName = "Data_Report") => {
    const now = new Date();
    const datePart = now.toISOString().split('T')[0];
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');

    return `${baseName}_${datePart}_${hours}-${minutes}-${seconds}.csv`;
  };

  const exportDataToCSV = () => {
    const table = document.querySelector(".pvtTable");
    if (!table) return;

    const filename = getDynamicFileName();

    // @ts-ignore
    const rows = Array.from(table.rows);
    const grid: string[][] = [];

    rows.forEach((row, rowIndex) => {
      if (!grid[rowIndex]) { // @ts-ignore
        grid[rowIndex] = [];
      }

      let colIndex = 0;
      // @ts-ignore
      Array.from(row.cells).forEach((cell) => {
        while (grid[rowIndex][colIndex] !== undefined) {
          colIndex++;
        }

        // @ts-ignore
        const rowspan = cell.rowSpan || 1;
        // @ts-ignore
        const colspan = cell.colSpan || 1;
        // @ts-ignore
        const content = `"${cell.innerText.trim().replace(/"/g, '""')}"`;

        // Fill the grid for the current rowspan and colspan
        for (let r = 0; r < rowspan; r++) {
          for (let c = 0; c < colspan; c++) {
            const targetRow = rowIndex + r;
            const targetCol = colIndex + c;

            if (!grid[targetRow]) { // @ts-ignore
              grid[targetRow] = [];
            }

            // @ts-ignore
            grid[targetRow][targetCol] = (r === 0 && c === 0) ? content : '""';
          }
        }
        colIndex += colspan;
      });
    });

    // @ts-ignore
    const csvContent = grid
        .map(row => row.join(","))
        .join("\n");

    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("download", filename);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

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

          {loadedChartData?.length > 0 && (
              <>
                <Button
                    size="sm"
                    kind="primary"
                    renderIcon={Download}
                    className={`dwh-btn-width`}
                    onClick={exportDataToCSV}
                >
                  Download
                </Button>

                <Button
                    size="sm"
                    kind="danger--tertiary"
                    renderIcon={FilterRemove}
                    className={`dwh-btn-width`}
                    onClick={() => setIsClearModalOpen(true)}
                >
                  Clear All
                </Button>
              </>

          )}

          {/*<OverflowMenu aria-label="overflow-menu" align="bottom" flipped>*/}
          {/*  <OverflowMenuItem hasDivider itemText="More Options" />*/}
          {/*  <OverflowMenuItem hasDivider itemText="Download" />*/}
          {/*  {isDataReady && (*/}
          {/*    <OverflowMenuItem*/}
          {/*      hasDivider*/}
          {/*      isDelete*/}
          {/*      onClick={clearAllSelections}*/}
          {/*      itemText="Clear All"*/}
          {/*    />*/}
          {/*  )}*/}
          {/*</OverflowMenu>*/}
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
      {isClearModalOpen && (
          <ClearDataConfirmationModal
              setIsClearModalOpen={setIsClearModalOpen}
              isClearModalOpen={isClearModalOpen}
              handleConfirmClear={clearAllSelections}
          />
      )}
    </div>
  );
};

export default DataVisualizer;
