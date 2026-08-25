import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  Button,
  ComboBox,
  Dropdown,
  Popover,
  PopoverContent,
  Search,
  TreeView,
} from "@carbon/react";
import { Add, ChevronDown, Download, Filter, Upload } from "@carbon/react/icons";
import * as XLSX from "xlsx";

import {
  useGetIssuesQuery,
  useLazyGetIssuesQuery,
  useGetIssuesSummaryByProgramQuery,
  useGetIssueByCodeQuery,
} from "../api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";

function getIssueCodeFromUrl(): string {
  if (typeof window === "undefined") return "";
  const params = new URLSearchParams(window.location.search);
  const codeParam = params.get("issueCode") || params.get("issue") || params.get("code");
  if (codeParam && codeParam.trim()) {
    return codeParam.trim();
  }
  const parts = window.location.pathname.split("/").filter(Boolean);
  const lastPart = parts[parts.length - 1];
  if (lastPart && /^HMIS-\d+$/i.test(lastPart)) {
    return lastPart.trim();
  }
  return "";
}

import DataList from "../../../data-visualizer/src/pages/components/data-table/data-table.component.tsx";
import { getAvailablePeriods, periodType } from "../../../data-visualizer/src/pages/Constants.tsx";
import {
  useGetDataSetsQuery,
  useLazyGetDataSetElementsQuery,
  type Dataset,
  type ThemeElement,
} from "../../../data-visualizer/src/pages/modals/data-model/data-model.ts";
import { useGetHierarchyQuery } from "../../../data-visualizer/src/pages/modals/orgunit/org-unit.ts";

import { IssueModal } from "../component/issue-modal.component.tsx";
import { ImportIssuesModal } from "../component/import-issues-modal.component.tsx";
import { AssignModal } from "../component/assign-modal.component.tsx";
import { OrgUnitNode } from "../component/tree-node.component.tsx";
import { headers, IMPORT_TEMPLATE_HEADERS } from "../lib/constants.ts";
import IssueDetail from "./issue-detail/issue-detail.component.tsx";

import "./issue-tracker.scss";

export type Issue = {
  id?: string;
  issue_id: number;
  issue_code: string;
  dataset: string;
  data_element: string;
  org_unit: string;
  region?: string;
  district?: string;
  issue: string;
  date_reported: string;
  reported_by: string;
  status: string;
  issue_type: string;
  updated_by: string;
  updated_date: string;
  priority?: string;
  severity?: string;
  time_period?: string;
  time_Period?: string;
  assigned_to?: string;
};

type SelectEvent<T> = {
  selectedItem?: T | null;
};

type PeriodOption = {
  label: string;
  value?: string;
};

type OrgUnit = {
  id: string;
  name: string;
  children?: OrgUnit[];
};

function normalizeIssueRows(rows: Issue[] | undefined): Issue[] {
  return (
    rows?.map((item) => ({
      ...item,
      id: String(item.issue_id),
    })) ?? []
  );
}

function containsSearchTerm(value: unknown, searchTerm: string): boolean {
  return String(value ?? "")
    .toLowerCase()
    .includes(searchTerm);
}

function findNodeAndCollectSubtreeNames(
  nodes: OrgUnit[],
  targetName: string,
): Set<string> {
  const result = new Set<string>();
  const targetLower = targetName.trim().toLowerCase();

  const collectAll = (node: OrgUnit) => {
    if (node.name) {
      result.add(node.name.trim().toLowerCase());
    }
    if (node.children && Array.isArray(node.children)) {
      node.children.forEach(collectAll);
    }
  };

  const searchAndCollect = (nodeList: OrgUnit[]): boolean => {
    for (const node of nodeList) {
      if (node.name?.trim().toLowerCase() === targetLower) {
        collectAll(node);
        return true;
      }
      if (node.children && Array.isArray(node.children)) {
        if (searchAndCollect(node.children)) {
          return true;
        }
      }
    }
    return false;
  };

  if (Array.isArray(nodes)) {
    searchAndCollect(nodes);
  }

  if (result.size === 0) {
    result.add(targetLower);
  }

  return result;
}

