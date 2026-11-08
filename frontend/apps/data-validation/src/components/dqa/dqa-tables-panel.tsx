import {
  Button,
  ComposedModal,
  DataTable,
  ModalBody,
  ModalHeader,
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

import { useToast } from "@moh-sso/ui";

import { useDeleteDQATableMutation, useListDQATablesQuery, useUpsertDQATableMutation } from "../../api";
import type { DQATableMapping } from "../../dqa.types";
import { DQATableForm } from "./dqa-table-form";

const headers = [
  { key: "table_id", header: "Table ID" },
  { key: "physical_table", header: "Physical table" },
  { key: "period_column", header: "Period column" },
  { key: "dim_columns", header: "Dimensions" },
  { key: "filter_columns", header: "Filters" },
  { key: "actions", header: "" },
];

type TableEditorState = { mode: "create" } | { mode: "edit"; table: DQATableMapping };

export function DQATablesPanel() {
  const { data: tables = [], isLoading } = useListDQATablesQuery();
  const [upsertTable, { isLoading: isSaving }] = useUpsertDQATableMutation();
  const [deleteTable] = useDeleteDQATableMutation();
  const toast = useToast();
  const [editor, setEditor] = useState<TableEditorState | null>(null);

  const handleSave = async (table: DQATableMapping) => {
    try {
      await upsertTable(table).unwrap();
      toast.success("Table saved", `${table.table_id} now maps to ${table.physical_table}.`);
      setEditor(null);
    } catch {
      toast.error("Could not save table mapping", "Please try again.");
    }
  };

  const summarise = (columns: string[]) =>
    columns.length === 0 ? "All" : columns.length <= 2 ? columns.join(", ") : `${columns.length} columns`;

  const rows = tables.map((t) => ({
    id: t.table_id,
    table_id: t.table_id,
    physical_table: t.physical_table,
    period_column: t.period_column || "—",
    dim_columns: summarise(t.dim_columns ?? []),
    filter_columns: summarise(t.filter_columns ?? []),
  }));

  return (
    <div>
      <p style={{ marginBottom: "1rem", color: "var(--cds-text-secondary)" }}>
        Register a logical table_id used by rules to the real DWH table it scans, e.g.{" "}
        <code>vht_monthly</code> → <code>report.echis_vht_monthly</code>.
      </p>

      <div style={{ display: "flex", justifyContent: "flex-end", marginBottom: "1.5rem" }}>
        <Button renderIcon={Add} onClick={() => setEditor({ mode: "create" })}>
          Register table
        </Button>
      </div>

      {isLoading ? (
        <p>Loading…</p>
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
            <TableContainer>
              <Table {...getTableProps()}>
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
                      <TableCell colSpan={headers.length}>No tables registered yet.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) =>
                            cell.info.header === "actions" ? (
                              <TableCell key={cell.id}>
                                <div style={{ display: "flex", gap: "0.5rem" }}>
                                  <Button
                                    kind="ghost"
                                    size="sm"
                                    onClick={() => {
                                      const table = tables.find((t) => t.table_id === row.id);
                                      if (table) setEditor({ mode: "edit", table });
                                    }}
                                  >
                                    Configure
                                  </Button>
                                  <Button
                                    kind="ghost"
                                    size="sm"
                                    hasIconOnly
                                    iconDescription="Remove"
                                    renderIcon={TrashCan}
                                    onClick={() => deleteTable(row.id)}
                                  />
                                </div>
                              </TableCell>
                            ) : (
                              <TableCell key={cell.id}>{cell.value}</TableCell>
                            ),
                          )}
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
            title={editor.mode === "edit" ? `Configure ${editor.table.table_id}` : "Register a table"}
          />
          <ModalBody hasScrollingContent>
            <DQATableForm
              key={editor.mode === "edit" ? editor.table.table_id : "new"}
              initialTable={editor.mode === "edit" ? editor.table : undefined}
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
