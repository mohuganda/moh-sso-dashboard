import { Select, SelectItem } from "@carbon/react";
import "./surveillance.css";

type FilterOption = {
  value: string;
  label: string;
};

interface SurveillanceFiltersProps {
  epiWeek?: string;
  region?: string;
  district?: string;
  subCounty?: string;
  epiWeekOptions?: FilterOption[];
  regionOptions?: FilterOption[];
  districtOptions?: FilterOption[];
  subCountyOptions?: FilterOption[];
  onEpiWeekChange?: (value: string) => void;
  onRegionChange?: (value: string) => void;
  onDistrictChange?: (value: string) => void;
  onSubCountyChange?: (value: string) => void;
}

const defaultOptions: FilterOption[] = [{ value: "", label: "Select..." }];

export default function SurveillanceFilters({
  epiWeek = "",
  region = "",
  district = "",
  subCounty = "",
  epiWeekOptions = defaultOptions,
  regionOptions = defaultOptions,
  districtOptions = defaultOptions,
  subCountyOptions = defaultOptions,
  onEpiWeekChange,
  onRegionChange,
  onDistrictChange,
  onSubCountyChange,
}: SurveillanceFiltersProps) {
  return (
    <div className="surveillance-filters">
      <div className="surveillance-filters__grid">
        <div className="surveillance-filters__item">
          <Select
            id="epi-week"
            labelText="EPI Week"
            value={epiWeek}
            onChange={(event) => onEpiWeekChange?.(event.target.value)}
          >
            {epiWeekOptions.map((option) => (
              <SelectItem
                key={option.value || "epi-week-placeholder"}
                value={option.value}
                text={option.label}
              />
            ))}
          </Select>
        </div>

        <div className="surveillance-filters__item">
          <Select
            id="region"
            labelText="Region"
            value={region}
            onChange={(event) => onRegionChange?.(event.target.value)}
          >
            {regionOptions.map((option) => (
              <SelectItem
                key={option.value || "region-placeholder"}
                value={option.value}
                text={option.label}
              />
            ))}
          </Select>
        </div>

        <div className="surveillance-filters__item">
          <Select
            id="district"
            labelText="District"
            value={district}
            onChange={(event) => onDistrictChange?.(event.target.value)}
          >
            {districtOptions.map((option) => (
              <SelectItem
                key={option.value || "district-placeholder"}
                value={option.value}
                text={option.label}
              />
            ))}
          </Select>
        </div>

        <div className="surveillance-filters__item">
          <Select
            id="sub-county"
            labelText="Sub County"
            value={subCounty}
            onChange={(event) => onSubCountyChange?.(event.target.value)}
          >
            {subCountyOptions.map((option) => (
              <SelectItem
                key={option.value || "sub-county-placeholder"}
                value={option.value}
                text={option.label}
              />
            ))}
          </Select>
        </div>
      </div>
    </div>
  );
}
