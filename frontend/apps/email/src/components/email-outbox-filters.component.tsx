import { Button, Dropdown, Search } from "@carbon/react";
import "../pages/email-outbox.scss";

type Option = {
  id: string;
  label: string;
};

type EmailOutboxFiltersProps<TStatus extends string> = {
  search: string;
  status: TStatus;
  statusOptions: readonly Option[];

  onSearchChange: (value: string) => void;
  onStatusChange: (value: TStatus) => void;
  onReset: () => void;
};

export function EmailOutboxFilters<TStatus extends string>({
  search,
  status,
  statusOptions,
  onSearchChange,
  onStatusChange,
  onReset,
}: EmailOutboxFiltersProps<TStatus>) {
  const statusItems: Option[] = [...statusOptions];

  return (
    <div className="email-outbox-filter-grid">
      <Search
        id="email-outbox-search"
        labelText="Search emails"
        placeholder="Search by subject, recipient, status, or ID..."
        value={search}
        onChange={(event) => onSearchChange(event.target.value)}
      />

      <Dropdown
        id="email-status-filter"
        titleText="Status"
        label="Status"
        items={statusItems}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={statusItems.find((item) => item.id === status)}
        onChange={({ selectedItem }) => {
          if (selectedItem) {
            onStatusChange(selectedItem.id as TStatus);
          }
        }}
      />

      <Button kind="ghost" onClick={onReset}>
        Reset
      </Button>
    </div>
  );
}
