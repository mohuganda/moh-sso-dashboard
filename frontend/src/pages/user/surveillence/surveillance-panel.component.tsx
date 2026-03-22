import React, { useMemo } from "react";
import type {
  DistrictWeeklyStatus,
  RegionWeeklyStatus,
  NationalWeeklyStatus,
  FacilityWeeklyMetric,
} from "../../../store/types/surveillance.types";

type SurveillanceRankItem = {
  label: string;
  value: number | string;
};

interface SurveillancePanelsProps {
  selectedWeekId?: string;
  selectedWeek?: {
    id?: string;
    year?: number;
    week?: number;
  };
  region?: string;
  district?: string;
  subCounty?: string;
  districtStatuses?: DistrictWeeklyStatus[];
  regionStatuses?: RegionWeeklyStatus[];
  nationalStatuses?: NationalWeeklyStatus[];
  facilityMetrics?: FacilityWeeklyMetric[];
  loading?: boolean;
  mapContent?: React.ReactNode;
}

interface RankingPanelProps {
  title: string;
  items: SurveillanceRankItem[];
  loading?: boolean;
}

function RankingPanel({ title, items, loading }: RankingPanelProps) {
  return (
    <section className="surveillance-panel surveillance-panel--side">
      <h3 className="surveillance-panel__title">{title}</h3>

      {loading ? (
        <p className="surveillance-panel__empty">Loading...</p>
      ) : items.length > 0 ? (
        <ol className="surveillance-panel__ranking-list">
          {items.map((item, index) => (
            <li
              key={`${title}-${item.label}-${index}`}
              className="surveillance-panel__ranking-item"
            >
              <span className="surveillance-panel__ranking-label">
                {item.label}: {item.value}
              </span>
            </li>
          ))}
        </ol>
      ) : (
        <p className="surveillance-panel__empty">No data available.</p>
      )}
    </section>
  );
}

function DefaultMapPlaceholder({
  nationalStatuses,
  selectedWeek,
}: {
  nationalStatuses: NationalWeeklyStatus[];
  selectedWeek?: { year?: number; week?: number };
}) {
  const summary = nationalStatuses[0];

  return (
    <div className="surveillance-map">
      <div className="surveillance-map__toolbar">
        <button type="button" className="surveillance-map__tool-button">
          Reset
        </button>
      </div>

      <div className="surveillance-map__canvas">
        <div className="surveillance-map__placeholder-shape">Uganda Map</div>
      </div>

      <div className="surveillance-map__legend">
        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--alert" />
          Yellow: Alert {summary?.yellow ? `(${summary.yellow})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--watch" />
          Green: Watch {summary?.green ? `(${summary.green})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--take-action" />
          Red: Take Action {summary?.red ? `(${summary.red})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--immediate-action" />
          Maroon: Immediate Action {summary?.maroon ? `(${summary.maroon})` : ""}
        </span>
      </div>

      <div className="surveillance-map__meta">
        <p>
          {selectedWeek?.year && selectedWeek?.week
            ? `Week ${selectedWeek.week}, ${selectedWeek.year}`
            : "Current reporting week"}
        </p>
      </div>
    </div>
  );
}

function normalize(value?: string) {
  return (value ?? "").trim().toLowerCase();
}

export default function SurveillancePanels({
  selectedWeekId,
  selectedWeek,
  region = "",
  district = "",
  subCounty = "",
  districtStatuses = [],
  regionStatuses = [],
  nationalStatuses = [],
  facilityMetrics = [],
  loading = false,
  mapContent,
}: SurveillancePanelsProps) {
  const filteredDistrictStatuses = useMemo(() => {
    return districtStatuses.filter((item) => {
      const matchesDistrict = !district || normalize(item.district_name) === normalize(district);

      if (!matchesDistrict) {
        return false;
      }

      if (!region) {
        return true;
      }

      const relatedRegion = regionStatuses.find(
        (regionItem) =>
          normalize(regionItem.region_name) === normalize(region) &&
          (!selectedWeekId || regionItem.epi_week_id === selectedWeekId),
      );

      return Boolean(relatedRegion);
    });
  }, [districtStatuses, district, region, regionStatuses, selectedWeekId]);

  const filteredFacilityMetrics = useMemo(() => {
    return facilityMetrics.filter((item) => {
      const matchesRegion = !region || normalize(item.district_name).includes(normalize(region));
      const matchesDistrict = !district || normalize(item.district_name) === normalize(district);
      const matchesSubCounty =
        !subCounty || normalize(item.subcounty_name) === normalize(subCounty);

      return matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [facilityMetrics, region, district, subCounty]);

  const facilitiesReporting = useMemo<SurveillanceRankItem[]>(() => {
    const grouped = new Map<string, number>();

    for (const item of filteredFacilityMetrics) {
      const label = item.district_name?.trim() || item.facility_name?.trim() || "Unknown";
      const current = grouped.get(label) ?? 0;
      grouped.set(label, current + Number(item.value ?? 0));
    }

    return Array.from(grouped.entries())
      .map(([label, value]) => ({ label, value }))
      .sort((a, b) => Number(b.value) - Number(a.value))
      .slice(0, 10);
  }, [filteredFacilityMetrics]);

  const eidsrAlerts = useMemo<SurveillanceRankItem[]>(() => {
    return filteredDistrictStatuses
      .map((item) => {
        const maroonCount = item.maroon?.length ?? 0;
        const redCount = item.red?.length ?? 0;
        const yellowCount = item.yellow?.length ?? 0;
        const totalAlerts = maroonCount + redCount + yellowCount;

        return {
          label: item.district_name?.trim() || "Unknown",
          value: totalAlerts,
        };
      })
      .filter((item) => Number(item.value) > 0)
      .sort((a, b) => Number(b.value) - Number(a.value))
      .slice(0, 10);
  }, [filteredDistrictStatuses]);

  return (
    <div className="surveillance-panels">
      <div className="surveillance-panels__grid">
        <RankingPanel title="Facilities Reporting" items={facilitiesReporting} loading={loading} />

        <section className="surveillance-panel surveillance-panel--map">
          {mapContent ?? (
            <DefaultMapPlaceholder
              nationalStatuses={nationalStatuses}
              selectedWeek={selectedWeek}
            />
          )}
        </section>

        <RankingPanel title="EIDSR Alerts" items={eidsrAlerts} loading={loading} />
      </div>
    </div>
  );
}
