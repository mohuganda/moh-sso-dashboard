import DataList from "../../../data-visualizer/src/pages/components/data-table/data-table.component.tsx";
import {Button, Dropdown, Popover, PopoverContent, ComboBox, Search, TreeView} from "@carbon/react";
import {Add, Filter, ChevronDown, Upload, Download} from "@carbon/react/icons";
import "./issue-tracker.scss";
import {useEffect, useState, useCallback, useRef, useMemo} from "react";
import {IssueModal} from "../component/issue-modal.component.tsx";
import {ImportIssuesModal} from "../component/import-issues-modal.component.tsx";
import { useGetIssuesQuery } from "@moh-sso/api";
import { headers, IMPORT_TEMPLATE_HEADERS } from "../lib/constants.ts";
import IssueDetail from "./issue-detail/issue-detail.component.tsx";
import { getAvailablePeriods, periodType } from "../../../data-visualizer/src/pages/Constants.tsx";
import {
  useGetDataSetsQuery,
  useLazyGetDataSetElementsQuery,
  type Dataset,
  type ThemeElement,
} from "../../../data-visualizer/src/pages/modals/data-model/data-model.ts";
import {useGetHierarchyQuery} from "../../../data-visualizer/src/pages/modals/orgunit/org-unit.ts";
import {OrgUnitNode} from "../component/tree-node.component.tsx";
import * as XLSX from "xlsx";

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

