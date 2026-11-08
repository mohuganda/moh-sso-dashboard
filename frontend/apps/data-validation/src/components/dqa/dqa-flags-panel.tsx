import {
  DataTable,
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
import { useState } from "react";

import { TableStatusTag } from "@moh-sso/ui";

import { useListDQAFlagsQuery } from "../../api";

const headers = [
  { key: "severity", header: "Severity" },
  { key: "rule_code", header: "Rule" },
  { key: "category", header: "Category" },
  { key: "entity_name", header: "Entity" },
  { key: "district", header: "District" },
  { key: "period_date", header: "Period" },
  { key: "detail", header: "Detail" },
];

export function DQAFlagsPanel({ runId }: { runId: number }) {
  const [severity, setSeverity] = useState<string>("");
  const { data: flags = [], isLoading } = useListDQAFlagsQuery({ runId, severity: severity || undefined });

  const rows = flags.map((f) => ({
    id: String(f.id),
    severity: f.severity,
    rule_code: f.rule_code,
    category: f.category,
    entity_name: f.entity_name || f.entity_id || "—",
    district: f.district || "—",
    period_date: f.period_date || "—",
    detail: f.detail,
  }));

  return (
    <div>
      <div style={{ marginBottom: "1rem", maxWidth: "16rem" }}>
        <Select
          id="dqa-flags-severity"
          labelText="Filter by severity"
          value={severity}
          onChange={(e) => setSeverity(e.target.value)}
        >
          <SelectItem value="" text="All" />
          <SelectItem value="fail" text="Fail" />
          <SelectItem value="warn" text="Warn" />
        </Select>
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
                      <TableCell colSpan={headers.length}>No flags for this run.</TableCell>
                    </TableRow>
                  ) : (
                    rows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) =>
                            cell.info.header === "severity" ? (
                              <TableCell key={cell.id}>
                                <TableStatusTag
                                  status={String(cell.value)}
                                  kind={cell.value === "fail" ? "red" : "magenta"}
                                />
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
    </div>
  );
}
