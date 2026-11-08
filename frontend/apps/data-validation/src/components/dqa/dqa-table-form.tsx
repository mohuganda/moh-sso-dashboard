import {
  Button,
  InlineNotification,
  MultiSelect,
  Select,
  SelectItem,
  Stack,
  TextInput,
} from "@carbon/react";
import { useEffect, useState } from "react";

import { useLazyListDQAPhysicalColumnsQuery } from "../../api";
import type { DQATableMapping } from "../../dqa.types";

type Props = {
  initialTable?: DQATableMapping;
  isSubmitting: boolean;
  onSubmit: (table: DQATableMapping) => Promise<void> | void;
  onClose: () => void;
};

export function DQATableForm({ initialTable, isSubmitting, onSubmit, onClose }: Props) {
  const [tableId, setTableId] = useState(initialTable?.table_id ?? "");
  const [physicalTable, setPhysicalTable] = useState(initialTable?.physical_table ?? "");
  const [description, setDescription] = useState(initialTable?.description ?? "");
  const [periodColumn, setPeriodColumn] = useState(initialTable?.period_column ?? "");
  const [dimColumns, setDimColumns] = useState<string[]>(initialTable?.dim_columns ?? []);
  const [filterColumns, setFilterColumns] = useState<string[]>(initialTable?.filter_columns ?? []);
  const [error, setError] = useState<string | null>(null);

  const [loadColumns, { data: columns = [], isFetching, isError }] = useLazyListDQAPhysicalColumnsQuery();

  // Load the columns of an already-configured table so the pickers open populated.
  useEffect(() => {
    if (initialTable?.physical_table) {
      loadColumns(initialTable.physical_table);
    }
  }, [initialTable?.physical_table, loadColumns]);

  const columnNames = columns.map((c) => c.name);
  const hasColumns = columnNames.length > 0;

  const handleLoadColumns = () => {
    const table = physicalTable.trim();
    if (!table) {
      setError("Enter the physical table first.");
      return;
    }
    setError(null);
    loadColumns(table);
  };

  const handleSave = async () => {
    if (!tableId.trim() || !physicalTable.trim()) {
      setError("Table ID and physical table are required.");
      return;
    }
    setError(null);
    await onSubmit({
      table_id: tableId.trim(),
      physical_table: physicalTable.trim(),
      description: description.trim(),
      is_active: initialTable?.is_active ?? true,
      period_column: periodColumn || undefined,
      dim_columns: dimColumns,
      filter_columns: filterColumns,
    });
  };

  return (
    <Stack gap={5}>
      <TextInput
        id="dqa-table-id"
        labelText="Table ID"
        helperText="The logical name rules refer to, e.g. vht_monthly"
        value={tableId}
        disabled={!!initialTable}
        onChange={(e) => setTableId(e.target.value)}
      />

      <Stack orientation="horizontal" gap={4} style={{ alignItems: "flex-end" }}>
        <div style={{ flex: 1 }}>
          <TextInput
            id="dqa-table-physical"
            labelText="Physical table (schema.table)"
            placeholder="e.g. report.cht_form_097b_with_vhts"
            value={physicalTable}
            onChange={(e) => setPhysicalTable(e.target.value)}
          />
        </div>
        <Button kind="tertiary" onClick={handleLoadColumns} disabled={isFetching}>
          {isFetching ? "Loading…" : "Load columns"}
        </Button>
      </Stack>

      <TextInput
        id="dqa-table-description"
        labelText="Description"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />

      {isError ? (
        <InlineNotification
          kind="error"
          title="Table not found"
          subtitle="Check the schema and table name, then load the columns again."
          lowContrast
          hideCloseButton
        />
      ) : null}

      {hasColumns ? (
        <>
          <Select
            id="dqa-table-period-column"
            labelText="Period column"
            helperText="Used to scope a run to a reporting period"
            value={periodColumn}
            onChange={(e) => setPeriodColumn(e.target.value)}
          >
            <SelectItem value="" text="— none —" />
            {columns.map((c) => (
              <SelectItem key={c.name} value={c.name} text={`${c.name} (${c.type})`} />
            ))}
          </Select>

          <MultiSelect
            id="dqa-table-dim-columns"
            titleText="Dimension columns to keep"
            helperText="Stored on every flag. Leave empty to keep every identifying column."
            label={dimColumns.length ? `${dimColumns.length} selected` : "Select columns"}
            items={columnNames}
            selectedItems={dimColumns}
            onChange={({ selectedItems }) => setDimColumns(selectedItems ?? [])}
          />

          <MultiSelect
            id="dqa-table-filter-columns"
            titleText="Filter columns"
            helperText="Offered as filters when a run is launched, e.g. VHT or facility."
            label={filterColumns.length ? `${filterColumns.length} selected` : "Select columns"}
            items={columnNames}
            selectedItems={filterColumns}
            onChange={({ selectedItems }) => setFilterColumns(selectedItems ?? [])}
          />
        </>
      ) : (
        <p style={{ fontSize: "0.75rem", color: "var(--cds-text-secondary)" }}>
          Load the columns to choose the period, dimension and filter columns.
        </p>
      )}

      {error ? (
        <InlineNotification kind="error" title="Check the form" subtitle={error} lowContrast hideCloseButton />
      ) : null}

      <Stack orientation="horizontal" gap={4}>
        <Button kind="primary" onClick={handleSave} disabled={isSubmitting}>
          {isSubmitting ? "Saving…" : initialTable ? "Save changes" : "Register table"}
        </Button>
        <Button kind="ghost" onClick={onClose}>
          Cancel
        </Button>
      </Stack>
    </Stack>
  );
}