const PROGRAM_COLORS = [
  "#0f62fe", // Blue
  "#008575", // Teal
  "#d07000", // Amber
  "#8a3ffc", // Purple
  "#da1e28", // Red/Ruby
  "#1192e8", // Cyan
  "#005d5d", // Dark Teal
  "#6f6f6f", // Slate/Grey
];

const IssueTracker = () => {
  const currentYear = new Date().getFullYear();

  const [showModal, setShowModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedRowIds, setSelectedRowIds] = useState<string[]>([]);

  const [issues, setIssues] = useState<Issue[]>([]);
  const [selectedIssue, setSelectedIssue] = useState<Issue>();
  const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);

  const [urlIssueCode, setUrlIssueCode] = useState<string>(() => getIssueCodeFromUrl());

  const { data: directIssueData, isLoading: isLoadingDirectIssue } = useGetIssueByCodeQuery(
    urlIssueCode,
    { skip: !urlIssueCode },
  );

  useEffect(() => {
    if (directIssueData && urlIssueCode) {
      setSelectedIssue(directIssueData);
      setIsViewIssueDetail(true);
    }
  }, [directIssueData, urlIssueCode]);

  const [tableSearchTerm, setTableSearchTerm] = useState("");

  const [selectedProgram, setSelectedProgram] = useState<string | undefined>();

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const { data, isLoading, error } = useGetIssuesQuery({
    limit: pageSize,
    offset: (page - 1) * pageSize,
    program: selectedProgram,
  });

  const [triggerGetIssues] = useLazyGetIssuesQuery();

  const { data: summaryData, isLoading: isLoadingSummary, error: summaryError } = useGetIssuesSummaryByProgramQuery();

  const summaryRows = useMemo(() => {
    return (summaryData ?? []).map((item, index) => ({
      id: item.program || `unspecified-${index}`,
      program: item.program || "Unspecified",
      issue_count: item.issue_count,
      open_count: item.open_count,
      resolved_count: item.resolved_count,
    }));
  }, [summaryData]);

  /*
   * Period filters
   */
  const [selectedYear, setSelectedYear] = useState(currentYear);

  const [selectedPeriodType, setSelectedPeriodType] = useState("Monthly");

  const [selectedPeriod, setSelectedPeriod] = useState("");

  const [availablePeriods, setAvailablePeriods] = useState<PeriodOption[]>(
    getAvailablePeriods("Monthly", currentYear.toString()),
  );

  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState(false);

  const periodPopoverRef = useRef<HTMLDivElement>(null);

  /*
   * Dataset filters
   */
  const [selectedDataset, setSelectedDataset] = useState("");

  const [selectedDataElement, setSelectedDataElement] = useState("");

  const [dataElements, setDataElements] = useState<ThemeElement[]>([]);

  const [isDataPopoverOpen, setIsDataPopoverOpen] = useState(false);

  const dataPopoverRef = useRef<HTMLDivElement>(null);

  const { data: datasets = [] } = useGetDataSetsQuery();

  const [triggerGetDataSetElements] = useLazyGetDataSetElementsQuery();

  /*
   * Organisation-unit filters
   */
  const [selectedOrgUnit, setSelectedOrgUnit] = useState("");

  const [orgSearchTerm, setOrgSearchTerm] = useState("");

  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());

  const [isOrgPopoverOpen, setIsOrgPopoverOpen] = useState(false);

  const orgPopoverRef = useRef<HTMLDivElement>(null);

  const { data: hierarchyData = [] } = useGetHierarchyQuery();

  const years = useMemo(
    () => Array.from({ length: 10 }, (_, index) => currentYear - index),
    [currentYear],
  );

  const datasetNames = useMemo(
    () => (datasets as Dataset[]).map((dataset) => dataset.display_name).filter(Boolean),
    [datasets],
  );

  const dataElementNames = useMemo(
    () =>
      dataElements
        .map((element) => element.data_element_short_name ?? element.data_element_short_name ?? "")
        .filter(Boolean),
    [dataElements],
  );

  const hasActiveFilters =
    selectedYear !== currentYear ||
    selectedPeriodType !== "Monthly" ||
    Boolean(selectedPeriod) ||
    Boolean(selectedDataset) ||
    Boolean(selectedDataElement) ||
    Boolean(selectedOrgUnit) ||
    Boolean(selectedProgram);

  useEffect(() => {
    setPage(1);
  }, [selectedProgram, selectedYear, selectedPeriod, selectedDataset, selectedDataElement, selectedOrgUnit]);

  const closeIssueModal = () => {
    setShowModal(false);
  };

  const closeImportModal = () => {
    setShowImportModal(false);
  };

  const closeAllPopovers = useCallback(() => {
    setIsPeriodPopoverOpen(false);
    setIsDataPopoverOpen(false);
    setIsOrgPopoverOpen(false);
  }, []);

  /* -----------------------------
   * Excel export columns
   * ----------------------------- */
  const EXPORT_COLUMNS: { key: string; header: string }[] = [
    ...headers,
    { key: "region", header: "Region" },
    { key: "district", header: "District" },
    { key: "time_period", header: "Reporting Period" },
    { key: "issue_type", header: "Issue Type" },
    { key: "priority", header: "Priority" },
    { key: "severity", header: "Severity" },
    { key: "reported_by", header: "Reported By" },
  ];

  const downloadTemplate = () => {
    const workbook = XLSX.utils.book_new();

    const worksheet = XLSX.utils.aoa_to_sheet([
      IMPORT_TEMPLATE_HEADERS,
      [
        "Example Dataset",
        "Example Data Element",
        "Example Org Unit",
        "Central Region",
        "Kampala District",
        "Example issue description",
        "Outliers",
        "High",
        "Moderate",
        "2025Q1",
      ],
    ]);

    worksheet["!cols"] = IMPORT_TEMPLATE_HEADERS.map(() => ({
      wch: 22,
    }));

    XLSX.utils.book_append_sheet(workbook, worksheet, "Issues Template");

    XLSX.writeFile(workbook, "issue-import-template.xlsx");
  };

  const downloadIssues = async () => {
    try {
      const allIssuesResponse = await triggerGetIssues({
        limit: 10000,
        offset: 0,
        program: selectedProgram,
      }).unwrap();

      const normalizedAll = normalizeIssueRows(allIssuesResponse.items);

      const searchTerm = tableSearchTerm.trim().toLowerCase();
      const issuesToDownload = searchTerm
        ? normalizedAll.filter((issue) =>
            [
              issue.issue_code,
              issue.dataset,
              issue.data_element,
              issue.issue,
              issue.status,
              issue.org_unit,
              issue.date_reported,
              issue.issue_type,
              issue.priority,
              issue.severity,
            ].some((field) => containsSearchTerm(field, searchTerm)),
          )
        : normalizedAll;

      const headerRow = EXPORT_COLUMNS.map((col) => col.header);

      const dataRows = issuesToDownload.map((issue) =>
        EXPORT_COLUMNS.map((col) => {
          if (col.key === "time_period") {
            return (issue as Record<string, unknown>)["time_period"] ??
                   (issue as Record<string, unknown>)["time_Period"] ??
                   "";
          }
          return (issue as Record<string, unknown>)[col.key] ?? "";
        }),
      );

      const workbook = XLSX.utils.book_new();

      const worksheet = XLSX.utils.aoa_to_sheet([headerRow, ...dataRows]);

      worksheet["!cols"] = EXPORT_COLUMNS.map(() => ({ wch: 22 }));

      XLSX.utils.book_append_sheet(workbook, worksheet, "Issues");

      XLSX.writeFile(workbook, `issues-${new Date().toISOString().slice(0, 10)}.xlsx`);
    } catch (err) {
      console.error("Failed to download issues:", err);
    }
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;

      if (periodPopoverRef.current && !periodPopoverRef.current.contains(target)) {
        setIsPeriodPopoverOpen(false);
      }

      if (dataPopoverRef.current && !dataPopoverRef.current.contains(target)) {
        setIsDataPopoverOpen(false);
      }

      if (orgPopoverRef.current && !orgPopoverRef.current.contains(target)) {
        setIsOrgPopoverOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  useEffect(() => {
    if (!isLoading && data) {
      setIssues(normalizeIssueRows(data.items));
    }

    if (error) {
      console.error("Error encountered while fetching issues:", error);
    }
  }, [data, error, isLoading]);

  const filteredIssues = useMemo(() => {
    let result = issues;

    if (selectedOrgUnit) {
      const targetNamesSet = findNodeAndCollectSubtreeNames(
        hierarchyData as OrgUnit[],
        selectedOrgUnit,
      );

      result = result.filter((issue) => {
        const issueOrg = issue.org_unit?.trim().toLowerCase();
        const issueRegion = issue.region?.trim().toLowerCase();
        const issueDistrict = issue.district?.trim().toLowerCase();

        return (
          (issueOrg && targetNamesSet.has(issueOrg)) ||
          (issueRegion && targetNamesSet.has(issueRegion)) ||
          (issueDistrict && targetNamesSet.has(issueDistrict))
        );
      });
    }

    if (selectedDataset) {
      const targetDataset = selectedDataset.trim().toLowerCase();
      result = result.filter(
        (issue) => issue.dataset?.trim().toLowerCase() === targetDataset,
      );
    }

    if (selectedDataElement) {
      const targetElement = selectedDataElement.trim().toLowerCase();
      result = result.filter(
        (issue) => issue.data_element?.trim().toLowerCase() === targetElement,
      );
    }

    if (selectedPeriod) {
      const targetPeriod = selectedPeriod.trim().toLowerCase();
      result = result.filter(
        (issue) =>
          issue.time_period?.trim().toLowerCase() === targetPeriod ||
          issue.time_Period?.trim().toLowerCase() === targetPeriod,
      );
    } else if (selectedYear && selectedYear !== currentYear) {
      const yearStr = String(selectedYear);
      result = result.filter(
        (issue) =>
          issue.time_period?.includes(yearStr) || issue.time_Period?.includes(yearStr),
      );
    }

    const searchTerm = tableSearchTerm.trim().toLowerCase();

    if (!searchTerm) {
      return result;
    }

    return result.filter((issue) =>
      [
        issue.issue_code,
        issue.dataset,
        issue.data_element,
        issue.issue,
        issue.status,
        issue.org_unit,
        issue.date_reported,
        issue.issue_type,
        issue.priority,
        issue.severity,
        issue.assigned_to,
      ].some((field) => containsSearchTerm(field, searchTerm)),
    );
  }, [
    issues,
    hierarchyData,
    selectedOrgUnit,
    selectedDataset,
    selectedDataElement,
    selectedPeriod,
    selectedYear,
    currentYear,
    tableSearchTerm,
  ]);

  const handleSelectRow = useCallback((rowId: string, checked: boolean) => {
    setSelectedRowIds((prev) => {
      if (checked) {
        return prev.includes(rowId) ? prev : [...prev, rowId];
      }
      return prev.filter((id) => id !== rowId);
    });
  }, []);

  const handleSelectAll = useCallback(
    (checked: boolean) => {
      if (checked) {
        const allIds = filteredIssues.map((item) => String(item.id ?? item.issue_id));
        setSelectedRowIds(allIds);
      } else {
        setSelectedRowIds([]);
      }
    },
    [filteredIssues],
  );

  const selectedIssueObjects = useMemo(() => {
    return issues.filter((item) => selectedRowIds.includes(String(item.id ?? item.issue_id)));
  }, [issues, selectedRowIds]);

  const selectedIssueCodes = useMemo(() => {
    return selectedIssueObjects
      .map((item) => item.issue_code)
      .filter(Boolean);
  }, [selectedIssueObjects]);

  const areAllSelectedAssigned = useMemo(() => {
    return (
      selectedIssueObjects.length > 0 &&
      selectedIssueObjects.every((item) => Boolean(item.assigned_to))
    );
  }, [selectedIssueObjects]);

  const handleIssueClick = (row: { id?: string }) => {
    const selectedItem = issues.find((item) => String(item.issue_id) === String(row.id ?? ""));

    if (!selectedItem) {
      return;
    }

    setSelectedIssue(selectedItem);
    setIsViewIssueDetail(true);

    if (selectedItem.issue_code && typeof window !== "undefined") {
      const url = new URL(window.location.href);
      url.searchParams.set("issueCode", selectedItem.issue_code);
      window.history.replaceState({}, "", url.pathname + url.search);
    }
  };

  const handleYearChange = ({ selectedItem }: SelectEvent<number>) => {
    if (selectedItem == null) {
      return;
    }

    setSelectedYear(selectedItem);

    setAvailablePeriods(getAvailablePeriods(selectedPeriodType, selectedItem.toString()));

    setSelectedPeriod("");
  };

  const handlePeriodTypeChange = ({
    selectedItem,
  }: SelectEvent<{
    label: string;
    value: string;
  }>) => {
    if (!selectedItem) {
      return;
    }

    setSelectedPeriodType(selectedItem.value);

    setAvailablePeriods(getAvailablePeriods(selectedItem.value, selectedYear.toString()));

    setSelectedPeriod("");
  };

  const handlePeriodChange = ({ selectedItem }: SelectEvent<PeriodOption>) => {
    setSelectedPeriod(selectedItem?.label ?? "");
  };

  const handleDatasetChange = async ({ selectedItem }: SelectEvent<string>) => {
    const datasetName = selectedItem ?? "";

    setSelectedDataset(datasetName);
    setSelectedDataElement("");

    if (!datasetName) {
      setDataElements([]);
      return;
    }

    const dataset = (datasets as Dataset[]).find((item) => item.display_name === datasetName);

    if (!dataset) {
      setDataElements([]);
      return;
    }

    try {
      const elements = await triggerGetDataSetElements(dataset.dataset_id).unwrap();

      setDataElements(elements);
    } catch (requestError) {
      console.error("Failed to fetch data elements:", requestError);

      setDataElements([]);
    }
  };

  const nodeMatchesSearch = useCallback((node: OrgUnit, term: string): boolean => {
    if (node.name.toLowerCase().includes(term.toLowerCase())) {
      return true;
    }

    return Boolean(node.children?.some((child) => nodeMatchesSearch(child, term)));
  }, []);

  const handleOrgSelect = (name: string) => {
    setSelectedOrgUnit((prev) => (prev === name ? "" : name));
  };

  const handleToggleNode = useCallback((nodeId: string, isExpanded: boolean) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (isExpanded) {
        next.add(nodeId);
      } else {
        next.delete(nodeId);
      }
      return next;
    });
  }, []);

  const renderRecursive = (nodes: OrgUnit[], idPrefix = "filter-org"): ReactNode[] => {
    if (!Array.isArray(nodes)) {
      return [];
    }

    return nodes
      .filter((node) => !orgSearchTerm || nodeMatchesSearch(node, orgSearchTerm))
      .map((node) => (
        <OrgUnitNode
          key={node.id}
          node={node}
          searchTerm={orgSearchTerm}
          selectedOrgUnit={selectedOrgUnit}
          expandedNodes={expandedNodes}
          onSelect={handleOrgSelect}
          onToggleNode={handleToggleNode}
          renderRecursive={renderRecursive}
          idPrefix={idPrefix}
        />
      ));
  };

  const handleFilter = (
    overrides: {
      period?: string;
      year?: number;
      dataset?: string;
      dataElement?: string;
      orgUnit?: string;
    } = {},
  ) => {
    closeAllPopovers();

    const sourceRows = data as Issue[] | undefined;

    if (!sourceRows) {
      return;
    }

    const filters = {
      period: selectedPeriod,
      year: selectedYear,
      dataset: selectedDataset,
      dataElement: selectedDataElement,
      orgUnit: selectedOrgUnit,
      ...overrides,
    };

    let nextIssues = normalizeIssueRows(sourceRows);

    if (filters.period) {
      nextIssues = nextIssues.filter(
        (issue) => issue.time_period === filters.period || issue.time_Period === filters.period,
      );
    } else if (filters.year) {
      const year = String(filters.year);

      nextIssues = nextIssues.filter(
        (issue) => issue.time_period?.includes(year) || issue.time_Period?.includes(year),
      );
    }

    if (filters.dataset) {
      nextIssues = nextIssues.filter((issue) => issue.dataset === filters.dataset);
    }

    if (filters.dataElement) {
      nextIssues = nextIssues.filter((issue) => issue.data_element === filters.dataElement);
    }

    if (filters.orgUnit) {
      nextIssues = nextIssues.filter((issue) => issue.org_unit === filters.orgUnit);
    }

    setIssues(nextIssues);
    closeAllPopovers();
  };

  const handleReset = () => {
    setSelectedYear(currentYear);
    setSelectedPeriodType("Monthly");
    setSelectedPeriod("");

    setAvailablePeriods(getAvailablePeriods("Monthly", currentYear.toString()));

    setSelectedDataset("");
    setSelectedDataElement("");
    setDataElements([]);

    setSelectedOrgUnit("");
    setOrgSearchTerm("");
    setTableSearchTerm("");
    setSelectedProgram(undefined);

    setIssues(normalizeIssueRows(data as Issue[] | undefined));

    closeAllPopovers();
  };

  const handleBackFromDetail = () => {
    setSelectedIssue(undefined);
    setIsViewIssueDetail(false);
    setUrlIssueCode("");
    if (typeof window !== "undefined") {
      const url = new URL(window.location.href);
      url.searchParams.delete("issueCode");
      url.searchParams.delete("issue");
      url.searchParams.delete("code");
      window.history.replaceState({}, "", url.pathname + url.search);
    }
  };

  if (isLoadingDirectIssue && urlIssueCode && !selectedIssue) {
    return (
      <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
        <div style={{ padding: "3rem 2rem", textAlign: "center", color: "#525252" }}>
          <h4>Loading Issue Details...</h4>
          <p style={{ marginTop: "0.5rem" }}>Fetching issue [{urlIssueCode}] from server.</p>
        </div>
      </PermissionGuard>
    );
  }

  if (isViewIssueDetail && selectedIssue) {
    return (
      <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
        <IssueDetail
          selectedIssue={selectedIssue}
          goToBack={handleBackFromDetail}
        />
      </PermissionGuard>
    );
  }

  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <>
        <div className="program-summary-tiles-container">
          {isLoadingSummary ? (
            <div className="issue-loading-state">
              <p>Loading program summary...</p>
            </div>
          ) : summaryError ? (
            <div className="issue-empty-state">
              <h4>Failed to load summary</h4>
              <p>An error occurred while fetching the issues summary by program.</p>
            </div>
          ) : (
            <div className="summary-tiles-grid">
              {summaryRows.map((tile, index) => {
                const isActive = selectedProgram === (tile.program === "Unspecified" ? "" : tile.program);
                const tileColor = PROGRAM_COLORS[index % PROGRAM_COLORS.length];
                return (
                  <button
                    key={tile.id}
                    type="button"
                    className={`summary-tile ${isActive ? "active" : ""}`}
                    style={{ borderLeftColor: tileColor }}
                    onClick={() => {
                      if (isActive) {
                        setSelectedProgram(undefined);
                      } else {
                        setSelectedProgram(tile.program === "Unspecified" ? "" : tile.program);
                      }
                    }}
                  >
                    <div className="summary-tile-header">
                      <span className="summary-tile-program">{tile.program}</span>
                      <span className="summary-tile-count">{tile.issue_count}</span>
                    </div>
                    <div className="summary-tile-substats">
                      <span className="summary-substat open">
                        <span className="dot open-dot" />
                        <span className="label">Open:</span> <strong>{tile.open_count}</strong>
                      </span>
                      <span className="summary-substat resolved">
                        <span className="dot resolved-dot" />
                        <span className="label">Resolved:</span> <strong>{tile.resolved_count}</strong>
                      </span>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <div className="dv-toolbar issue-label-container">
          <div>
            <span className="issue-label">Registered Issues</span>
          </div>

          <div className="issue-toolbar-actions">
            <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
              <Button
                size="md"
                kind="tertiary"
                disabled={selectedIssueCodes.length === 0}
                onClick={() => setShowAssignModal(true)}
              >
                {areAllSelectedAssigned ? "Re-assign Selected" : "Assign Selected"}{" "}
                {selectedIssueCodes.length > 0 ? `(${selectedIssueCodes.length})` : ""}
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button size="md" kind="ghost" renderIcon={Download} onClick={downloadTemplate}>
                Template
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button
                size="md"
                kind="secondary"
                renderIcon={Upload}
                onClick={() => setShowImportModal(true)}
              >
                Import Issues
              </Button>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
              <Button
                size="md"
                kind="primary"
                renderIcon={Add}
                className="dwh-btn-width"
                onClick={() => setShowModal(true)}
              >
                New Issue
              </Button>
            </PermissionGuard>
          </div>
        </div>

        {selectedProgram !== undefined && (
          <div
            className="program-filter-banner"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              gap: "1rem",
              marginBottom: "1rem",
              padding: "0.75rem 1rem",
              backgroundColor: "#edf5ff",
              borderRadius: "4px",
              border: "1px solid #d0e2ff",
              color: "#0f62fe",
            }}
          >
            <div>
              Filtered by Program: <strong>{selectedProgram || "Unspecified"}</strong>
            </div>
            <Button
              size="sm"
              kind="ghost"
              onClick={() => setSelectedProgram(undefined)}
              style={{
                minHeight: "unset",
                padding: "4px 8px",
                color: "#0f62fe",
              }}
            >
              Clear Program Filter
            </Button>
          </div>
        )}
            <div className="issue-filter-container">
              <div ref={periodPopoverRef} className="issue-filter-wrapper">
                <Popover open={isPeriodPopoverOpen} align="bottom-left" dropShadow>
                  <Button
                    size="md"
                    kind="tertiary"
                    renderIcon={ChevronDown}
                    onClick={() => {
                      setIsPeriodPopoverOpen((current) => !current);
                      setIsDataPopoverOpen(false);
                      setIsOrgPopoverOpen(false);
                    }}
                  >
                    Period
                  </Button>

                  <PopoverContent className="filter-popover-content">
                    <div className="popover-inner">
                      <Dropdown
                        id="issue-filter-year"
                        titleText="Year"
                        label="Select year"
                        items={years}
                        selectedItem={selectedYear}
                        itemToString={(item) => (item == null ? "" : String(item))}
                        onChange={handleYearChange}
                      />

                      <Dropdown
                        id="issue-filter-period-type"
                        titleText="Period type"
                        label="Select period type"
                        items={periodType}
                        selectedItem={periodType.find((item) => item.value === selectedPeriodType)}
                        itemToString={(item) => item?.label ?? ""}
                        onChange={handlePeriodTypeChange}
                      />

                      <ComboBox
                        id="issue-filter-period"
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
                            handleFilter({
                              period: "",
                            });
                          }}
                        >
                          Clear
                        </Button>

                        <Button size="sm" onClick={() => handleFilter()}>
                          Update
                        </Button>
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </div>

              <div ref={dataPopoverRef} className="issue-filter-wrapper">
                <Popover open={isDataPopoverOpen} align="bottom-left" dropShadow>
                  <Button
                    size="md"
                    kind="tertiary"
                    renderIcon={ChevronDown}
                    onClick={() => {
                      setIsDataPopoverOpen((current) => !current);
                      setIsPeriodPopoverOpen(false);
                      setIsOrgPopoverOpen(false);
                    }}
                  >
                    Data
                  </Button>

                  <PopoverContent className="filter-popover-content">
                    <div className="popover-inner">
                      <ComboBox
                        id="issue-filter-dataset"
                        titleText="Dataset"
                        placeholder="Select dataset"
                        items={datasetNames}
                        selectedItem={selectedDataset || null}
                        onChange={handleDatasetChange}
                      />

                      <ComboBox
                        id="issue-filter-data-element"
                        titleText="Data element"
                        placeholder="Select data element"
                        items={dataElementNames}
                        selectedItem={selectedDataElement || null}
                        disabled={!selectedDataset || dataElementNames.length === 0}
                        onChange={({ selectedItem }: SelectEvent<string>) =>
                          setSelectedDataElement(selectedItem ?? "")
                        }
                      />

                      <div className="popover-footer">
                        <Button
                          size="sm"
                          kind="ghost"
                          onClick={() => {
                            setSelectedDataset("");
                            setSelectedDataElement("");
                            setDataElements([]);

                            handleFilter({
                              dataset: "",
                              dataElement: "",
                            });
                          }}
                        >
                          Clear
                        </Button>

                        <Button size="sm" onClick={() => handleFilter()}>
                          Update
                        </Button>
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </div>

              <div ref={orgPopoverRef} className="issue-filter-wrapper">
                <Popover open={isOrgPopoverOpen} align="bottom-left" dropShadow>
                  <Button
                    size="md"
                    kind="tertiary"
                    renderIcon={ChevronDown}
                    onClick={() => {
                      setIsOrgPopoverOpen((current) => !current);
                      setIsPeriodPopoverOpen(false);
                      setIsDataPopoverOpen(false);
                    }}
                  >
                    Organisation Unit
                  </Button>

                  <PopoverContent className="filter-popover-content">
                    <div className="popover-inner org-filter-inner">
                      <Search
                        labelText="Search organisation unit"
                        placeholder="Search..."
                        value={orgSearchTerm}
                        onChange={(event) => setOrgSearchTerm(event.target.value)}
                        size="sm"
                      />

                      <div className="org-tree-container">
                        <TreeView label="Organisation Units" hideLabel>
                          {renderRecursive(hierarchyData as OrgUnit[], "filter-org")}
                        </TreeView>
                      </div>

                      <div className="popover-footer">
                        <Button
                          size="sm"
                          kind="ghost"
                          onClick={() => {
                            setSelectedOrgUnit("");
                            setOrgSearchTerm("");

                            handleFilter({
                              orgUnit: "",
                            });
                          }}
                        >
                          Clear
                        </Button>

                        <Button size="sm" onClick={() => handleFilter()}>
                          Update
                        </Button>
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </div>

              <Button
                size="md"
                kind="ghost"
                renderIcon={Filter}
                disabled={!hasActiveFilters}
                onClick={handleReset}
              >
                Reset Filters
              </Button>
            </div>

            <div className="issue-table-search">
              <div className="issue-table-search-bar">
                <Search
                  labelText="Search issues"
                  placeholder="Search by issue code, dataset, data element, issue, status, or organisation unit"
                  value={tableSearchTerm}
                  onChange={(event) => setTableSearchTerm(event.target.value)}
                  size="lg"
                  closeButtonLabelText="Clear search"
                />
                <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
                  <Button size="md" kind="ghost" renderIcon={Download} onClick={downloadIssues}>
                    Download
                  </Button>
                </PermissionGuard>
              </div>
            </div>

            <DataList
              columns={headers}
              data={filteredIssues}
              totalItems={hasActiveFilters || Boolean(tableSearchTerm) ? filteredIssues.length : (data?.totalCount ?? 0)}
              currentPage={page}
              currentPageSize={pageSize}
              onPageChange={(newPage, newPageSize) => {
                setPage(newPage);
                setPageSize(newPageSize);
              }}
              handleIssueClick={handleIssueClick}
              closeView={() => {
                setSelectedIssue(undefined);
                setIsViewIssueDetail(false);
              }}
              selectedRowIds={selectedRowIds}
              onSelectRow={handleSelectRow}
              onSelectAll={handleSelectAll}
            />

        {showModal && (
          <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
            <IssueModal onClose={closeIssueModal} />
          </PermissionGuard>
        )}

        {showImportModal && (
          <PermissionGuard permission={PERMISSIONS.issueTrackerWrite}>
            <ImportIssuesModal onClose={closeImportModal} />
          </PermissionGuard>
        )}

        {showAssignModal && selectedIssueCodes.length > 0 && (
          <PermissionGuard permission={PERMISSIONS.issueTrackerAssign}>
            <AssignModal
              issueCodes={selectedIssueCodes}
              isReassign={areAllSelectedAssigned}
              onClose={() => setShowAssignModal(false)}
              onSuccess={() => setSelectedRowIds([])}
            />
          </PermissionGuard>
        )}
      </>
    </PermissionGuard>
  );
};

export default IssueTracker;
