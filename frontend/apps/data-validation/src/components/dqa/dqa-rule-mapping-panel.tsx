import {
  Button,
  Checkbox,
  ComposedModal,
  InlineNotification,
  ModalBody,
  ModalFooter,
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
import { useState } from "react";

import { TableStatusTag, useToast } from "@moh-sso/ui";

import {
  useApplyRuleMappingMutation,
  useListDQATablesQuery,
  usePreviewRuleMappingMutation,
} from "../../api";
import type { DQARuleMappingCandidate } from "../../dqa.types";

export function DQARuleMappingPanel() {
  const toast = useToast();
  const { data: tables = [] } = useListDQATablesQuery();
  const [tableId, setTableId] = useState("");
  const [candidates, setCandidates] = useState<DQARuleMappingCandidate[] | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const [previewMapping, { isLoading: isChecking }] = usePreviewRuleMappingMutation();
  const [applyMapping, { isLoading: isApplying }] = useApplyRuleMappingMutation();

  const handleCheck = async () => {
    if (!tableId) {
      toast.error("Pick a table", "Choose the table you want to map rules onto.");
      return;
    }
    try {
      const result = await previewMapping({ table_id: tableId }).unwrap();
      setCandidates(result);
      setSelected(new Set(result.filter((c) => c.mappable && !c.already_on_table).map((c) => c.code)));
    } catch {
      toast.error("Could not check rules", "Please try again.");
    }
  };

  const toggle = (code: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(code)) next.delete(code);
      else next.add(code);
      return next;
    });
  };

  const handleApply = async () => {
    const codes = [...selected];
    if (codes.length === 0) {
      toast.error("Nothing selected", "Select at least one rule to apply.");
      return;
    }
    try {
      const result = await applyMapping({ table_id: tableId, codes }).unwrap();
      if (result.failed.length > 0) {
        toast.error(
          "Some rules were not applied",
          `${result.applied.length} applied, ${result.failed.length} failed.`,
        );
      } else {
        toast.success("Rules applied", `${result.applied.length} rule(s) now run on ${tableId}.`);
      }
      setCandidates(null);
      setSelected(new Set());
    } catch {
      toast.error("Could not apply rules", "Please try again.");
    }
  };

  const mappable = candidates?.filter((c) => c.mappable) ?? [];
  const blocked = candidates?.filter((c) => !c.mappable) ?? [];

  return (
    <div>
      <p style={{ marginBottom: "1rem", color: "var(--cds-text-secondary)" }}>
        Check the rules that already exist on other tables against a table&apos;s column structure, then
        apply the ones that fit.
      </p>

      <div style={{ display: "flex", gap: "1rem", alignItems: "flex-end", marginBottom: "1rem" }}>
        <div style={{ minWidth: "20rem" }}>
          <Select
            id="dqa-mapping-table"
            labelText="Map rules onto"
            value={tableId}
            onChange={(e) => setTableId(e.target.value)}
          >
            <SelectItem value="" text="Select a table…" />
            {tables.map((t) => (
              <SelectItem key={t.table_id} value={t.table_id} text={`${t.table_id} → ${t.physical_table}`} />
            ))}
          </Select>
        </div>
        <Button onClick={handleCheck} disabled={isChecking}>
          {isChecking ? "Checking…" : "Check rules"}
        </Button>
      </div>

      {candidates ? (
        <ComposedModal open size="lg" onClose={() => setCandidates(null)}>
          <ModalHeader title={`Rules that can map onto ${tableId}`} />
          <ModalBody hasScrollingContent>
            <InlineNotification
              kind={mappable.length > 0 ? "success" : "warning"}
              title={`${mappable.length} of ${candidates.length} rules fit this table`}
              subtitle={
                blocked.length > 0
                  ? `${blocked.length} cannot be mapped because the table is missing columns they use.`
                  : "Every existing rule compiles against this table."
              }
              lowContrast
              hideCloseButton
            />

            <TableContainer title="Can be mapped" style={{ marginTop: "1rem" }}>
              <Table size="sm">
                <TableHead>
                  <TableRow>
                    <TableHeader />
                    <TableHeader>Code</TableHeader>
                    <TableHeader>From</TableHeader>
                    <TableHeader>Type</TableHeader>
                    <TableHeader>Name</TableHeader>
                    <TableHeader>Status</TableHeader>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {mappable.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6}>No existing rule fits this table.</TableCell>
                    </TableRow>
                  ) : (
                    mappable.map((c) => (
                      <TableRow key={c.code}>
                        <TableCell>
                          <Checkbox
                            id={`map-${c.code}`}
                            labelText=""
                            checked={selected.has(c.code)}
                            onChange={() => toggle(c.code)}
                          />
                        </TableCell>
                        <TableCell>{c.code}</TableCell>
                        <TableCell>{c.source_table_id}</TableCell>
                        <TableCell>{c.type}</TableCell>
                        <TableCell>{c.name || "—"}</TableCell>
                        <TableCell>
                          {c.already_on_table ? (
                            <TableStatusTag status="already mapped" kind="blue" />
                          ) : (
                            <TableStatusTag status="new" kind="green" />
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>

            {blocked.length > 0 ? (
              <TableContainer title="Cannot be mapped" style={{ marginTop: "1.5rem" }}>
                <Table size="sm">
                  <TableHead>
                    <TableRow>
                      <TableHeader>Code</TableHeader>
                      <TableHeader>From</TableHeader>
                      <TableHeader>Why not</TableHeader>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {blocked.map((c) => (
                      <TableRow key={c.code}>
                        <TableCell>{c.code}</TableCell>
                        <TableCell>{c.source_table_id}</TableCell>
                        <TableCell>{c.reason}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            ) : null}
          </ModalBody>
          <ModalFooter>
            <Button kind="secondary" onClick={() => setCandidates(null)}>
              Cancel
            </Button>
            <Button kind="primary" onClick={handleApply} disabled={isApplying || selected.size === 0}>
              {isApplying ? "Applying…" : `Apply ${selected.size} rule(s)`}
            </Button>
          </ModalFooter>
        </ComposedModal>
      ) : null}
    </div>
  );
}
