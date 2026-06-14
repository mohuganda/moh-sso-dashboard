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

type SuccessOption = {
  id: SuccessFilter;
  label: string;
};

const SUCCESS_OPTIONS: SuccessOption[] = [
  { id: "true", label: "Success" },
  { id: "false", label: "Failure" },
];

interface AuditLogFiltersProps {
  action?: string;
  clientId?: string;
  userId?: string;
  ip?: string;
  success?: SuccessFilter;

  onFromChange: (v: string) => void;
  onToChange: (v: string) => void;
  onActionChange: (v?: string) => void;
  onClientChange: (v?: string) => void;
  onUserChange: (v?: string) => void;
  onIpChange: (v?: string) => void;
  onSuccessChange: (v?: SuccessFilter) => void;
  onClear: () => void;
  onExportLoadedCsv: () => void;
  onExportLoadedJson: () => void;
  onExportFilteredCsv: () => void;
  onExportFilteredJson: () => void;
}

function toRFC3339(d: Date) {
  return d.toISOString();
}

export function AuditLogFilters({
  action,
  clientId,
  userId,
  ip,
  success,
  onFromChange,
  onToChange,
  onActionChange,
  onClientChange,
  onUserChange,
  onIpChange,
  onSuccessChange,
  onClear,
  onExportLoadedCsv,
  onExportLoadedJson,
  onExportFilteredCsv,
  onExportFilteredJson,
}: AuditLogFiltersProps) {
  return (
    <Stack
      orientation="horizontal"
      gap={5}
      style={{
        width: "100%",
        justifyContent: "flex-end",
        alignItems: "flex-end",
        flexWrap: "wrap",
      }}
    >
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

      <Search
        id="audit-action"
        labelText="Action"
        placeholder="Action"
        value={action ?? ""}
        style={{ width: 180 }}
        onChange={(event) => {
          onActionChange(event.target.value || undefined);
        }}
      />

      <Search
        id="audit-client"
        labelText="Client"
        placeholder="Client ID"
        value={clientId ?? ""}
        style={{ width: 180 }}
        onChange={(event) => {
          onClientChange(event.target.value || undefined);
        }}
      />

      <Search
        id="audit-user"
        labelText="User ID"
        placeholder="User ID"
        value={userId ?? ""}
        style={{ width: 180 }}
        onChange={(event) => {
          onUserChange(event.target.value || undefined);
        }}
      />

      <Search
        id="audit-ip"
        labelText="IP address"
        placeholder="IP address"
        value={ip ?? ""}
        style={{ width: 160 }}
        onChange={(event) => {
          onIpChange(event.target.value || undefined);
        }}
      />

      <ComboBox
        id="audit-result"
        titleText="Result"
        placeholder="Result"
        style={{ width: 160 }}
        items={SUCCESS_OPTIONS}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={SUCCESS_OPTIONS.find((item) => item.id === success) ?? null}
        onChange={({ selectedItem }) => {
          onSuccessChange(selectedItem?.id);
        }}
      />

      <Stack orientation="horizontal" gap={3}>
        <Button kind="secondary" size="md" onClick={onClear}>
          Clear
        </Button>

        <OverflowMenu ariaLabel="Export audit logs" flipped>
          <OverflowMenuItem itemText="Export filtered CSV" onClick={onExportFilteredCsv} />
          <OverflowMenuItem itemText="Export filtered JSON" onClick={onExportFilteredJson} />
          <OverflowMenuItem itemText="Export loaded CSV" onClick={onExportLoadedCsv} />
          <OverflowMenuItem itemText="Export loaded JSON" onClick={onExportLoadedJson} />
        </OverflowMenu>
      </Stack>
    </Stack>
  );
}
