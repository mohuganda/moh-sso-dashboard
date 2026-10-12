import { useEffect, useMemo, useRef, useState, Fragment } from "react";
import {
  Button,
  ComboBox,
  Dropdown,
  Popover,
  PopoverContent,
  Search,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableExpandHeader,
  TableExpandRow,
  TableExpandedRow,
  InlineLoading,
} from "@carbon/react";
import {
  ChevronDown,
  Download,
  Filter,
  ArrowRight,
  List,
  Renew,
} from "@carbon/react/icons";
import * as XLSX from "xlsx";
import { useNavigate } from "react-router-dom";
import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";

import { useGetIssuesQuery, useGetIssuesSummaryByProgramQuery } from "../api";
import type { Issue } from "../types";
import { getAvailablePeriods, periodType } from "../../../data-visualizer/src/pages/Constants.tsx";
import {
  useGetDataSetsQuery,
  type Dataset,
} from "../../../data-visualizer/src/pages/modals/data-model/data-model.ts";
import { useGetHierarchyQuery } from "../../../data-visualizer/src/pages/modals/orgunit/org-unit.ts";

import "./issue-dashboard.scss";

type DistrictSummary = {
  district: string;
  high: number;
  moderate: number;
  low: number;
  other: number;
  open: number;
  resolved: number;
  total: number;
};

type RegionSummaryRow = {
  id: string;
  region: string;
  high: number;
  moderate: number;
  low: number;
  critical?: number;
  info?: number;
  unspecified?: number;
  open: number;
  resolved: number;
  total: number;
  percentage: number;
  districts: DistrictSummary[];
  [key: string]: unknown;
};

type SelectEvent<T> = {
  selectedItem?: T | null;
};

type PeriodOption = {
  label: string;
  value?: string;
};

type SortHeader = "region" | "high" | "moderate" | "low" | "open" | "resolved" | "total" | "percentage" | string;

function safeString(val: unknown): string {
  if (val == null) return "";
  return String(val).trim();
}

function safeLower(val: unknown): string {
  return safeString(val).toLowerCase();
}

