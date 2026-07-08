import { Button, Dropdown, Search } from "@carbon/react";
import "./announcements.components.scss";

type Option = {
  id: string;
  label: string;
};

type AnnouncementFiltersProps<
  TStatus extends string,
  TLevel extends string,
  TAudience extends string,
  TPin extends string,
> = {
  search: string;

  status: TStatus;
  level: TLevel;
  audience: TAudience;
  pin: TPin;

  statusOptions: readonly Option[];
  levelOptions: readonly Option[];
  audienceOptions: readonly Option[];
  pinOptions: readonly Option[];

  onSearchChange: (value: string) => void;
  onStatusChange: (value: TStatus) => void;
  onLevelChange: (value: TLevel) => void;
  onAudienceChange: (value: TAudience) => void;
  onPinChange: (value: TPin) => void;
  onReset: () => void;
};

export function AnnouncementFilters<
  TStatus extends string,
  TLevel extends string,
  TAudience extends string,
  TPin extends string,
>({
  search,
  status,
  level,
  audience,
  pin,
  statusOptions,
  levelOptions,
  audienceOptions,
  pinOptions,
  onSearchChange,
  onStatusChange,
  onLevelChange,
  onAudienceChange,
  onPinChange,
  onReset,
}: AnnouncementFiltersProps<TStatus, TLevel, TAudience, TPin>) {
  const statusItems = [...statusOptions];
  const levelItems = [...levelOptions];
  const audienceItems = [...audienceOptions];
  const pinItems = [...pinOptions];

  return (
    <div className="announcement-filter-grid">
      <Search
        id="announcement-search"
        labelText="Search announcements"
        placeholder="Search by title, message, tag, level, status..."
        value={search}
        onChange={(event) => onSearchChange(event.target.value)}
      />

      <Dropdown
        id="announcement-status-filter"
        titleText="Status"
        label="Status"
        items={statusItems}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={statusItems.find((item) => item.id === status)}
        onChange={({ selectedItem }) => {
          if (selectedItem) onStatusChange(selectedItem.id as TStatus);
        }}
      />

      <Dropdown
        id="announcement-level-filter"
        titleText="Level"
        label="Level"
        items={levelItems}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={levelItems.find((item) => item.id === level)}
        onChange={({ selectedItem }) => {
          if (selectedItem) onLevelChange(selectedItem.id as TLevel);
        }}
      />

      <Dropdown
        id="announcement-audience-filter"
        titleText="Audience"
        label="Audience"
        items={audienceItems}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={audienceItems.find((item) => item.id === audience)}
        onChange={({ selectedItem }) => {
          if (selectedItem) onAudienceChange(selectedItem.id as TAudience);
        }}
      />

      <Dropdown
        id="announcement-pin-filter"
        titleText="Pin"
        label="Pin"
        items={pinItems}
        itemToString={(item) => item?.label ?? ""}
        selectedItem={pinItems.find((item) => item.id === pin)}
        onChange={({ selectedItem }) => {
          if (selectedItem) onPinChange(selectedItem.id as TPin);
        }}
      />

      <Button kind="ghost" onClick={onReset}>
        Reset
      </Button>
    </div>
  );
}
