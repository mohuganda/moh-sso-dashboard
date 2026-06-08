import {
  ComboBox,
  InlineLoading,
  InlineNotification,
  Modal,
  NumberInput,
  Search,
  TextArea,
  TreeView,
} from "@carbon/react";
import { ChevronDown, ChevronUp } from "@carbon/react/icons";
import {
  type Theme,
  type ThemeElement,
  getAvailablePeriods,
  periodType,
  useGetHierarchyQuery,
  useGetThemesQuery,
  useLazyGetThemeElementsQuery,
} from "@moh-sso/data-visualizer";
import { selectUser } from "@moh-sso/auth";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useSelector } from "react-redux";

import { IssueTypes, Priority } from "../lib/constants";
import { OrgUnitNode } from "./tree-node.component";
import type { Issue } from "@moh-sso/types";
import { useCreateIssueMutation, useUpdateIssueMutation } from "@moh-sso/api";

type IssueModalProps = {
  onClose: () => void;
  onSuccess?: () => void;
  selectedIssue?: Issue | null;
};

type OrgUnitTreeNode = {
  id: string;
  name: string;
  children?: OrgUnitTreeNode[];
};

type ComboBoxChange<T> = {
  selectedItem?: T | null;
};

const CURRENT_YEAR = new Date().getFullYear();
const DEFAULT_PERIOD_TYPE = "Quarterly";

function getErrorMessage(error: unknown, fallback: string) {
  if (!error) return fallback;

  if (typeof error === "object" && error !== null) {
    const err = error as {
      data?: { message?: string };
      error?: string;
      message?: string;
    };

    return err.data?.message || err.error || err.message || fallback;
  }

  return fallback;
}

function nodeMatchesSearch(node: OrgUnitTreeNode, term: string): boolean {
  const normalizedTerm = term.trim().toLowerCase();

  if (!normalizedTerm) return true;

  if (node.name.toLowerCase().includes(normalizedTerm)) {
    return true;
  }

  return node.children?.some((child) => nodeMatchesSearch(child, term)) ?? false;
}

