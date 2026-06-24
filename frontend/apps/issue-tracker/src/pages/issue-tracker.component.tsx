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

import { useGetIssuesQuery } from "@moh-sso/api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";

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

const IssueTracker = () => {
  const currentYear = new Date().getFullYear();

  const [showModal, setShowModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);

  const [issues, setIssues] = useState<Issue[]>([]);
  const [selectedIssue, setSelectedIssue] = useState<Issue>();
  const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);

  const [tableSearchTerm, setTableSearchTerm] = useState("");

  const { data, isLoading, error } = useGetIssuesQuery();

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
    Boolean(selectedOrgUnit);

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

  const downloadIssues = () => {
    const headerRow = EXPORT_COLUMNS.map((col) => col.header);

    const dataRows = filteredIssues.map((issue) =>
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
    if (!isLoading) {
      setIssues(normalizeIssueRows(data as Issue[] | undefined));
    }

    if (error) {
      console.error("Error encountered while fetching issues:", error);
    }
  }, [data, error, isLoading]);

  const filteredIssues = useMemo(() => {
    const searchTerm = tableSearchTerm.trim().toLowerCase();

    if (!searchTerm) {
      return issues;
    }

    return issues.filter((issue) =>
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
    );
  }, [issues, tableSearchTerm]);

  const handleIssueClick = (row: { id?: string }) => {
    const selectedItem = issues.find((item) => String(item.issue_id) === String(row.id ?? ""));

    if (!selectedItem) {
      return;
    }

    setSelectedIssue(selectedItem);
    setIsViewIssueDetail(true);
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
    setSelectedOrgUnit(name);
  };

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
          onSelect={handleOrgSelect}
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

    setIssues(normalizeIssueRows(data as Issue[] | undefined));

    closeAllPopovers();
  };

  if (isViewIssueDetail && selectedIssue) {
    return (
      <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
        <IssueDetail
          selectedIssue={selectedIssue}
          goToBack={() => {
            setSelectedIssue(undefined);
            setIsViewIssueDetail(false);
          }}
        />
      </PermissionGuard>
    );
  }

  return (
    <PermissionGuard permission={PERMISSIONS.issueTrackerRead}>
      <>
        <div className="dv-toolbar issue-label-container">
          <div>
            <span className="issue-label">Registered Issues</span>
          </div>

          <div className="issue-toolbar-actions">
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
          handleIssueClick={handleIssueClick}
          closeView={() => {
            setSelectedIssue(undefined);
            setIsViewIssueDetail(false);
          }}
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
      </>
    </PermissionGuard>
  );
};

export default IssueTracker;
