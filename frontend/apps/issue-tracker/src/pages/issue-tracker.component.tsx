import DataList from "../../../data-visualizer/src/pages/components/data-table/data-table.component.tsx";
import {Button, Dropdown, Popover, PopoverContent, ComboBox, Search, TreeView} from "@carbon/react";
import {Add, Filter, ChevronDown} from "@carbon/react/icons";
import "./issue-tracker.scss";
import {useEffect, useState, useCallback, useRef} from "react";
import {IssueModal} from "../component/issue-modal.component.tsx";
import { useGetIssuesQuery } from "@moh-sso/api";
import { headers } from "../lib/constants.ts";
import IssueDetail from "./issue-detail/issue-detail.component.tsx";
import {getAvailablePeriods, periodType} from "../../../data-visualizer/src/pages/Constants.tsx";
import {
  useGetDataSetsQuery,
  useLazyGetDataSetElementsQuery,
  type ThemeElement, type Dataset
} from "../../../data-visualizer/src/pages/modals/data-model/data-model.ts";
import {useGetHierarchyQuery} from "../../../data-visualizer/src/pages/modals/orgunit/org-unit.ts";
import {OrgUnitNode} from "../component/tree-node.component.tsx";

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
}

const IssueTracker = () => {
  const CURRENT_YEAR = new Date().getFullYear();
  const [showModal, setShowModal] = useState(false);
  const [issues, setIssues] = useState<Issue[]>([]);
  const { data, isLoading, error } = useGetIssuesQuery();
  const [selectedIssue, setSelectedIssue] = useState<Issue>();
  const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);

  // Period Filters
  const [selectedYear, setSelectedYear] = useState(CURRENT_YEAR);
  const [selectedPeriodType, setSelectedPeriodType] = useState("Monthly");
  const [selectedPeriod, setSelectedPeriod] = useState("");
  const [availablePeriods, setAvailablePeriods] = useState(getAvailablePeriods("Monthly", CURRENT_YEAR.toString()));
  const [isPeriodPopoverOpen, setIsPeriodPopoverOpen] = useState(false);
  const periodPopoverRef = useRef<HTMLDivElement>(null);

  // Data Filters
  const [selectedDataset, setSelectedDataset] = useState("");
  const [selectedDataElement, setSelectedDataElement] = useState("");
  const [dataElements, setDataElements] = useState<ThemeElement[]>([]);
  const [isDataPopoverOpen, setIsDataPopoverOpen] = useState(false);
  const dataPopoverRef = useRef<HTMLDivElement>(null);
  const { data: datasets } = useGetDataSetsQuery();
  const [ triggerGetDataSetElements ] = useLazyGetDataSetElementsQuery();

  // Org Unit Filters
  const [selectedOrgUnit, setSelectedOrgUnit] = useState("");
  const [orgSearchTerm, setOrgSearchTerm] = useState("");
  const [isOrgPopoverOpen, setIsOrgPopoverOpen] = useState(false);
  const orgPopoverRef = useRef<HTMLDivElement>(null);
  const { data: hierarchyData } = useGetHierarchyQuery();

  const years = Array.from({ length: 10 }, (_, i) => CURRENT_YEAR - i);

  const close = () => {
    setShowModal(false);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (periodPopoverRef.current && !periodPopoverRef.current.contains(event.target as Node)) {
        setIsPeriodPopoverOpen(false);
      }
      if (dataPopoverRef.current && !dataPopoverRef.current.contains(event.target as Node)) {
        setIsDataPopoverOpen(false);
      }
      if (orgPopoverRef.current && !orgPopoverRef.current.contains(event.target as Node)) {
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
      const rows = data?.data?.map(item => ({
        id: item.issue_id.toString(),
        ...item
      }));
      setIssues(rows || []);
    }
    if (error) {
      console.error("Error Encountered while fetching issues:: " + error)
    }
  }, [data, error, isLoading]);

  const handleIssueClick = (issue) => {
    const selectedItem = issues?.find(item => item?.issue_id.toString() === issue?.id);
    if (selectedItem) {
      setSelectedIssue(selectedItem);
      setIsViewIssueDetail(true);
    }
  };

  const handleYearChange = ({ selectedItem }) => {
    if (selectedItem) {
      setSelectedYear(selectedItem);
      const newPeriods = getAvailablePeriods(selectedPeriodType, selectedItem.toString());
      setAvailablePeriods(newPeriods);
      setSelectedPeriod("");
    }
  };

  const handlePeriodTypeChange = ({ selectedItem }) => {
    if (selectedItem) {
      setSelectedPeriodType(selectedItem.value);
      const newPeriods = getAvailablePeriods(selectedItem.value, selectedYear.toString());
      setAvailablePeriods(newPeriods);
      setSelectedPeriod("");
    }
  };

  const handlePeriodChange = ({ selectedItem }) => {
    if (selectedItem) {
      setSelectedPeriod(selectedItem.label);
    }
  };

  const handleDatasetChange = async (event) => {
    const datasetName = event?.selectedItem;
    setSelectedDataset(datasetName);
    setSelectedDataElement("");

    if (!datasetName) {
      setDataElements([]);
      return;
    }

    const theme = (datasets as Dataset[])?.find(t => t.display_name === datasetName);
    if (theme) {
      try {
        const elements = await triggerGetDataSetElements(theme.dataset_id).unwrap();
        setDataElements(elements);
      } catch (err) {
        console.error("Failed to fetch data elements:", err);
      }
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

  const renderRecursive = (nodes: any[], idPrefix: string = "filter-org") => {
    if (!nodes || !Array.isArray(nodes)) return [];

    return nodes
      .filter(node => !orgSearchTerm || nodeMatchesSearch(node, orgSearchTerm))
      .map(node => (
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

  const handleFilter = (overrides: any = {}) => {
    if (!data?.data) return;

    const filters = {
      period: selectedPeriod,
      year: selectedYear,
      dataset: selectedDataset,
      dataElement: selectedDataElement,
      orgUnit: selectedOrgUnit,
      ...overrides
    };

    let filtered = data.data.map(item => ({
      id: item.issue_id.toString(),
      ...item
    }));

    if (filters.period) {
      filtered = filtered.filter(issue => 
        issue.time_period === filters.period || issue.time_Period === filters.period
      );
    } else if (filters.year) {
      const yearStr = filters.year.toString();
      filtered = filtered.filter(issue => 
        (issue.time_period && issue.time_period.includes(yearStr)) || 
        (issue.time_Period && issue.time_Period.includes(yearStr))
      );
    }

    if (filters.dataset) {
      filtered = filtered.filter(issue => issue.dataset === filters.dataset);
    }

    if (filters.dataElement) {
      filtered = filtered.filter(issue => issue.data_element === filters.dataElement);
    }

    if (filters.orgUnit) {
      filtered = filtered.filter(issue => issue.org_unit === filters.orgUnit);
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
          const rows = data.data.map(item => ({
              id: item.issue_id.toString(),
              ...item
          }));
          setIssues(rows);
      }
  };

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


              <div className="issue-container">
                <div className="issue-filter-container">
                  <div className="issue-filters">
                    <div className="filter-popovers">
                      {/* Org Unit Filter Popover */}
                      <div ref={orgPopoverRef}>
                        <Popover open={isOrgPopoverOpen} align="bottom-left">
                          <Button
                              kind="ghost"
                              size="md"
                              onClick={() => setIsOrgPopoverOpen(!isOrgPopoverOpen)}
                              renderIcon={ChevronDown}
                          >
                            Org Unit: {selectedOrgUnit || "All"}
                          </Button>
                          <PopoverContent className="filter-popover-content">
                            <div className="popover-inner org-filter-inner">
                              <Search
                                  labelText="Search Org Unit"
                                  placeholder="Search..."
                                  value={orgSearchTerm}
                                  onChange={(e) => setOrgSearchTerm(e.target.value)}
                                  size="sm"
                              />
                              <div className="org-tree-container">
                                <TreeView label="Org Units" hideLabel>
                                  {renderRecursive(hierarchyData || [], "filter-org")}
                                </TreeView>
                              </div>
                              <div className="popover-footer">
                                <Button size="sm" kind="ghost" onClick={() => {
                                  setSelectedOrgUnit("");
                                  setOrgSearchTerm("");
                                  handleFilter({ orgUnit: "" });
                                }}>Clear</Button>
                                <Button size="sm" onClick={() => handleFilter()}>Update</Button>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
                      </div>

                      {/* Data Filter Popover */}
                      <div ref={dataPopoverRef}>
                        <Popover open={isDataPopoverOpen} align="bottom-left">
                          <Button
                              kind="ghost"
                              size="md"
                              onClick={() => setIsDataPopoverOpen(!isDataPopoverOpen)}
                              renderIcon={ChevronDown}
                          >
                            Data: {selectedDataset || "All"}
                          </Button>
                          <PopoverContent className="filter-popover-content">
                            <div className="popover-inner">
                              <ComboBox
                                  id="dataset-filter"
                                  titleText="Dataset"
                                  placeholder="Select Dataset"
                                  items={(datasets as Dataset[])?.map(t => t.display_name) || []}
                                  selectedItem={selectedDataset}
                                  onChange={handleDatasetChange}
                              />
                              <ComboBox
                                  id="dataelement-filter"
                                  titleText="Data Element"
                                  placeholder="Select Data Element"
                                  items={dataElements.map(e => e.data_element_short_name)}
                                  selectedItem={selectedDataElement}
                                  onChange={({ selectedItem }) => setSelectedDataElement(selectedItem || "")}
                              />
                              <div className="popover-footer">
                                <Button size="sm" kind="ghost" onClick={() => {
                                  setSelectedDataset("");
                                  setSelectedDataElement("");
                                  setDataElements([]);
                                  handleFilter({ dataset: "", dataElement: "" });
                                }}>Clear</Button>
                                <Button size="sm" onClick={() => handleFilter()}>Update</Button>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
                      </div>

                      {/* Period Filter Popover */}
                      <div ref={periodPopoverRef}>
                        <Popover open={isPeriodPopoverOpen} align="bottom-left">
                          <Button
                              kind="ghost"
                              size="md"
                              onClick={() => setIsPeriodPopoverOpen(!isPeriodPopoverOpen)}
                              renderIcon={ChevronDown}
                          >
                            Period: {selectedPeriod || selectedYear}
                          </Button>
                          <PopoverContent className="filter-popover-content">
                            <div className="popover-inner">
                              <Dropdown
                                  id="year-filter"
                                  titleText="Year"
                                  label="Select Year"
                                  items={years}
                                  selectedItem={selectedYear}
                                  onChange={handleYearChange}
                              />
                              <Dropdown
                                  id="period-type-filter"
                                  titleText="Period Type"
                                  label="Select Period Type"
                                  items={periodType}
                                  itemToString={(item) => item?.label ?? ""}
                                  selectedItem={periodType.find(p => p.value === selectedPeriodType)}
                                  onChange={handlePeriodTypeChange}
                              />
                              <Dropdown
                                  id="period-filter"
                                  titleText="Period"
                                  label="Select Period"
                                  items={availablePeriods}
                                  itemToString={(item) => item?.label ?? ""}
                                  selectedItem={availablePeriods.find(p => p.label === selectedPeriod)}
                                  onChange={handlePeriodChange}
                              />
                              <div className="popover-footer">
                                <Button size="sm" kind="ghost" onClick={() => {
                                  setSelectedYear(CURRENT_YEAR);
                                  setSelectedPeriodType("Monthly");
                                  setSelectedPeriod("");
                                  setAvailablePeriods(getAvailablePeriods("Monthly", CURRENT_YEAR.toString()));
                                  setSelectedDataset("");
                                  setSelectedDataElement("");
                                  setDataElements([]);
                                  handleFilter({
                                    period: "",
                                    dataset: "",
                                    dataElement: ""
                                  });
                                }}>Reset</Button>
                                <Button size="sm" onClick={() => handleFilter()}>Update</Button>
                              </div>
                            </div>
                          </PopoverContent>
                        </Popover>
                      </div>
                    </div>

                    <div className="filter-buttons">
                      <Button
                          size="md"
                          kind="secondary"
                          renderIcon={Filter}
                          onClick={handleFilter}
                      >
                        Filter
                      </Button>
                      <Button
                          size="md"
                          kind="ghost"
                          onClick={handleReset}
                      >
                        Reset
                      </Button>
                    </div>
                  </div>
                </div>
                <DataList columns={headers} data={issues} handleIssueClick={handleIssueClick} closeView={() => setIsViewIssueDetail(false)}/>
              </div>
              {
                  showModal && <IssueModal onClose={close}/>
              }
            </>
        )}
      </>
  );
};

export default IssueTracker;
