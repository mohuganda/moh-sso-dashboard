import {
  Button,
  DataTable,
  Search,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";
import { Add } from "@carbon/react/icons";
import { useEffect, useMemo, useState } from "react";

import { useImportValidationRulesMutation, useListValidationRulesQuery } from "../api";
import {
  DataTablePagination,
  DataTableShell,
  ErrorState,
  RowActionsCell,
  TableStatusTag,
  useHeaderPanel,
  useToast,
} from "@moh-sso/ui";
import { ValidationRuleActionsMenu } from "../components/validation-rule-actions-menu";
import { ValidationRuleDetailsPanel } from "../components/validation-rule-details-panel";
import { ValidationRulePanel } from "../components/validation-rule-panel";
import type { ValidationRule, ValidationRulePayload } from "../types";

import "./data-validation.page.scss";

const builtInRules: ValidationRule[] = [
  "malaria",
  "diarrhoea",
  "pneumonia",
  "danger_signs",
  "pregnancy",
  "family_planning",
  "activity",
  "commodities",
  "outliers",
  "timeliness",
  "completeness",
].map((code) => ({
  id: `builtin-${code}`,
  type: "builtin",
  code,
  description: "Built-in rule category",
}));

const headers = [
  { key: "type", header: "Type" },
  { key: "code", header: "Code / Category" },
  { key: "severity", header: "Severity" },
  { key: "table", header: "Table" },
  { key: "column", header: "Column" },
  { key: "operator", header: "Operator" },
  { key: "description", header: "Description" },
  { key: "actions", header: "" },
  { key: "raw", header: "" },
];

function normalize(value: string) {
  return value.trim().toLowerCase();
}

function toRulePayload(rule: ValidationRule): ValidationRulePayload {
  const payload: ValidationRulePayload = {
    table_id: rule.table || undefined,
    category: "custom",
    code: rule.code,
    severity: rule.severity ?? "error",
    description: rule.description,
    column: rule.column ?? "",
    op: rule.operator ?? "contains",
  };

  if (rule.compareTo === "column") {
    payload.value_column = rule.value || undefined;
  } else {
    payload.value = rule.value || undefined;
  }

  return payload;
}

export default function DataValidationPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const toast = useToast();
  const {
    data: customRules = [],
    isLoading,
    isFetching,
    isError,
    refetch,
  } = useListValidationRulesQuery();
  const [importValidationRules, { isLoading: isSaving }] = useImportValidationRulesMutation();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const rules = useMemo(() => [...builtInRules, ...customRules], [customRules]);

  useEffect(() => {
    setPage(1);
  }, [search]);

  const filteredRules = useMemo(() => {
    const term = normalize(search);

    if (!term) {
      return rules;
    }

    return rules.filter((rule) =>
      [
        rule.type,
        rule.code,
        rule.severity,
        rule.table,
        rule.column,
        rule.operator,
        rule.description,
      ]
        .filter(Boolean)
        .some((value) => normalize(String(value)).includes(term)),
    );
  }, [rules, search]);

  const paginatedRules = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredRules.slice(start, start + pageSize);
  }, [filteredRules, page, pageSize]);

  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil(filteredRules.length / pageSize));

    if (page > maxPage) {
      setPage(maxPage);
    }
  }, [filteredRules.length, page, pageSize]);

  const rows = paginatedRules.map((rule) => ({
    id: rule.id,
    type: rule.type,
    code: rule.code,
    severity: rule.severity ?? "",
    table: rule.table ?? "",
    column: rule.column ?? "",
    operator: rule.operator ?? "",
    description: rule.description,
    actions: "",
    raw: rule,
  }));

  const nextCustomCode = `CUS-${String(
    rules.filter((rule) => rule.type === "custom").length + 1,
  ).padStart(2, "0")}`;

  const openCreatePanel = () => {
    openPanel({
      title: "Add custom rule",
      size: "md",
      content: (
        <ValidationRulePanel
          key={`validation-rule-${nextCustomCode}`}
          isSubmitting={isSaving}
          initialCode={nextCustomCode}
          onSubmit={async (rule) => {
            try {
              const result = await importValidationRules([toRulePayload(rule)]).unwrap();
              if (result.skipped > 0) {
                toast.error("Validation rule not saved", result.errors[0]?.message ?? "Check the rule.");
                return;
              }
              toast.success("Validation rule added", `${rule.code} is now available.`);
              closePanel();
            } catch {
              toast.error("Validation rule not saved", "Please try again.");
            }
          }}
          onClose={closePanel}
        />
      ),
    });
  };

  const openViewPanel = (rule: ValidationRule) => {
    openPanel({
      title: `Rule: ${rule.code}`,
      size: "md",
      content: <ValidationRuleDetailsPanel rule={rule} />,
    });
  };

  const openEditPanel = (rule: ValidationRule) => {
    openPanel({
      title: `Edit rule: ${rule.code}`,
      size: "md",
      content: (
        <ValidationRulePanel
          key={`edit-rule-${rule.id}`}
          mode="edit"
          isSubmitting={isSaving}
          initialRule={rule}
          onSubmit={async (updatedRule) => {
            try {
              const result = await importValidationRules([toRulePayload(updatedRule)]).unwrap();
              if (result.skipped > 0) {
                toast.error("Validation rule not saved", result.errors[0]?.message ?? "Check the rule.");
                return;
              }
              toast.success("Validation rule updated", `${updatedRule.code} was saved.`);
              closePanel();
            } catch {
              toast.error("Validation rule not saved", "Please try again.");
            }
          }}
          onClose={closePanel}
        />
      ),
    });
  };

  const tableState = isError ? (
    <ErrorState
      title="Unable to load validation rules"
      description="Refresh the table and try again."
      primaryAction={{
        label: "Refresh",
        onClick: refetch,
      }}
    />
  ) : undefined;

  return (
    <div className="data-validation-page">
      <DataTableShell
        title="Validation rules"
        description="Built-in and custom rules applied during validation."
        rows={rows}
        headers={headers}
        getRowId={(row) => row.id}
        isLoading={isLoading || isFetching}
        tableState={tableState}
        emptyTitle="No validation rules"
        emptyDescription="Try changing your search or add a custom rule."
        filters={
          <div className="data-validation-filters">
            <Search
              id="data-validation-search"
              labelText="Search validation rules"
              placeholder="Search validation rules..."
              size="lg"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
            <Button kind="primary" renderIcon={Add} onClick={openCreatePanel}>
              Add custom rule
            </Button>
          </div>
        }
      >
        {({ rows, headers }) => (
          <DataTable rows={rows} headers={headers}>
            {({ rows, headers, getHeaderProps, getRowProps }) => (
              <>
                <Table size="lg" className="data-validation-page__table">
                  <TableHead>
                    <TableRow>
                      {headers
                        .filter((header) => header.key !== "raw")
                        .map((header) => {
                          const { key, ...headerProps } = getHeaderProps({ header });

                          return (
                            <TableHeader key={key} {...headerProps}>
                              {header.header}
                            </TableHeader>
                          );
                        })}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {rows.map((row) => {
                      const rule = row.cells.find((cell) => cell.info.header === "raw")
                        ?.value as ValidationRule;

                      const { key, ...rowProps } = getRowProps({ row });

                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "raw") return null;

                            if (cell.info.header === "type") {
                              return (
                                <TableCell key={cell.id}>
                                  <TableStatusTag
                                    status={String(cell.value)}
                                    kind={rule.type === "builtin" ? "gray" : "blue"}
                                  />
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "severity") {
                              return (
                                <TableCell key={cell.id}>
                                  {rule.severity ? (
                                    <TableStatusTag
                                      status={rule.severity}
                                      kind={
                                        rule.severity === "error"
                                          ? "red"
                                          : rule.severity === "warning"
                                            ? "magenta"
                                            : "blue"
                                      }
                                    />
                                  ) : (
                                    "—"
                                  )}
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "actions") {
                              return (
                                <RowActionsCell key={cell.id}>
                                  <ValidationRuleActionsMenu
                                    rule={rule}
                                    onView={() => openViewPanel(rule)}
                                    onEdit={() => openEditPanel(rule)}
                                  />
                                </RowActionsCell>
                              );
                            }

                            return <TableCell key={cell.id}>{cell.value || "—"}</TableCell>;
                          })}
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>

                <DataTablePagination
                  page={page}
                  pageSize={pageSize}
                  totalItems={filteredRules.length}
                  onChange={({ page, pageSize }) => {
                    setPage(page);
                    setPageSize(pageSize);
                  }}
                />
              </>
            )}
          </DataTable>
        )}
      </DataTableShell>
    </div>
  );
}
