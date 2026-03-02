import {
  Button,
  ComboBox,
  DatePicker,
  DatePickerInput,
  Search,
  Stack,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";

type SuccessFilter = "true" | "false";

interface AuditLogFiltersProps {
  onFromChange: (v: string) => void;
  onToChange: (v: string) => void;
  onClientChange: (v?: string) => void;
  onSuccessChange: (v?: SuccessFilter) => void;
  onClear: () => void;
  onExportCsv: () => void;
  onExportJson: () => void;
}

function toRFC3339(d: Date) {
  return d.toISOString();
}

export function AuditLogFilters({
  onFromChange,
  onToChange,
  onClientChange,
  onSuccessChange,
  onClear,
  onExportCsv,
  onExportJson,
}: AuditLogFiltersProps) {
  return (
    <Stack
      orientation="horizontal"
      gap={5}
      style={{
        width: "100%",
        justifyContent: "flex-end",
        alignItems: "flex-end",
        flexWrap: "nowrap",
      }}
    >
      {/* Date range */}
      <DatePicker
        datePickerType="range"
        onChange={(dates) => {
          if (Array.isArray(dates)) {
            if (dates[0] instanceof Date) {
              onFromChange(toRFC3339(dates[0]));
            }
            if (dates[1] instanceof Date) {
              onToChange(toRFC3339(dates[1]));
            }
          }
        }}
      >
        <DatePickerInput id="audit-from" labelText="From date" hideLabel placeholder="From" />
        <DatePickerInput id="audit-to" labelText="To date" hideLabel placeholder="To" />
      </DatePicker>

      {/* Client */}
      <Search
        id="audit-client"
        labelText="Client"
        placeholder="Client ID"
        style={{ width: 180 }}
        onChange={(e) => {
          onClientChange(e.target.value || undefined);
        }}
      />

      {/* Result */}
      <ComboBox
        id="audit-result"
        titleText="Result"
        placeholder="Result"
        style={{ width: 160 }}
        items={[
          { id: "true", label: "Success" },
          { id: "false", label: "Failure" },
        ]}
        itemToString={(item) => item?.label ?? ""}
        onChange={({ selectedItem }) => {
          onSuccessChange(selectedItem?.id as SuccessFilter | undefined);
        }}
      />

      {/* Actions */}
      <Stack orientation="horizontal" gap={3}>
        <Button kind="secondary" size="md" onClick={onClear}>
          Clear
        </Button>

        {/* Export menu */}
        <OverflowMenu ariaLabel="Export audit logs" flipped>
          <OverflowMenuItem itemText="Export CSV" onClick={onExportCsv} />
          <OverflowMenuItem itemText="Export JSON" onClick={onExportJson} />
        </OverflowMenu>
      </Stack>
    </Stack>
  );
}