export function IssueModal({ onClose, onSuccess, selectedIssue }: IssueModalProps) {
  const isEdit = Boolean(selectedIssue);
  const user = useSelector(selectUser);

  const [selectedDataset, setSelectedDataset] = useState(selectedIssue?.dataset ?? "");
  const [selectedDataElement, setSelectedDataElement] = useState(selectedIssue?.data_element ?? "");
  const [selectedOrgUnit, setSelectedOrgUnit] = useState(selectedIssue?.org_unit ?? "");
  const [description, setDescription] = useState(selectedIssue?.issue ?? "");
  const [selectedIssueType, setSelectedIssueType] = useState(selectedIssue?.issue_type ?? "");
  const [priority, setPriority] = useState(selectedIssue?.priority ?? "");
  const [severity, setSeverity] = useState(selectedIssue?.severity ?? "");

  const [isOrgExpanded, setIsOrgExpanded] = useState(false);
  const [orgSearchTerm, setOrgSearchTerm] = useState("");

  const [selectedYear, setSelectedYear] = useState(CURRENT_YEAR);
  const [selectedPeriodType, setSelectedPeriodType] = useState(
    selectedIssue ? "" : DEFAULT_PERIOD_TYPE,
  );
  const [selectedPeriod, setSelectedPeriod] = useState(
    selectedIssue?.time_Period || selectedIssue?.time_period || "",
  );

  const [formError, setFormError] = useState<string | null>(null);

  const {
    data: themes = [],
    isLoading: isLoadingThemes,
    isError: isThemesError,
    error: themesError,
  } = useGetThemesQuery();

  const [
    triggerGetThemeElements,
    {
      data: themeElements = [],
      isFetching: isFetchingThemeElements,
      isError: isThemeElementsError,
      error: themeElementsError,
    },
  ] = useLazyGetThemeElementsQuery();

  const {
    data: hierarchyData,
    isLoading: isLoadingHierarchy,
    isError: isHierarchyError,
    error: hierarchyError,
  } = useGetHierarchyQuery();

  const [createIssue, createState] = useCreateIssueMutation();
  const [updateIssue, updateState] = useUpdateIssueMutation();

  const isSubmitting = createState.isLoading || updateState.isLoading;

  const datasets = useMemo(() => {
    return themes.map((item: Theme) => item.theme_name).filter(Boolean);
  }, [themes]);

  const dataElements = useMemo(() => {
    return themeElements.map((item: ThemeElement) => item.data_element_short_name).filter(Boolean);
  }, [themeElements]);

  const availablePeriods = useMemo(() => {
    if (!selectedPeriodType) return [];

    return getAvailablePeriods(selectedPeriodType, selectedYear);
  }, [selectedPeriodType, selectedYear]);

  const periodItems = useMemo(() => {
    return availablePeriods.map((item) => item.label);
  }, [availablePeriods]);

  const orgUnits = useMemo<OrgUnitTreeNode[]>(() => {
    if (Array.isArray(hierarchyData)) {
      return hierarchyData;
    }

    if (hierarchyData && typeof hierarchyData === "object" && "children" in hierarchyData) {
      return [hierarchyData as OrgUnitTreeNode];
    }

    return [];
  }, [hierarchyData]);

  const canSubmit = Boolean(
    selectedDataset &&
    selectedDataElement &&
    selectedOrgUnit &&
    selectedIssueType &&
    selectedPeriod &&
    description.trim(),
  );

  useEffect(() => {
    if (!selectedDataset || !themes.length) return;

    const themeId = themes.find((item: Theme) => item.theme_name === selectedDataset)?.theme_id;

    if (!themeId) return;

    triggerGetThemeElements(themeId);
  }, [selectedDataset, themes, triggerGetThemeElements]);

  const renderRecursive = useCallback(
    (nodes: OrgUnitTreeNode[]) => {
      if (!Array.isArray(nodes)) return [];

      return nodes
        .filter((node) => nodeMatchesSearch(node, orgSearchTerm))
        .map((node) => (
          <OrgUnitNode
            key={node.id}
            node={node}
            searchTerm={orgSearchTerm}
            selectedOrgUnit={selectedOrgUnit}
            onSelect={(name: string) => {
              setSelectedOrgUnit(name);
              setOrgSearchTerm("");
              setTimeout(() => setIsOrgExpanded(false), 150);
            }}
            renderRecursive={renderRecursive}
          />
        ));
    },
    [orgSearchTerm, selectedOrgUnit],
  );

  const handleDatasetChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSelectedDataset(selectedItem ?? "");
    setSelectedDataElement("");
  };

  const handleDataElementChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSelectedDataElement(selectedItem ?? "");
  };

  const handleIssueTypeChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSelectedIssueType(selectedItem ?? "");
  };

  const handlePriorityChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setPriority(selectedItem ?? "");
  };

  const handleSeverityChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSeverity(selectedItem ?? "");
  };

  const handlePeriodTypeChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSelectedPeriodType(selectedItem ?? "");
    setSelectedPeriod("");
  };

  const handleSelectedPeriodChange = ({ selectedItem }: ComboBoxChange<string>) => {
    setSelectedPeriod(selectedItem ?? "");
  };

  const handleYearChange = (_event: unknown, data: { value: number | string }) => {
    const nextYear = Number(data.value);

    if (Number.isNaN(nextYear)) return;

    setSelectedYear(nextYear);
    setSelectedPeriod("");
  };

  const handleSubmit = async () => {
    setFormError(null);

    if (!canSubmit) {
      setFormError("Please fill all required fields before submitting.");
      return;
    }

    const payload = {
      dataset: selectedDataset,
      data_element: selectedDataElement,
      org_unit: selectedOrgUnit,
      issue: description.trim(),
      issue_type: selectedIssueType,
      reported_by: selectedIssue?.reported_by || user?.username || "",
      updated_by: isEdit ? user?.username || "" : "",
      priority,
      severity,
      time_period: selectedPeriod,
    };

    try {
      if (isEdit) {
        await updateIssue({
          id: selectedIssue?.issue_code ?? "",
          body: payload,
        }).unwrap();
      } else {
        await createIssue(payload).unwrap();
      }

      onSuccess?.();
      onClose();
    } catch (error) {
      setFormError(
        getErrorMessage(error, isEdit ? "Failed to update issue." : "Failed to create issue."),
      );
    }
  };

  return (
    <Modal
      aria-label="issue-modal"
      open
      modalHeading={
        isEdit ? `Edit Issue: ${selectedIssue?.issue_code ?? ""}` : "Register a New Issue"
      }
      primaryButtonText={isSubmitting ? "Saving..." : "Submit Issue"}
      secondaryButtonText="Cancel"
      primaryButtonDisabled={isSubmitting || !canSubmit}
      onRequestClose={onClose}
      onRequestSubmit={handleSubmit}
      size="lg"
    >
      <p style={{ marginBottom: "2rem" }}>
        Register a new issue relating to data anomalies and their possible causes.
      </p>

      {formError && (
        <InlineNotification
          kind="error"
          lowContrast
          title="Unable to save issue"
          subtitle={formError}
          style={{ marginBottom: "1rem" }}
        />
      )}

      {(isThemesError || isHierarchyError || isThemeElementsError) && (
        <InlineNotification
          kind="warning"
          lowContrast
          title="Some data could not be loaded"
          subtitle={
            getErrorMessage(themesError, "") ||
            getErrorMessage(hierarchyError, "") ||
            getErrorMessage(themeElementsError, "") ||
            "Please refresh and try again."
          }
          style={{ marginBottom: "1rem" }}
        />
      )}

      <div style={{ marginBottom: "24px" }}>
        <p className="cds--label">Organisation Unit</p>

        <button
          type="button"
          onClick={() => setIsOrgExpanded((current) => !current)}
          style={{
            width: "100%",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "0 1rem",
            height: "40px",
            background: "#f4f4f4",
            border: 0,
            borderBottom: "1px solid #8d8d8d",
            cursor: "pointer",
            textAlign: "left",
          }}
        >
          <span style={{ color: selectedOrgUnit ? "#161616" : "#6f6f6f" }}>
            {selectedOrgUnit || "Select Organisation Unit"}
          </span>

          {isOrgExpanded ? <ChevronUp /> : <ChevronDown />}
        </button>

        {isOrgExpanded && (
          <div
            style={{
              border: "1px solid #e0e0e0",
              background: "white",
              maxHeight: "250px",
              overflowY: "auto",
              marginTop: "2px",
              padding: "8px",
            }}
          >
            <Search
              labelText="Search organisation unit"
              size="md"
              id="org-unit-search"
              placeholder="Search for org unit..."
              value={orgSearchTerm}
              onChange={(event) => setOrgSearchTerm(event.target.value)}
              style={{ marginBottom: "8px" }}
            />

            {isLoadingHierarchy ? (
              <InlineLoading description="Loading organisation units..." />
            ) : (
              <TreeView label="Org Units" hideLabel>
                {renderRecursive(orgUnits)}
              </TreeView>
            )}
          </div>
        )}
      </div>

      <div style={{ marginBottom: "24px" }}>
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-dataset-combobox"
          onChange={handleDatasetChange}
          items={datasets}
          titleText="Datasets"
          selectedItem={selectedDataset}
          disabled={isLoadingThemes}
        />

        {isLoadingThemes && (
          <div style={{ marginTop: "0.5rem" }}>
            <InlineLoading description="Loading datasets..." />
          </div>
        )}
      </div>

      <div style={{ marginBottom: "24px" }}>
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-data-element-combobox"
          onChange={handleDataElementChange}
          items={dataElements}
          titleText="Data Elements"
          selectedItem={selectedDataElement}
          disabled={!selectedDataset || isFetchingThemeElements}
        />

        {isFetchingThemeElements && (
          <div style={{ marginTop: "0.5rem" }}>
            <InlineLoading description="Loading data elements..." />
          </div>
        )}
      </div>

      <div
        style={{
          marginBottom: "24px",
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "1rem",
        }}
      >
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-period-type-combobox"
          onChange={handlePeriodTypeChange}
          items={periodType.map((item) => item.label)}
          titleText="Period Type"
          selectedItem={selectedPeriodType}
        />

        <NumberInput
          id="issue-period-year-input"
          invalidText="Input is not a valid year"
          label="Year"
          locale="en"
          max={CURRENT_YEAR}
          min={1900}
          size="md"
          step={1}
          type="number"
          value={selectedYear}
          onChange={handleYearChange}
        />
      </div>

      <div style={{ marginBottom: "24px" }}>
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-period-combobox"
          onChange={handleSelectedPeriodChange}
          items={periodItems}
          titleText="Available Periods"
          selectedItem={selectedPeriod}
          disabled={!selectedPeriodType}
        />
      </div>

      <div style={{ marginBottom: "24px" }}>
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-type-combobox"
          onChange={handleIssueTypeChange}
          items={IssueTypes}
          titleText="Issue Type"
          selectedItem={selectedIssueType}
        />
      </div>

      <div
        style={{
          marginBottom: "24px",
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "1rem",
        }}
      >
        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-priority-combobox"
          onChange={handlePriorityChange}
          items={Priority}
          titleText="Priority"
          selectedItem={priority}
        />

        <ComboBox
          allowCustomValue
          autoAlign
          id="issue-severity-combobox"
          onChange={handleSeverityChange}
          items={Priority}
          titleText="Severity"
          selectedItem={severity}
        />
      </div>

      <TextArea
        id="issue-description-textarea"
        labelText="Issue Description"
        style={{ marginBottom: "24px" }}
        value={description}
        onChange={(event) => setDescription(event.target.value)}
        rows={7}
      />
    </Modal>
  );
}