const IssueTracker = () => {
  const CURRENT_YEAR = new Date().getFullYear();

  const [showModal, setShowModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [issues, setIssues] = useState<Issue[]>([]);
  const [selectedIssue, setSelectedIssue] = useState<Issue>();
  const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);
  const [tableSearchTerm, setTableSearchTerm] = useState("");

  const { data, isLoading, error } = useGetIssuesQuery();

  // Period filters
  const [selectedYear, setSelectedYear] = useState(CURRENT_YEAR);
  const [selectedPeriodType, setSelectedPeriodType] = useState("Monthly");
  const [selectedPeriod, setSelectedPeriod] = useState("");
  const [availablePeriods, setAvailablePeriods] = useState(
    getAvailablePeriods("Monthly", CURRENT_YEAR.toString()),
  );
  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState(false);
  const periodPopoverRef = useRef<HTMLDivElement>(null);

  // Data filters
  const [selectedDataset, setSelectedDataset] = useState("");
  const [selectedDataElement, setSelectedDataElement] = useState("");
  const [dataElements, setDataElements] = useState<ThemeElement[]>([]);
  const [isDataPopoverOpen, setIsDataPopoverOpen] = useState(false);
  const dataPopoverRef = useRef<HTMLDivElement>(null);

  const { data: datasets } = useGetDataSetsQuery();
  const [triggerGetDataSetElements] = useLazyGetDataSetElementsQuery();

  // Organisation-unit filters
  const [selectedOrgUnit, setSelectedOrgUnit] = useState("");
  const [orgSearchTerm, setOrgSearchTerm] = useState("");
  const [isOrgPopoverOpen, setIsOrgPopoverOpen] = useState(false);
  const orgPopoverRef = useRef<HTMLDivElement>(null);

  const { data: hierarchyData } = useGetHierarchyQuery();

  const years = Array.from({ length: 10 }, (_, index) => CURRENT_YEAR - index);

  const close = () => {
    setShowModal(false);
  };

  const closeImportModal = () => {
    setShowImportModal(false);
  };

  const downloadTemplate = () => {
    const wb = XLSX.utils.book_new();
    const ws = XLSX.utils.aoa_to_sheet([
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
    ws["!cols"] = IMPORT_TEMPLATE_HEADERS.map(() => ({ wch: 22 }));
    XLSX.utils.book_append_sheet(wb, ws, "Issues Template");
    XLSX.writeFile(wb, "issue-import-template.xlsx");
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
      const rows =
        data?.data?.map((item) => ({
          id: item.issue_id.toString(),
          ...item,
        })) ?? [];

      setIssues(rows);
    }

    if (error) {
      console.error("Error encountered while fetching issues:", error);
    }
  }, [data, error, isLoading]);

  const filteredIssues = useMemo(() => {
    if (!tableSearchTerm.trim()) return issues;
    const term = tableSearchTerm.toLowerCase();
    return issues.filter(issue =>
      [issue.issue_code, issue.dataset, issue.data_element, issue.issue, issue.status, issue.org_unit, issue.date_reported, issue.issue_type]
        .some(field => field?.toLowerCase().includes(term))
    );
  }, [issues, tableSearchTerm]);

  const handleIssueClick = (issue) => {
    const selectedItem = issues?.find(item => item?.issue_id.toString() === issue?.id);
    if (selectedItem) {
      setSelectedIssue(selectedItem);
      setIsViewIssueDetail(true);
    }
  };

  const handleYearChange = ({ selectedItem }: SelectEvent<number>) => {
    if (selectedItem == null) {
      return;
    }

    setSelectedYear(selectedItem);

    const newPeriods = getAvailablePeriods(selectedPeriodType, selectedItem.toString());

    setAvailablePeriods(newPeriods);
    setSelectedPeriod("");
  };

  const handlePeriodTypeChange = ({
    selectedItem,
  }: SelectEvent<{ label: string; value: string }>) => {
    if (!selectedItem) {
      return;
    }

    setSelectedPeriodType(selectedItem.value);

    const newPeriods = getAvailablePeriods(selectedItem.value, selectedYear.toString());

    setAvailablePeriods(newPeriods);
    setSelectedPeriod("");
  };

  const handlePeriodChange = ({ selectedItem }: SelectEvent<{ label: string }>) => {
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

    const dataset = (datasets as Dataset[] | undefined)?.find(
      (item) => item.display_name === datasetName,
    );

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

  const nodeMatchesSearch = useCallback((node: any, term: string): boolean => {
    if (node.name.toLowerCase().includes(term.toLowerCase())) {
      return true;
    }

    if (node.children) {
      return node.children.some((child: any) => nodeMatchesSearch(child, term));
    }

    return false;
  }, []);

  const handleOrgSelect = (name: string) => {
    setSelectedOrgUnit(name);
  };

  const renderRecursive = (nodes: any[], idPrefix = "filter-org"): React.ReactNode[] => {
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

  const handleFilter = (overrides: Record<string, unknown> = {}) => {
    if (!data?.data) {
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

    let filtered = data.data.map((item) => ({
      id: item.issue_id.toString(),
      ...item,
    }));

    if (filters.period) {
      filtered = filtered.filter(
        (issue) => issue.time_period === filters.period || issue.time_Period === filters.period,
      );
    } else if (filters.year) {
      const year = filters.year.toString();

      filtered = filtered.filter(
        (issue) => issue.time_period?.includes(year) || issue.time_Period?.includes(year),
      );
    }

    if (filters.dataset) {
      filtered = filtered.filter((issue) => issue.dataset === filters.dataset);
    }

    if (filters.dataElement) {
      filtered = filtered.filter((issue) => issue.data_element === filters.dataElement);
    }

    if (filters.orgUnit) {
      filtered = filtered.filter((issue) => issue.org_unit === filters.orgUnit);
    }

    setIssues(filtered);
    setIsPeriodPopoverOpen(false);
    setIsDataPopoverOpen(false);
    setIsOrgPopoverOpen(false);
  };

  const handleReset = () => {
    setSelectedYear(CURRENT_YEAR);
    setSelectedPeriodType("Monthly");
    setSelectedPeriod("");
    setAvailablePeriods(getAvailablePeriods("Monthly", CURRENT_YEAR.toString()));

    setSelectedDataset("");
    setSelectedDataElement("");
    setDataElements([]);

    setSelectedOrgUnit("");
    setOrgSearchTerm("");

    if (data?.data) {
      const rows = data.data.map((item) => ({
        id: item.issue_id.toString(),
        ...item,
      }));

      setIssues(rows);
    }
  };

  if (isViewIssueDetail && selectedIssue) {
    return (
      <IssueDetail selectedIssue={selectedIssue} goToBack={() => setIsViewIssueDetail(false)} />
    );
  }

  return (
      <>
        { isViewIssueDetail && selectedIssue ? (
            <IssueDetail selectedIssue={selectedIssue} goToBack={() => setIsViewIssueDetail(false)}/>
        ) : (
            <>
              <div className="dv-toolbar issue-label-container">
                <div>
                  <span className="issue-label"> Registered Issues </span>
                </div>
                <div className="issue-toolbar-actions">
                  <Button
                      size="md"
                      kind="ghost"
                      renderIcon={Download}
                      onClick={downloadTemplate}
                  >
                    Template
                  </Button>
                  <Button
                      size="md"
                      kind="secondary"
                      renderIcon={Upload}
                      onClick={() => setShowImportModal(true)}
                  >
                    Import Issues
                  </Button>
                  <Button
                      size="md"
                      kind="primary"
                      renderIcon={Add}
                      className={`dwh-btn-width`}
                      onClick={()=> setShowModal(true)}
                  >
                    New Issue
                  </Button>
                </div>
              </div>

                  <PopoverContent className="filter-popover-content">
                    <div className="popover-inner org-filter-inner">
                      <Search
                        labelText="Search Org Unit"
                        placeholder="Search..."
                        value={orgSearchTerm}
                        onChange={(event) => setOrgSearchTerm(event.target.value)}
                        size="sm"
                      />

                      <div className="org-tree-container">
                        <TreeView label="Org Units" hideLabel>
                          {renderRecursive(hierarchyData || [], "filter-org")}
                        </TreeView>
                      </div>

                      <div className="popover-footer">
                        <Button
                          size="sm"
                          kind="ghost"
                          onClick={() => {
                            setSelectedOrgUnit("");
                            setOrgSearchTerm("");
                            handleFilter({ orgUnit: "" });
                          }}
                        >
                          Clear
                        </Button>

                        <Button size="sm" onClick={() => handleFilter()}>
                          Update
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="issue-table-search">
                  <Search
                    labelText="Search issues"
                    placeholder="Search by issue code, dataset, data element, issue, status, org unit..."
                    value={tableSearchTerm}
                    onChange={(e) => setTableSearchTerm(e.target.value)}
                    size="lg"
                    closeButtonLabelText="Clear search"
                  />
                </div>
                <DataList columns={headers} data={filteredIssues} handleIssueClick={handleIssueClick} closeView={() => setIsViewIssueDetail(false)}/>
              </div>
              {
                  showModal && <IssueModal onClose={close}/>
              }
              {
                  showImportModal && <ImportIssuesModal onClose={closeImportModal}/>
              }
            </>
        )}
      </>
  );
};

export default IssueTracker;