function normalizeSeverity(val?: string | null): string {
  if (!val) return "Unspecified";
  const trimmed = safeLower(val);
  if (trimmed === "high") return "High";
  if (trimmed === "moderate" || trimmed === "medium") return "Moderate";
  if (trimmed === "low") return "Low";
  if (trimmed === "critical") return "Critical";
  if (trimmed === "info") return "Info";
  const str = safeString(val);
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function isIssueResolved(status?: string | null): boolean {
  if (!status) return false;
  const s = safeLower(status);
  return s === "resolved" || s === "closed" || s === "done";
}

const IssueDashboard = () => {
  const navigate = useNavigate();
  const currentYear = new Date().getFullYear();

  /* -------------------------------------------------------------
   * API Queries
   * ------------------------------------------------------------- */
  const [selectedProgram, setSelectedProgram] = useState<string | undefined>();

  const {
    data: issuesData,
    isLoading: isLoadingIssues,
    error: issuesError,
    refetch: refetchIssues,
  } = useGetIssuesQuery({
    limit: 10000,
    offset: 0,
    program: selectedProgram,
  });

  const { data: summaryPrograms = [] } = useGetIssuesSummaryByProgramQuery();
  const { data: datasets = [] } = useGetDataSetsQuery();
  const { data: hierarchyData = [] } = useGetHierarchyQuery();

  /* -------------------------------------------------------------
   * Filters State
   * ------------------------------------------------------------- */
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);
  const [selectedPeriodType, setSelectedPeriodType] = useState<string>("Monthly");
  const [selectedPeriod, setSelectedPeriod] = useState<string>("");
  const [availablePeriods, setAvailablePeriods] = useState<PeriodOption[]>(
    getAvailablePeriods("Monthly", currentYear.toString()),
  );
  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState<boolean>(false);
  const periodPopoverRef = useRef<HTMLDivElement>(null);

  const [selectedDataset, setSelectedDataset] = useState<string>("");
  const [regionSearchTerm, setRegionSearchTerm] = useState<string>("");

  /* Sorting state */
  const [sortHeader, setSortHeader] = useState<SortHeader>("total");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  const years = useMemo(
    () => Array.from({ length: 10 }, (_, index) => currentYear - index),
    [currentYear],
  );

  const datasetNames = useMemo(
    () => (datasets as Dataset[]).map((d) => d.display_name).filter(Boolean),
    [datasets],
  );

  const programNames = useMemo(
    () => summaryPrograms.map((p) => p.program).filter(Boolean),
    [summaryPrograms],
  );

  const hasActiveFilters =
    selectedYear !== currentYear ||
    selectedPeriodType !== "Monthly" ||
    Boolean(selectedPeriod) ||
    Boolean(selectedDataset) ||
    Boolean(selectedProgram);

  /* Close popover on outside click */
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      if (periodPopoverRef.current && !periodPopoverRef.current.contains(target)) {
        setIsPeriodPopoverOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  /* -------------------------------------------------------------
   * Hierarchy Org-to-Region & Org-to-District mapping
   * ------------------------------------------------------------- */
  const { orgToRegionMap, orgToDistrictMap } = useMemo(() => {
    const regionMap = new Map<string, string>();
    const districtMap = new Map<string, string>();

    if (!hierarchyData) {
      return { orgToRegionMap: regionMap, orgToDistrictMap: districtMap };
    }

    const rootList: any[] = Array.isArray(hierarchyData)
      ? hierarchyData
      : typeof hierarchyData === "object"
      ? [hierarchyData]
      : [];

    function traverseTree(
      nodes: any[],
      currentRegion: string | null,
      currentDistrict: string | null,
    ) {
      if (!Array.isArray(nodes)) return;

      for (const node of nodes) {
        if (!node || typeof node !== "object") continue;

        const nodeName = safeString(node.name);
        const nodeId = safeString(node.id);

        const nodeRegion = currentRegion;
        const nodeDistrict = currentDistrict;

        // If top root (Level 1 e.g. "MOH - Uganda"), its children are Level 2 Regions
        if (!currentRegion) {
          if (node.children && Array.isArray(node.children) && node.children.length > 0) {
            for (const regionNode of node.children) {
              if (!regionNode || typeof regionNode !== "object") continue;

              const regName = safeString(regionNode.name);
              const regId = safeString(regionNode.id);

              if (regName) {
                regionMap.set(regName.toLowerCase(), regName);
                if (regId) regionMap.set(regId.toLowerCase(), regName);
              }

              // Children of region are Level 3 Districts
              if (regionNode.children && Array.isArray(regionNode.children) && regionNode.children.length > 0) {
                for (const distNode of regionNode.children) {
                  if (!distNode || typeof distNode !== "object") continue;

                  const distName = safeString(distNode.name);
                  const distId = safeString(distNode.id);

                  if (distName) {
                    districtMap.set(distName.toLowerCase(), distName);
                    if (regName) regionMap.set(distName.toLowerCase(), regName);
                    if (distId) {
                      districtMap.set(distId.toLowerCase(), distName);
                      if (regName) regionMap.set(distId.toLowerCase(), regName);
                    }
                  }
                  // Any facility or sub-county under district
                  if (distNode.children && Array.isArray(distNode.children) && distNode.children.length > 0) {
                    traverseTree(distNode.children, regName, distName);
                  }
                }
              }
            }
            continue;
          }
        }

        if (nodeName) {
          if (nodeRegion) regionMap.set(nodeName.toLowerCase(), nodeRegion);
          if (nodeDistrict) districtMap.set(nodeName.toLowerCase(), nodeDistrict);
        }
        if (nodeId) {
          if (nodeRegion) regionMap.set(nodeId.toLowerCase(), nodeRegion);
          if (nodeDistrict) districtMap.set(nodeId.toLowerCase(), nodeDistrict);
        }

        if (node.children && Array.isArray(node.children) && node.children.length > 0) {
          traverseTree(node.children, nodeRegion, nodeDistrict);
        }
      }
    }

    traverseTree(rootList, null, null);
    return { orgToRegionMap: regionMap, orgToDistrictMap: districtMap };
  }, [hierarchyData]);

  /* -------------------------------------------------------------
   * Filtered Issues list
   * ------------------------------------------------------------- */
  const rawIssues: Issue[] = useMemo(() => {
    return issuesData?.items ?? [];
  }, [issuesData]);

  const filteredIssues = useMemo(() => {
    return rawIssues.filter((issue) => {
      // Dataset filter
      if (selectedDataset) {
        if (safeLower(issue.dataset) !== safeLower(selectedDataset)) {
          return false;
        }
      }

      // Period filter
      if (selectedPeriod) {
        const targetPeriod = safeLower(selectedPeriod);
        const p1 = safeLower(issue.time_period);
        const p2 = safeLower(issue.time_Period);
        if (p1 !== targetPeriod && p2 !== targetPeriod) {
          return false;
        }
      } else if (selectedYear && selectedYear !== currentYear) {
        const yearStr = String(selectedYear);
        const p1 = safeString(issue.time_period);
        const p2 = safeString(issue.time_Period);
        if (!p1.includes(yearStr) && !p2.includes(yearStr)) {
          return false;
        }
      }

      return true;
    });
  }, [rawIssues, selectedDataset, selectedPeriod, selectedYear, currentYear]);

  /* -------------------------------------------------------------
   * Group and Aggregate by Region & Severity
   * ------------------------------------------------------------- */
  const { regionalSummary, grandTotals } = useMemo(() => {
    const regionGroups = new Map<
      string,
      {
        high: number;
        moderate: number;
        low: number;
        critical: number;
        info: number;
        unspecified: number;
        otherSeverities: Record<string, number>;
        open: number;
        resolved: number;
        total: number;
        districtGroups: Map<
          string,
          {
            high: number;
            moderate: number;
            low: number;
            other: number;
            open: number;
            resolved: number;
            total: number;
          }
        >;
      }
    >();

    const detectedSeverities = new Set<string>();

    let totalHigh = 0;
    let totalModerate = 0;
    let totalLow = 0;
    let totalCritical = 0;
    let totalInfo = 0;
    let totalUnspecified = 0;
    let totalOpen = 0;
    let totalResolved = 0;
    let totalAll = 0;

    for (const issue of filteredIssues) {
      // 1. Resolve Region
      let regionName = safeString(issue.region);
      if (!regionName && issue.org_unit) {
        regionName = orgToRegionMap.get(safeLower(issue.org_unit)) || "";
      }
      if (!regionName && issue.district) {
        regionName = orgToRegionMap.get(safeLower(issue.district)) || "";
      }
      if (!regionName) {
        regionName = "Unspecified";
      }

      // 2. Resolve District
      let districtName = safeString(issue.district);
      if (!districtName && issue.org_unit) {
        districtName = orgToDistrictMap.get(safeLower(issue.org_unit)) || safeString(issue.org_unit);
      }
      if (!districtName) {
        districtName = "Unspecified";
      }

      // 3. Resolve Severity
      const normalizedSev = normalizeSeverity(issue.severity);
      detectedSeverities.add(normalizedSev);

      // 4. Resolve Status
      const resolved = isIssueResolved(issue.status);

      // Init Region group if absent
      if (!regionGroups.has(regionName)) {
        regionGroups.set(regionName, {
          high: 0,
          moderate: 0,
          low: 0,
          critical: 0,
          info: 0,
          unspecified: 0,
          otherSeverities: {},
          open: 0,
          resolved: 0,
          total: 0,
          districtGroups: new Map(),
        });
      }

      const rGroup = regionGroups.get(regionName)!;
      rGroup.total += 1;
      totalAll += 1;

      if (resolved) {
        rGroup.resolved += 1;
        totalResolved += 1;
      } else {
        rGroup.open += 1;
        totalOpen += 1;
      }

      if (normalizedSev === "High") {
        rGroup.high += 1;
        totalHigh += 1;
      } else if (normalizedSev === "Moderate") {
        rGroup.moderate += 1;
        totalModerate += 1;
      } else if (normalizedSev === "Low") {
        rGroup.low += 1;
        totalLow += 1;
      } else if (normalizedSev === "Critical") {
        rGroup.critical += 1;
        totalCritical += 1;
      } else if (normalizedSev === "Info") {
        rGroup.info += 1;
        totalInfo += 1;
      } else if (normalizedSev === "Unspecified") {
        rGroup.unspecified += 1;
        totalUnspecified += 1;
      } else {
        rGroup.otherSeverities[normalizedSev] = (rGroup.otherSeverities[normalizedSev] || 0) + 1;
      }

      // District aggregation inside Region
      if (!rGroup.districtGroups.has(districtName)) {
        rGroup.districtGroups.set(districtName, {
          high: 0,
          moderate: 0,
          low: 0,
          other: 0,
          open: 0,
          resolved: 0,
          total: 0,
        });
      }

      const dGroup = rGroup.districtGroups.get(districtName)!;
      dGroup.total += 1;
      if (resolved) {
        dGroup.resolved += 1;
      } else {
        dGroup.open += 1;
      }

      if (normalizedSev === "High") {
        dGroup.high += 1;
      } else if (normalizedSev === "Moderate") {
        dGroup.moderate += 1;
      } else if (normalizedSev === "Low") {
        dGroup.low += 1;
      } else {
        dGroup.other += 1;
      }
    }

    // Build rows list
    const rows: RegionSummaryRow[] = [];
    regionGroups.forEach((val, region) => {
      const districtsList: DistrictSummary[] = [];
      val.districtGroups.forEach((dVal, dName) => {
        districtsList.push({
          district: dName,
          high: dVal.high,
          moderate: dVal.moderate,
          low: dVal.low,
          other: dVal.other,
          open: dVal.open,
          resolved: dVal.resolved,
          total: dVal.total,
        });
      });

      // Sort districts by total descending
      districtsList.sort((a, b) => b.total - a.total);

      const percentage = totalAll > 0 ? (val.total / totalAll) * 100 : 0;

      rows.push({
        id: `region-${region}`,
        region,
        high: val.high,
        moderate: val.moderate,
        low: val.low,
        critical: val.critical,
        info: val.info,
        unspecified: val.unspecified,
        open: val.open,
        resolved: val.resolved,
        total: val.total,
        percentage,
        districts: districtsList,
      });
    });

    return {
      regionalSummary: rows,
      grandTotals: {
        high: totalHigh,
        moderate: totalModerate,
        low: totalLow,
        critical: totalCritical,
        info: totalInfo,
        unspecified: totalUnspecified,
        open: totalOpen,
        resolved: totalResolved,
        total: totalAll,
        activeRegionsCount: regionGroups.size,
      },
    };
  }, [filteredIssues, orgToRegionMap, orgToDistrictMap]);

  /* -------------------------------------------------------------
   * Table Columns Definition
   * ------------------------------------------------------------- */
  const tableHeaders = useMemo(() => {
    const baseHeaders = [
      { key: "region", header: "Region" },
      { key: "high", header: "High" },
    ];

    if (grandTotals.critical > 0) {
      baseHeaders.push({ key: "critical", header: "Critical" });
    }
    if (grandTotals.info > 0) {
      baseHeaders.push({ key: "info", header: "Info" });
    }
    if (grandTotals.unspecified > 0) {
      baseHeaders.push({ key: "unspecified", header: "Unspecified" });
    }

    baseHeaders.push(
      { key: "open", header: "Open" },
      { key: "resolved", header: "Resolved" },
      { key: "total", header: "Total Issues" },
      { key: "percentage", header: "Share of Total" },
    );

    return baseHeaders;
  }, [grandTotals]);

  /* -------------------------------------------------------------
   * Sorted & Filtered Rows for Display
   * ------------------------------------------------------------- */
  const displayedRows = useMemo(() => {
    let result = [...regionalSummary];

    // Filter by Region search
    if (regionSearchTerm.trim()) {
      const term = safeLower(regionSearchTerm);
      result = result.filter(
        (r) =>
          safeLower(r.region).includes(term) ||
          r.districts.some((d) => safeLower(d.district).includes(term)),
      );
    }

    // Sort
    result.sort((a, b) => {
      let aVal = a[sortHeader];
      let bVal = b[sortHeader];

      if (typeof aVal === "string") {
        aVal = (aVal as string).toLowerCase();
        bVal = String(bVal ?? "").toLowerCase();
        if (sortDirection === "ASC") {
          return (aVal as string).localeCompare(bVal as string);
        }
        return (bVal as string).localeCompare(aVal as string);
      }

      const aNum = Number(aVal ?? 0);
      const bNum = Number(bVal ?? 0);
      if (sortDirection === "ASC") {
        return aNum - bNum;
      }
      return bNum - aNum;
    });

    return result;
  }, [regionalSummary, regionSearchTerm, sortHeader, sortDirection]);

  /* -------------------------------------------------------------
   * Handlers
   * ------------------------------------------------------------- */
  const handleSort = (headerKey: string) => {
    if (sortHeader === headerKey) {
      setSortDirection((prev) => (prev === "ASC" ? "DESC" : "ASC"));
    } else {
      setSortHeader(headerKey);
      setSortDirection("DESC");
    }
  };

  const handleYearChange = ({ selectedItem }: SelectEvent<number>) => {
    if (selectedItem == null) return;
    setSelectedYear(selectedItem);
    setAvailablePeriods(getAvailablePeriods(selectedPeriodType, selectedItem.toString()));
    setSelectedPeriod("");
  };

  const handlePeriodTypeChange = ({
    selectedItem,
  }: SelectEvent<{ label: string; value: string }>) => {
    if (!selectedItem) return;
    setSelectedPeriodType(selectedItem.value);
    setAvailablePeriods(getAvailablePeriods(selectedItem.value, selectedYear.toString()));
    setSelectedPeriod("");
  };

  const handlePeriodChange = ({ selectedItem }: SelectEvent<PeriodOption>) => {
    setSelectedPeriod(selectedItem?.label ?? "");
  };

  const handleResetFilters = () => {
    setSelectedYear(currentYear);
    setSelectedPeriodType("Monthly");
    setSelectedPeriod("");
    setAvailablePeriods(getAvailablePeriods("Monthly", currentYear.toString()));
    setSelectedDataset("");
    setSelectedProgram(undefined);
    setRegionSearchTerm("");
    setIsPeriodPopoverOpen(false);
  };

  /* -------------------------------------------------------------
   * Excel Export
   * ------------------------------------------------------------- */
  const handleExportExcel = () => {
    const exportHeaders = [
      "Region",
      "High Severity",
    ];

    if (grandTotals.critical > 0) exportHeaders.push("Critical Severity");
    if (grandTotals.info > 0) exportHeaders.push("Info Severity");
    if (grandTotals.unspecified > 0) exportHeaders.push("Unspecified Severity");

    exportHeaders.push("Open Issues", "Resolved Issues", "Total Issues", "Share (%)");

    const dataRows = regionalSummary.map((row) => {
      const rowData = [
        row.region,
        row.high,
      ];
      if (grandTotals.critical > 0) rowData.push(row.critical ?? 0);
      if (grandTotals.info > 0) rowData.push(row.info ?? 0);
      if (grandTotals.unspecified > 0) rowData.push(row.unspecified ?? 0);

      rowData.push(row.open, row.resolved, row.total, `${row.percentage.toFixed(1)}%`);
      return rowData;
    });

    // Grand total row
    const totalRow = [
      "TOTAL",
      grandTotals.high,
    ];
    if (grandTotals.critical > 0) totalRow.push(grandTotals.critical);
    if (grandTotals.info > 0) totalRow.push(grandTotals.info);
    if (grandTotals.unspecified > 0) totalRow.push(grandTotals.unspecified);

    totalRow.push(
      grandTotals.open,
      grandTotals.resolved,
      grandTotals.total,
      "100%",
    );

    const sheetData = [exportHeaders, ...dataRows, [], totalRow];
    const worksheet = XLSX.utils.aoa_to_sheet(sheetData);

    worksheet["!cols"] = exportHeaders.map(() => ({ wch: 18 }));

    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, "Severity Summary by Region");

    const dateStr = new Date().toISOString().slice(0, 10);
    XLSX.writeFile(workbook, `issue-severity-by-region-summary-${dateStr}.xlsx`);
  };

  const handleNavigateToIssues = (regionFilter?: string) => {
    if (regionFilter) {
      navigate(`/issues?region=${encodeURIComponent(regionFilter)}`);
    } else {
      navigate("/issues");
    }
  };

  /* Resolution rate */
  const resolutionRate =
    grandTotals.total > 0
      ? Math.round((grandTotals.resolved / grandTotals.total) * 100)
      : 0;

  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <div className="issue-dashboard-page">
        {/* ================= Header ================= */}
        <div className="dashboard-header">
          <div className="dashboard-header-text">
            <h2 className="dashboard-title">Issue Summary Dashboard</h2>
            <p className="dashboard-subtitle">
              Aggregated overview of reported data quality issues categorized by severity level and
              grouped by region.
            </p>
          </div>

          <div className="dashboard-actions">
            <Button
              size="md"
              kind="ghost"
              renderIcon={Renew}
              onClick={() => refetchIssues()}
              disabled={isLoadingIssues}
            >
              Refresh
            </Button>

            <Button
              size="md"
              kind="secondary"
              renderIcon={Download}
              onClick={handleExportExcel}
              disabled={regionalSummary.length === 0 || isLoadingIssues}
            >
              Export Summary (.xlsx)
            </Button>

            <Button
              size="md"
              kind="primary"
              renderIcon={List}
              onClick={() => handleNavigateToIssues()}
            >
              View Registered Issues
            </Button>
          </div>
        </div>

        {/* ================= KPI Metric Cards ================= */}
        <div className="kpi-grid">
          <div className="kpi-card kpi-card--total">
            <div className="kpi-header">
              <span className="kpi-label">Total Issues</span>
            </div>
            <div className="kpi-value">{isLoadingIssues ? "..." : grandTotals.total}</div>
            <p className="kpi-subtext">
              Across <strong>{grandTotals.activeRegionsCount}</strong> affected regions
            </p>
          </div>

          {grandTotals.critical > 0 && (
            <div className="kpi-card kpi-card--critical">
              <div className="kpi-header">
                <span className="kpi-label">Critical Severity</span>
                <span className="kpi-badge kpi-badge--critical">Critical</span>
              </div>
              <div className="kpi-value" style={{ color: "#a2191f" }}>
                {isLoadingIssues ? "..." : grandTotals.critical}
              </div>
              <p className="kpi-subtext">
                <strong>
                  {grandTotals.total > 0
                    ? `${((grandTotals.critical / grandTotals.total) * 100).toFixed(1)}%`
                    : "0%"}
                </strong>{" "}
                of total issues
              </p>
            </div>
          )}

          <div className="kpi-card kpi-card--high">
            <div className="kpi-header">
              <span className="kpi-label">High Severity</span>
              <span className="kpi-badge kpi-badge--high">High</span>
            </div>
            <div className="kpi-value" style={{ color: "#da1e28" }}>
              {isLoadingIssues ? "..." : grandTotals.high}
            </div>
            <p className="kpi-subtext">
              <strong>
                {grandTotals.total > 0
                  ? `${((grandTotals.high / grandTotals.total) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of total issues
            </p>
          </div>

          <div className="kpi-card kpi-card--moderate">
            <div className="kpi-header">
              <span className="kpi-label">Moderate Severity</span>
              <span className="kpi-badge kpi-badge--moderate">Moderate</span>
            </div>
            <div className="kpi-value" style={{ color: "#b26400" }}>
              {isLoadingIssues ? "..." : grandTotals.moderate}
            </div>
            <p className="kpi-subtext">
              <strong>
                {grandTotals.total > 0
                  ? `${((grandTotals.moderate / grandTotals.total) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of total issues
            </p>
          </div>

          <div className="kpi-card kpi-card--low">
            <div className="kpi-header">
              <span className="kpi-label">Low Severity</span>
              <span className="kpi-badge kpi-badge--low">Low</span>
            </div>
            <div className="kpi-value" style={{ color: "#0043ce" }}>
              {isLoadingIssues ? "..." : grandTotals.low}
            </div>
            <p className="kpi-subtext">
              <strong>
                {grandTotals.total > 0
                  ? `${((grandTotals.low / grandTotals.total) * 100).toFixed(1)}%`
                  : "0%"}
              </strong>{" "}
              of total issues
            </p>
          </div>

          <div className="kpi-card kpi-card--resolved">
            <div className="kpi-header">
              <span className="kpi-label">Resolution Rate</span>
              <span className="kpi-badge" style={{ backgroundColor: "#defbe6", color: "#198038" }}>
                {resolutionRate}%
              </span>
            </div>
            <div className="kpi-value" style={{ color: "#198038" }}>
              {isLoadingIssues ? "..." : grandTotals.resolved}
            </div>
            <p className="kpi-subtext">
              Resolved vs <strong>{grandTotals.open}</strong> open issues
            </p>
          </div>
        </div>

        {/* ================= Filters Bar ================= */}
        <div className="dashboard-filter-container">
          {/* Period Filter Popover */}
          <div ref={periodPopoverRef} className="issue-filter-wrapper">
            <Popover open={isPeriodPopoverOpen} align="bottom-left" dropShadow>
              <Button
                size="md"
                kind="tertiary"
                renderIcon={ChevronDown}
                onClick={() => setIsPeriodPopoverOpen((prev) => !prev)}
              >
                {selectedPeriod
                  ? `Period: ${selectedPeriod}`
                  : selectedYear !== currentYear
                  ? `Year: ${selectedYear}`
                  : "Reporting Period"}
              </Button>

              <PopoverContent className="filter-popover-content">
                <div className="popover-inner">
                  <Dropdown
                    id="dashboard-filter-year"
                    titleText="Year"
                    label="Select year"
                    items={years}
                    selectedItem={selectedYear}
                    itemToString={(item) => (item == null ? "" : String(item))}
                    onChange={handleYearChange}
                  />

                  <Dropdown
                    id="dashboard-filter-period-type"
                    titleText="Period Type"
                    label="Select period type"
                    items={periodType}
                    selectedItem={periodType.find((item) => item.value === selectedPeriodType)}
                    itemToString={(item) => item?.label ?? ""}
                    onChange={handlePeriodTypeChange}
                  />

                  <ComboBox
                    id="dashboard-filter-period"
                    titleText="Period"
                    placeholder="Select period"
                    items={availablePeriods}
                    selectedItem={
                      availablePeriods.find((item) => item.label === selectedPeriod) ?? null
                    }
                    itemToString={(item) => item?.label ?? ""}
                    onChange={handlePeriodChange}
                  />

                  <div className="popover-footer">
                    <Button
                      size="sm"
                      kind="ghost"
                      onClick={() => {
                        setSelectedPeriod("");
                        setIsPeriodPopoverOpen(false);
                      }}
                    >
                      Clear
                    </Button>
                    <Button size="sm" onClick={() => setIsPeriodPopoverOpen(false)}>
                      Apply
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </div>

          {/* Dataset Filter */}
          <div style={{ minWidth: "15rem" }}>
            <ComboBox
              id="dashboard-dataset-filter"
              size="md"
              titleText=""
              placeholder="Filter by Dataset"
              items={datasetNames}
              selectedItem={selectedDataset || null}
              onChange={({ selectedItem }: SelectEvent<string>) =>
                setSelectedDataset(selectedItem ?? "")
              }
            />
          </div>

          {/* Program Filter */}
          {programNames.length > 0 && (
            <div style={{ minWidth: "15rem" }}>
              <ComboBox
                id="dashboard-program-filter"
                size="md"
                titleText=""
                placeholder="Filter by Program"
                items={programNames}
                selectedItem={selectedProgram || null}
                onChange={({ selectedItem }: SelectEvent<string>) =>
                  setSelectedProgram(selectedItem ?? undefined)
                }
              />
            </div>
          )}

          {/* Reset Filters */}
          <Button
            size="md"
            kind="ghost"
            renderIcon={Filter}
            disabled={!hasActiveFilters}
            onClick={handleResetFilters}
          >
            Reset Filters
          </Button>

          {hasActiveFilters && (
            <span className="filter-active-indicator">
              Active Filters:{" "}
              {[
                selectedPeriod && `Period: ${selectedPeriod}`,
                selectedYear !== currentYear && !selectedPeriod && `Year: ${selectedYear}`,
                selectedDataset && `Dataset: ${selectedDataset}`,
                selectedProgram && `Program: ${selectedProgram}`,
              ]
                .filter(Boolean)
                .join(", ")}
            </span>
          )}
        </div>

        {/* ================= Search & Toolbar Row ================= */}
        <div className="dashboard-toolbar-row">
          <div className="dashboard-search">
            <Search
              labelText="Search regions or districts"
              placeholder="Search by region or district name..."
              value={regionSearchTerm}
              onChange={(e) => setRegionSearchTerm(e.target.value)}
              size="md"
              closeButtonLabelText="Clear search"
            />
          </div>
        </div>

        {/* ================= Data Table Section ================= */}
        {isLoadingIssues ? (
          <div className="dashboard-loading-state">
            <InlineLoading status="active" description="Loading regional issue summaries..." />
          </div>
        ) : issuesError ? (
          <div className="dashboard-empty-state">
            <h4>Failed to Load Issue Data</h4>
            <p>An error occurred while retrieving issue data from the server.</p>
            <Button size="sm" kind="tertiary" onClick={() => refetchIssues()}>
              Retry
            </Button>
          </div>
        ) : displayedRows.length === 0 ? (
          <div className="dashboard-empty-state">
            <h4>No Regional Issues Found</h4>
            <p>
              {hasActiveFilters || regionSearchTerm
                ? "No issues match the selected filters or search criteria. Try adjusting or resetting the filters."
                : "There are currently no recorded issues to summarize."}
            </p>
            {hasActiveFilters && (
              <Button size="sm" kind="tertiary" onClick={handleResetFilters}>
                Reset Filters
              </Button>
            )}
          </div>
        ) : (
          <div className="summary-table-container">
            <DataTable rows={displayedRows} headers={tableHeaders}>
              {({
                rows,
                headers,
                getTableProps,
                getHeaderProps,
                getRowProps,
                getExpandHeaderProps,
              }) => (
                <Table {...getTableProps()} aria-label="Issue severity summary by region table">
                  <TableHead>
                    <TableRow>
                      <TableExpandHeader
                        {...getExpandHeaderProps()}
                      />
                      {headers.map((header) => {
                        const isSorted = sortHeader === header.key;
                        const { key: headerKey, ...headerProps } = getHeaderProps({
                          header,
                          isSortable: true,
                          onClick: () => handleSort(header.key),
                        });
                        return (
                          <TableHeader
                            key={headerKey}
                            {...headerProps}
                            isSortable
                            isSortHeader={isSorted}
                            sortDirection={isSorted ? sortDirection : "NONE"}
                          >
                            {header.header}
                          </TableHeader>
                        );
                      })}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const summaryItem = displayedRows.find((item) => item.id === row.id);
                      const { key: rowKey, ...rowProps } = getRowProps({ row });

                      return (
                        <Fragment key={row.id}>
                          <TableExpandRow
                            key={rowKey}
                            {...rowProps}
                          >
                            {row.cells.map((cell) => {
                              const headerKey = cell.info.header;

                              if (headerKey === "region") {
                                return (
                                  <TableCell key={cell.id}>
                                    <div className="region-cell">
                                      <span>{cell.value}</span>
                                    </div>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "high") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span
                                      className={`severity-pill ${
                                        count > 0
                                          ? "severity-pill--high"
                                          : "severity-pill--zero"
                                      }`}
                                    >
                                      {count > 0 ? count : "-"}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "critical" || headerKey === "info" || headerKey === "unspecified") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span
                                      className={`severity-pill ${
                                        count > 0
                                          ? "severity-pill--neutral"
                                          : "severity-pill--zero"
                                      }`}
                                    >
                                      {count > 0 ? count : "-"}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "open") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="stat-cell">
                                      <span className="stat-dot stat-dot--open" />
                                      {count}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "resolved") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="stat-cell">
                                      <span className="stat-dot stat-dot--resolved" />
                                      {count}
                                    </span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "total") {
                                const count = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <span className="total-pill">{count}</span>
                                  </TableCell>
                                );
                              }

                              if (headerKey === "percentage") {
                                const pct = Number(cell.value || 0);
                                return (
                                  <TableCell key={cell.id}>
                                    <div className="share-bar-wrapper">
                                      <div className="share-bar-bg">
                                        <div
                                          className="share-bar-fill"
                                          style={{ width: `${Math.min(100, pct)}%` }}
                                        />
                                      </div>
                                      <span className="share-percent-text">
                                        {pct.toFixed(1)}%
                                      </span>
                                    </div>
                                  </TableCell>
                                );
                              }

                              return <TableCell key={cell.id}>{cell.value ?? "-"}</TableCell>;
                            })}
                          </TableExpandRow>

                          <TableExpandedRow
                            key={`expanded-${row.id}`}
                            colSpan={headers.length + 1}
                          >
                            {row.isExpanded && summaryItem && (
                              <div className="expanded-panel">
                                <div className="expanded-panel-header">
                                  <div>
                                    <h4 className="expanded-panel-title">
                                      District Breakdown for {summaryItem.region} Region
                                    </h4>
                                    <p style={{ fontSize: "0.8125rem", color: "#525252", margin: "0.25rem 0 0" }}>
                                      {summaryItem.districts.length} district(s) affected |{" "}
                                      <strong>{summaryItem.total}</strong> total issues (
                                      <span style={{ color: "#da1e28", fontWeight: 600 }}>
                                        {summaryItem.high} High
                                      </span>
                                      )
                                    </p>
                                  </div>

                                  <Button
                                    size="sm"
                                    kind="tertiary"
                                    renderIcon={ArrowRight}
                                    onClick={() => handleNavigateToIssues(summaryItem.region)}
                                  >
                                    View Issues in {summaryItem.region}
                                  </Button>
                                </div>

                                <div className="district-table-wrapper">
                                  <table>
                                    <thead>
                                      <tr>
                                        <th>District</th>
                                        <th>High Severity</th>
                                        <th>Open</th>
                                        <th>Resolved</th>
                                        <th>Total Issues</th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {summaryItem.districts.map((d) => (
                                        <tr key={d.district}>
                                          <td>
                                            <strong>{d.district}</strong>
                                          </td>
                                          <td>
                                            <span
                                              className={`severity-pill ${
                                                d.high > 0
                                                  ? "severity-pill--high"
                                                  : "severity-pill--zero"
                                              }`}
                                            >
                                              {d.high > 0 ? d.high : "-"}
                                            </span>
                                          </td>
                                          <td>
                                            <span className="stat-cell">
                                              <span className="stat-dot stat-dot--open" />
                                              {d.open}
                                            </span>
                                          </td>
                                          <td>
                                            <span className="stat-cell">
                                              <span className="stat-dot stat-dot--resolved" />
                                              {d.resolved}
                                            </span>
                                          </td>
                                          <td>
                                            <strong>{d.total}</strong>
                                          </td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                </div>
                              </div>
                            )}
                          </TableExpandedRow>
                        </Fragment>
                      );
                    })}

                    {/* ================= Grand Total Summary Row ================= */}
                    <TableRow className="grand-total-row">
                      <TableCell />
                      <TableCell>
                        <strong>TOTAL (All Regions)</strong>
                      </TableCell>
                      <TableCell>
                        <span className="severity-pill severity-pill--high">
                          {grandTotals.high}
                        </span>
                      </TableCell>
                      {grandTotals.critical > 0 && (
                        <TableCell>
                          <span className="severity-pill severity-pill--neutral">
                            {grandTotals.critical}
                          </span>
                        </TableCell>
                      )}
                      {grandTotals.info > 0 && (
                        <TableCell>
                          <span className="severity-pill severity-pill--neutral">
                            {grandTotals.info}
                          </span>
                        </TableCell>
                      )}
                      {grandTotals.unspecified > 0 && (
                        <TableCell>
                          <span className="severity-pill severity-pill--neutral">
                            {grandTotals.unspecified}
                          </span>
                        </TableCell>
                      )}
                      <TableCell>
                        <span className="stat-cell">
                          <span className="stat-dot stat-dot--open" />
                          <strong>{grandTotals.open}</strong>
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="stat-cell">
                          <span className="stat-dot stat-dot--resolved" />
                          <strong>{grandTotals.resolved}</strong>
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="total-pill">{grandTotals.total}</span>
                      </TableCell>
                      <TableCell>
                        <span style={{ fontSize: "0.8125rem", fontWeight: 700 }}>
                          100.0%
                        </span>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              )}
            </DataTable>
          </div>
        )}
      </div>
    </PermissionGuard>
  );
};

export default IssueDashboard;
