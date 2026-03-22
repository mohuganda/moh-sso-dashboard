import { Breadcrumb, BreadcrumbItem, InlineLoading } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import SurveillanceFilters from "./surveillance-filters.component";
import SurveillanceTiles from "./surveillance-tiles.component";
import SurveillancePanels from "./surveillance-panel.component";
import "./surveillance.css";
import {
  useListEpiWeeksQuery,
  useListDiseasesQuery,
  useListDistrictWeeklyStatusesByWeekQuery,
  useListRegionWeeklyStatusesByWeekQuery,
  useListNationalWeeklyStatusesByWeekQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
} from "../../../store/api/surveillance.api";

type FilterOption = {
  value: string;
  label: string;
};

function slugifyDiseaseName(name: string) {
  return name.trim().toLowerCase().replace(/\s+/g, "-");
}

export default function SurveillanceDashboardPage() {
  const navigate = useNavigate();

  const [selectedWeekId, setSelectedWeekId] = useState("");
  const [region, setRegion] = useState("");
  const [district, setDistrict] = useState("");
  const [subCounty, setSubCounty] = useState("");

  const { data: weeks = [], isLoading: weeksLoading } = useListEpiWeeksQuery();
  const { data: diseases = [], isLoading: diseasesLoading } = useListDiseasesQuery();

  useEffect(() => {
    if (!selectedWeekId && weeks.length > 0) {
      setSelectedWeekId(weeks[0].id);
    }
  }, [weeks, selectedWeekId]);

  const { data: districtStatuses = [], isFetching: districtsLoading } =
    useListDistrictWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: regionStatuses = [], isFetching: regionsLoading } =
    useListRegionWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: nationalStatuses = [], isFetching: nationalLoading } =
    useListNationalWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetrics = [], isFetching: facilitiesLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const loading =
    weeksLoading ||
    diseasesLoading ||
    districtsLoading ||
    regionsLoading ||
    nationalLoading ||
    facilitiesLoading;

  const selectedWeek = useMemo(
    () => weeks.find((week) => week.id === selectedWeekId),
    [weeks, selectedWeekId],
  );

  const currentEpiWeekLabel = useMemo(() => {
    if (!selectedWeek) return "--";
    return String(selectedWeek.week ?? "--");
  }, [selectedWeek]);

  const epiWeekOptions: FilterOption[] = useMemo(() => {
    return [
      { value: "", label: "Select..." },
      ...weeks.map((week) => ({
        value: week.id,
        label:
          week.year && week.week ? `${week.year} - Week ${week.week}` : `Week ${week.week ?? ""}`,
      })),
    ];
  }, [weeks]);

  const regionOptions: FilterOption[] = useMemo(() => {
    const names = Array.from(
      new Set(
        regionStatuses
          .map((item) => item.region_name)
          .filter((value): value is string => Boolean(value)),
      ),
    ).sort();

    return [
      { value: "", label: "Select..." },
      ...names.map((name) => ({
        value: name,
        label: name,
      })),
    ];
  }, [regionStatuses]);

  const districtOptions: FilterOption[] = useMemo(() => {
    const names = Array.from(
      new Set(
        districtStatuses
          .map((item) => item.district_name)
          .filter((value): value is string => Boolean(value)),
      ),
    ).sort();

    return [
      { value: "", label: "Select..." },
      ...names.map((name) => ({
        value: name,
        label: name,
      })),
    ];
  }, [districtStatuses]);

  const subCountyOptions: FilterOption[] = useMemo(() => {
    const names = Array.from(
      new Set(
        facilityMetrics
          .map((item) => item.subcounty_name)
          .filter((value): value is string => Boolean(value)),
      ),
    ).sort();

    return [
      { value: "", label: "Select..." },
      ...names.map((name) => ({
        value: name,
        label: name,
      })),
    ];
  }, [facilityMetrics]);

  const handleOpenDisease = (diseaseName: string) => {
    navigate(`/portal/surveillance/${encodeURIComponent(diseaseName)}`);
  };

  const immediateActionItems = useMemo(() => {
    return diseases.slice(0, 4).map((disease) => ({
      label: disease.name,
      onClick: () => handleOpenDisease(slugifyDiseaseName(disease.name)),
    }));
  }, [diseases]);

  const takeActionItems = useMemo(() => {
    return diseases.slice(4, 8).map((disease) => ({
      label: disease.name,
      onClick: () => handleOpenDisease(slugifyDiseaseName(disease.name)),
    }));
  }, [diseases]);

  const alertItems = useMemo(() => {
    const redDiseases = districtStatuses.flatMap((item) => item.red ?? []);
    return Array.from(new Set(redDiseases))
      .slice(0, 6)
      .map((name) => ({
        label: name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(name)),
      }));
  }, [districtStatuses]);

  const watchItems = useMemo(() => {
    const yellowDiseases = districtStatuses.flatMap((item) => item.yellow ?? []);
    return Array.from(new Set(yellowDiseases))
      .slice(0, 6)
      .map((name) => ({
        label: name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(name)),
      }));
  }, [districtStatuses]);

  useEffect(() => {
    setDistrict("");
    setSubCounty("");
  }, [region]);

  useEffect(() => {
    setSubCounty("");
  }, [district]);

  return (
    <div className="surveillance-dashboard-page">
      <div className="surveillance-dashboard-page__header">
        <Breadcrumb className="surveillance-dashboard-page__breadcrumb" noTrailingSlash>
          <BreadcrumbItem href="/portal/apps/dwh/data-visualizer">
            Data &amp; Statistics
          </BreadcrumbItem>

          <BreadcrumbItem isCurrentPage>
            <span>National Surveillance Reporting Dashboard: Epi-Week: {currentEpiWeekLabel}</span>
          </BreadcrumbItem>
        </Breadcrumb>

        <div className="surveillance-dashboard-page__hero">
          <h1 className="surveillance-dashboard-page__title">
            National Surveillance Reporting Dashboard
          </h1>
          <p className="surveillance-dashboard-page__subtitle">
            Monitor surveillance signals, reporting trends, alerts, and priority conditions across
            regions, districts, and sub-counties.
          </p>

          {loading ? (
            <div style={{ marginTop: "0.75rem" }}>
              <InlineLoading description="Loading surveillance data..." />
            </div>
          ) : null}
        </div>
      </div>

      <section className="surveillance-dashboard-page__section">
        <SurveillanceFilters
          epiWeek={selectedWeekId}
          region={region}
          district={district}
          subCounty={subCounty}
          epiWeekOptions={epiWeekOptions}
          regionOptions={regionOptions}
          districtOptions={districtOptions}
          subCountyOptions={subCountyOptions}
          onEpiWeekChange={setSelectedWeekId}
          onRegionChange={setRegion}
          onDistrictChange={setDistrict}
          onSubCountyChange={setSubCounty}
        />
      </section>

      <section className="surveillance-dashboard-page__section">
        <SurveillanceTiles
          immediateAction={{
            title: "Immediate Action",
            items: immediateActionItems,
          }}
          takeAction={{
            title: "Take Action",
            items: takeActionItems,
          }}
          alert={{
            title: "Alert",
            items: alertItems,
          }}
          watch={{
            title: "Watch",
            items: watchItems,
          }}
        />
      </section>

      <section className="surveillance-dashboard-page__section">
        <SurveillancePanels
          selectedWeekId={selectedWeekId}
          selectedWeek={selectedWeek}
          region={region}
          district={district}
          subCounty={subCounty}
          districtStatuses={districtStatuses}
          regionStatuses={regionStatuses}
          nationalStatuses={nationalStatuses}
          facilityMetrics={facilityMetrics}
          loading={loading}
        />
      </section>
    </div>
  );
}
