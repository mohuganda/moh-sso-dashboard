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

import {
  DataTablePagination,
  DataTableShell,
  RowActionsCell,
  TableStatusTag,
  useHeaderPanel,
} from "@moh-sso/ui";
import { ValidationRuleActionsMenu } from "../components/validation-rule-actions-menu";
import { ValidationRuleDetailsPanel } from "../components/validation-rule-details-panel";
import { ValidationRulePanel } from "../components/validation-rule-panel";
import type { ValidationRule } from "../types";

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

export default function DataValidationPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const [rules, setRules] = useState<ValidationRule[]>(builtInRules);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

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
          initialCode={nextCustomCode}
          onSubmit={(rule) => {
            setRules((current) => [rule, ...current]);
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
          initialRule={rule}
          onSubmit={(updatedRule) => {
            setRules((current) =>
              current.map((item) => (item.id === updatedRule.id ? updatedRule : item)),
            );
          }}
          onClose={closePanel}
        />
      ),
    });
  };

  const deleteRule = (rule: ValidationRule) => {
    setRules((current) => current.filter((item) => item.id !== rule.id));
  };

  return (
    <div className="data-validation-page">
      <DataTableShell
        title="Validation rules"
        description="Built-in and custom rules applied during validation."
        rows={rows}
        headers={headers}
        getRowId={(row) => row.id}
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
                        .map((header) => (
                          <TableHeader {...getHeaderProps({ header })} key={header.key}>
                            {header.header}
                          </TableHeader>
                        ))}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {rows.map((row) => {
                      const rule = row.cells.find((cell) => cell.info.header === "raw")
                        ?.value as ValidationRule;

                      return (
                        <TableRow {...getRowProps({ row })} key={row.id}>
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
                                    onDelete={() => deleteRule(rule)}
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
