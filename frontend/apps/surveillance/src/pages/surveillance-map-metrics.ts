import type { FacilityWeeklyMetric } from "../types";
import { districtKey } from "./surveillance-geography";

export type AreaMetrics = { cases: number; facilities: Set<string> };

export const subcountyMetricKey = (district: string, subcounty: string) =>
  `${districtKey(district)}:${subcounty.trim().toLowerCase()}`;

export function buildMapMetrics(metrics: FacilityWeeklyMetric[], weekId?: string) {
  const districts = new Map<string, AreaMetrics>();
  const subcounties = new Map<string, AreaMetrics>();
  const add = (
    areas: Map<string, AreaMetrics>,
    key: string,
    metric: FacilityWeeklyMetric,
    cases: number,
  ) => {
    const summary = areas.get(key) ?? { cases: 0, facilities: new Set<string>() };
    summary.cases += cases;
    summary.facilities.add(metric.facility_id);
    areas.set(key, summary);
  };
  for (const metric of metrics) {
    if (
      !weekId ||
      metric.epi_week_id !== weekId ||
      !metric.disease_id ||
      metric.indicator_id ||
      !metric.facility_id
    )
      continue;
    const rawValue = String(metric.metric_value ?? "").trim();
    const cases = Number(rawValue);
    if (!rawValue || !Number.isFinite(cases) || cases < 0) continue;
    const district = districtKey(metric.district_name);
    if (!district) continue;
    add(districts, district, metric, cases);
    const subcounty = metric.sub_county_name ?? metric.subcounty_name;
    if (subcounty?.trim()) add(subcounties, subcountyMetricKey(district, subcounty), metric, cases);
  }
  return { districts, subcounties };
}
