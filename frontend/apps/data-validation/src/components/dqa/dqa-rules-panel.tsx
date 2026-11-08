import {
  Button,
  ComposedModal,
  DataTable,
  ModalBody,
  ModalHeader,
  Select,
  SelectItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";
import { Add, TrashCan } from "@carbon/react/icons";
import { useState } from "react";

import { TableStatusTag, useToast } from "@moh-sso/ui";

import {
  useDeleteDQARuleMutation,
  useListDQARulesQuery,
  useListDQATablesQuery,
  useUpsertDQARuleMutation,
} from "../../api";
import type { DQARule, DQARuleInput } from "../../dqa.types";
import { DQARuleForm } from "./dqa-rule-form";

const headers = [
  { key: "enabled", header: "" },
  { key: "code", header: "Code" },
  { key: "table_id", header: "Table" },
  { key: "type", header: "Type" },
  { key: "category", header: "Category" },
  { key: "name", header: "Name" },
  { key: "actions", header: "" },
];

type RuleEditorState = { mode: "create" } | { mode: "edit"; rule: DQARule };

export function DQARulesPanel() {
  const toast = useToast();
  const [editor, setEditor] = useState<RuleEditorState | null>(null);
  const { data: tables = [] } = useListDQATablesQuery();
  const [tableFilter, setTableFilter] = useState("");
  const { data: rules = [], isLoading } = useListDQARulesQuery({ tableId: tableFilter || undefined });
  const [upsertRule, { isLoading: isSaving }] = useUpsertDQARuleMutation();
  const [deleteRule] = useDeleteDQARuleMutation();

  const tableOptions = tables.map((t) => t.table_id);

  const handleSave = async (input: DQARuleInput) => {
    try {
      await upsertRule(input).unwrap();
      toast.success("Rule saved", `${input.code} is now active on ${input.table_id}.`);
      setEditor(null);
    } catch {
      toast.error("Rule not saved", "Check the definition JSON and try again.");
    }
  };

  const rows = rules.map((r) => ({
    id: `${r.table_id}:${r.code}`,
    enabled: r.enabled,
    code: r.code,
    table_id: r.table_id,
    type: r.type,
    category: r.category,
    name: r.name || "—",
    raw: r,
  }));

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: "1rem" }}>
        <div style={{ minWidth: "16rem" }}>
          <Select id="dqa-rules-table-filter" labelText="Filter by table" value={tableFilter} onChange={(e) => setTableFilter(e.target.value)}>
            <SelectItem value="" text="All tables" />
            {tableOptions.map((t) => (
              <SelectItem key={t} value={t} text={t} />
            ))}
          </Select>
        </div>
        <Button renderIcon={Add} onClick={() => setEditor({ mode: "create" })}>
          Add rule
        </Button>
      </div>

      {isLoading ? (
        <p>Loading…</p>
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <Table {...getTableProps()} size="sm">
                <TableHead>
                  <TableRow>
                    {headers.map((header) => {
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
                  {rows.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={headers.length}>No rules yet.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const ruleRaw = rules.find((r) => `${r.table_id}:${r.code}` === row.id);
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === "enabled") {
                              return (
                                <TableCell key={cell.id}>
                                  <TableStatusTag status={cell.value ? "on" : "off"} kind={cell.value ? "green" : "gray"} />
                                </TableCell>
                              );
                            }
                            if (cell.info.header === "actions") {
                              return (
                                <TableCell key={cell.id}>
                                  <div style={{ display: "flex", gap: "0.5rem" }}>
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      onClick={() => ruleRaw && setEditor({ mode: "edit", rule: ruleRaw })}
                                    >
                                      Edit
                                    </Button>
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      hasIconOnly
                                      iconDescription="Delete"
                                      renderIcon={TrashCan}
                                      onClick={() => ruleRaw && deleteRule({ tableId: ruleRaw.table_id, code: ruleRaw.code })}
                                    />
                                  </div>
                                </TableCell>
                              );
                            }
                            return <TableCell key={cell.id}>{String(cell.value)}</TableCell>;
                          })}
                        </TableRow>
                      );
                    })
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DataTable>
      )}

      {editor ? (
        <ComposedModal open size="lg" onClose={() => setEditor(null)}>
          <ModalHeader
            title={editor.mode === "edit" ? `Edit rule: ${editor.rule.code}` : "Add DQA rule"}
          />
          <ModalBody hasScrollingContent>
            <DQARuleForm
              // Remount per rule so the form never shows the previously opened one.
              key={editor.mode === "edit" ? `${editor.rule.table_id}:${editor.rule.code}` : "new"}
              tableOptions={tableOptions}
              initialRule={editor.mode === "edit" ? editor.rule : undefined}
              isSubmitting={isSaving}
              onSubmit={handleSave}
              onClose={() => setEditor(null)}
            />
          </ModalBody>
        </ComposedModal>
      ) : null}
    </div>
  );
}
