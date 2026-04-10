import React, { useMemo } from "react";
import type {
  FacilityWeeklyMetric,
  WeeklyStatus,
  RiskLevel,
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
  regionId?: string;
  districtId?: string;
  subCountyId?: string;
  weeklyStatuses?: WeeklyStatus[];
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

function countStatuses(rows: WeeklyStatus[]) {
  return rows.reduce(
    (acc, item) => {
      switch (item.status) {
        case "MAROON":
          acc.maroon += 1;
          break;
        case "RED":
          acc.red += 1;
          break;
        case "YELLOW":
          acc.yellow += 1;
          break;
        case "GREEN":
          acc.green += 1;
          break;
        default:
          break;
      }
      return acc;
    },
    { maroon: 0, red: 0, yellow: 0, green: 0 },
  );
}

function DefaultMapFallback({
  weeklyStatuses,
  selectedWeek,
}: {
  weeklyStatuses: WeeklyStatus[];
  selectedWeek?: { year?: number; week?: number };
}) {
  const summary = useMemo(() => countStatuses(weeklyStatuses), [weeklyStatuses]);

  return (
    <div className="surveillance-map">
      <div className="surveillance-map__toolbar">
        <button
          type="button"
          className="surveillance-map__tool-button"
          disabled
          aria-disabled="true"
          title="Map not available"
        >
          Reset
        </button>
      </div>

      <div className="surveillance-map__canvas surveillance-map__canvas--placeholder">
        <div className="surveillance-map__placeholder-shape">Uganda Map</div>
      </div>

      <div className="surveillance-map__legend">
        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--alert" />
          Yellow: Alert {summary.yellow ? `(${summary.yellow})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--watch" />
          Green: Watch {summary.green ? `(${summary.green})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--take-action" />
          Red: Take Action {summary.red ? `(${summary.red})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--immediate-action" />
          Maroon: Immediate Action {summary.maroon ? `(${summary.maroon})` : ""}
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

function getStatusWeight(status: RiskLevel): number {
  switch (status) {
    case "MAROON":
      return 4;
    case "RED":
      return 3;
    case "YELLOW":
      return 2;
    case "GREEN":
      return 1;
    default:
      return 0;
  }
}

export default function SurveillancePanels({
  selectedWeekId,
  selectedWeek,
  regionId = "",
  districtId = "",
  subCountyId = "",
  weeklyStatuses = [],
  facilityMetrics = [],
  loading = false,
  mapContent,
}: SurveillancePanelsProps) {
  const filteredWeeklyStatuses = useMemo(() => {
    return weeklyStatuses.filter((item) => {
      const matchesWeek = !selectedWeekId || item.epi_week_id === selectedWeekId;
      const matchesRegion = !regionId || item.region_id === regionId;
      const matchesDistrict = !districtId || item.district_id === districtId;
      const matchesSubCounty = !subCountyId || item.sub_county_id === subCountyId;

      return matchesWeek && matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [weeklyStatuses, selectedWeekId, regionId, districtId, subCountyId]);

  const filteredFacilityMetrics = useMemo(() => {
    return facilityMetrics.filter((item) => {
      const matchesRegion =
        !regionId || ("region_id" in item && String(item.region_id ?? "") === regionId);

      const matchesDistrict =
        !districtId ||
        ("district_id" in item && String(item.district_id ?? "") === districtId) ||
        normalize(item.district_name) === normalize(districtId);

      const matchesSubCounty =
        !subCountyId ||
        ("sub_county_id" in item && String(item.sub_county_id ?? "") === subCountyId) ||
        normalize(item.subcounty_name) === normalize(subCountyId);

      return matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [facilityMetrics, regionId, districtId, subCountyId]);

  const facilitiesReporting = useMemo<SurveillanceRankItem[]>(() => {
    const grouped = new Map<string, number>();

    for (const item of filteredFacilityMetrics) {
      const label = item.facility_name?.trim() || item.district_name?.trim() || "Unknown";
      const current = grouped.get(label) ?? 0;
      grouped.set(label, current + Number(item.value ?? 0));
    }

    return Array.from(grouped.entries())
      .map(([label, value]) => ({ label, value }))
      .sort((a, b) => Number(b.value) - Number(a.value))
      .slice(0, 10);
  }, [filteredFacilityMetrics]);

  const eidsrAlerts = useMemo<SurveillanceRankItem[]>(() => {
    const grouped = new Map<string, number>();

    for (const item of filteredWeeklyStatuses) {
      const label = item.district_id || item.sub_county_id || item.region_id || "Unknown";

      const current = grouped.get(label) ?? 0;
      grouped.set(label, current + getStatusWeight(item.status));
    }

    return Array.from(grouped.entries())
      .map(([label, value]) => ({ label, value }))
      .filter((item) => Number(item.value) > 0)
      .sort((a, b) => Number(b.value) - Number(a.value))
      .slice(0, 10);
  }, [filteredWeeklyStatuses]);

  return (
    <div className="surveillance-panels">
      <div className="surveillance-panels__grid">
        <RankingPanel title="Facilities Reporting" items={facilitiesReporting} loading={loading} />

        <section className="surveillance-panel surveillance-panel--map">
          {mapContent ? (
            <div className="surveillance-map surveillance-map--embedded">{mapContent}</div>
          ) : (
            <DefaultMapFallback
              weeklyStatuses={filteredWeeklyStatuses}
              selectedWeek={selectedWeek}
            />
          )}
        </section>

        <RankingPanel title="EIDSR Alerts" items={eidsrAlerts} loading={loading} />
      </div>
    </div>
  );
}
