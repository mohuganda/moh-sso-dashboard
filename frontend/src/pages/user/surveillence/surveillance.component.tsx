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
  useListRegionsQuery,
  useListDistrictsByRegionQuery,
  useListSubcountiesByDistrictQuery,
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
  const [selectedRegionId, setSelectedRegionId] = useState("");
  const [selectedDistrictId, setSelectedDistrictId] = useState("");
  const [selectedSubCountyId, setSelectedSubCountyId] = useState("");
  const [selectedYear] = useState(2025);

  const { data: weeksResponse, isLoading: weeksLoading } = useListEpiWeeksQuery(selectedYear);
  const { data: diseasesResponse, isLoading: diseasesLoading } = useListDiseasesQuery();

  const { data: regionsResponse, isLoading: regionsReferenceLoading } = useListRegionsQuery();

  const { data: districtsResponse, isLoading: districtsReferenceLoading } =
    useListDistrictsByRegionQuery(selectedRegionId, {
      skip: !selectedRegionId,
    });

  const { data: subcountiesResponse, isLoading: subcountiesReferenceLoading } =
    useListSubcountiesByDistrictQuery(selectedDistrictId, {
      skip: !selectedDistrictId,
    });

  const weeks = Array.isArray(weeksResponse) ? weeksResponse : [];
  const diseases = Array.isArray(diseasesResponse) ? diseasesResponse : [];
  const regions = Array.isArray(regionsResponse) ? regionsResponse : [];
  const districts = Array.isArray(districtsResponse) ? districtsResponse : [];
  const subcounties = Array.isArray(subcountiesResponse) ? subcountiesResponse : [];

  const { data: districtStatusesResponse, isFetching: districtsLoading } =
    useListDistrictWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: regionStatusesResponse, isFetching: regionsLoading } =
    useListRegionWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: nationalStatusesResponse, isFetching: nationalLoading } =
    useListNationalWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetricsResponse, isFetching: facilitiesLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const districtStatuses = Array.isArray(districtStatusesResponse) ? districtStatusesResponse : [];
  const regionStatuses = Array.isArray(regionStatusesResponse) ? regionStatusesResponse : [];
  const nationalStatuses = Array.isArray(nationalStatusesResponse) ? nationalStatusesResponse : [];
  const facilityMetrics = Array.isArray(facilityMetricsResponse) ? facilityMetricsResponse : [];

  useEffect(() => {
    if (!selectedWeekId && weeks.length > 0) {
      setSelectedWeekId(weeks[0].id);
    }
  }, [weeks, selectedWeekId]);

  useEffect(() => {
    setSelectedDistrictId("");
    setSelectedSubCountyId("");
  }, [selectedRegionId]);

  useEffect(() => {
    setSelectedSubCountyId("");
  }, [selectedDistrictId]);

  const loading =
    weeksLoading ||
    diseasesLoading ||
    regionsReferenceLoading ||
    districtsReferenceLoading ||
    subcountiesReferenceLoading ||
    districtsLoading ||
    regionsLoading ||
    nationalLoading ||
    facilitiesLoading;

  const selectedWeek = useMemo(
    () => weeks.find((week) => week.id === selectedWeekId),
    [weeks, selectedWeekId],
  );

  const selectedRegion = useMemo(
    () => regions.find((item) => item.id === selectedRegionId),
    [regions, selectedRegionId],
  );

  const selectedDistrict = useMemo(
    () => districts.find((item) => item.id === selectedDistrictId),
    [districts, selectedDistrictId],
  );

  const selectedSubCounty = useMemo(
    () => subcounties.find((item) => item.id === selectedSubCountyId),
    [subcounties, selectedSubCountyId],
  );

  const currentEpiWeekLabel = useMemo(() => {
    if (!selectedWeek) return "--";
    return String(selectedWeek.epi_week ?? "--");
  }, [selectedWeek]);

  const epiWeekOptions: FilterOption[] = useMemo(() => {
    return [
      { value: "", label: "Select..." },
      ...weeks.map((week) => ({
        value: week.id,
        label:
          week.epi_year && week.epi_week
            ? `${week.epi_year} - Week ${week.epi_week}`
            : `Week ${week.epi_week ?? ""}`,
      })),
    ];
  }, [weeks]);

  const regionOptions: FilterOption[] = useMemo(() => {
    return [
      { value: "", label: "Select..." },
      ...regions.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ];
  }, [regions]);

  const districtOptions: FilterOption[] = useMemo(() => {
    return [
      { value: "", label: "Select..." },
      ...districts.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ];
  }, [districts]);

  const subCountyOptions: FilterOption[] = useMemo(() => {
    return [
      { value: "", label: "Select..." },
      ...subcounties.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ];
  }, [subcounties]);

  const filteredRegionStatuses = useMemo(() => {
    if (!selectedRegion?.name) return regionStatuses;

    return regionStatuses.filter(
      (item) => item.region_name?.trim().toLowerCase() === selectedRegion.name.trim().toLowerCase(),
    );
  }, [regionStatuses, selectedRegion]);

  const filteredDistrictStatuses = useMemo(() => {
    let result = districtStatuses;

    if (selectedRegion?.name) {
      result = result.filter(
        (item) =>
          item.region_name?.trim().toLowerCase() === selectedRegion.name.trim().toLowerCase(),
      );
    }

    if (selectedDistrict?.name) {
      result = result.filter(
        (item) =>
          item.district_name?.trim().toLowerCase() === selectedDistrict.name.trim().toLowerCase(),
      );
    }

    return result;
  }, [districtStatuses, selectedRegion, selectedDistrict]);

  const filteredFacilityMetrics = useMemo(() => {
    let result = facilityMetrics;

    if (selectedRegion?.name) {
      result = result.filter(
        (item) =>
          item.region_name?.trim().toLowerCase() === selectedRegion.name.trim().toLowerCase(),
      );
    }

    if (selectedDistrict?.name) {
      result = result.filter(
        (item) =>
          item.district_name?.trim().toLowerCase() === selectedDistrict.name.trim().toLowerCase(),
      );
    }

    if (selectedSubCounty?.name) {
      result = result.filter(
        (item) =>
          item.subcounty_name?.trim().toLowerCase() === selectedSubCounty.name.trim().toLowerCase(),
      );
    }

    return result;
  }, [facilityMetrics, selectedRegion, selectedDistrict, selectedSubCounty]);

  const handleOpenDisease = (diseaseSlug: string) => {
    navigate(`/portal/surveillance/${encodeURIComponent(diseaseSlug)}`);
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
    const redDiseases = filteredDistrictStatuses.flatMap((item) =>
      Array.isArray(item.red) ? item.red : [],
    );

    return Array.from(new Set(redDiseases.filter((name): name is string => Boolean(name?.trim()))))
      .slice(0, 6)
      .map((name) => ({
        label: name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(name)),
      }));
  }, [filteredDistrictStatuses]);

  const watchItems = useMemo(() => {
    const yellowDiseases = filteredDistrictStatuses.flatMap((item) =>
      Array.isArray(item.yellow) ? item.yellow : [],
    );

    return Array.from(
      new Set(yellowDiseases.filter((name): name is string => Boolean(name?.trim()))),
    )
      .slice(0, 6)
      .map((name) => ({
        label: name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(name)),
      }));
  }, [filteredDistrictStatuses]);

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
          region={selectedRegionId}
          district={selectedDistrictId}
          subCounty={selectedSubCountyId}
          epiWeekOptions={epiWeekOptions}
          regionOptions={regionOptions}
          districtOptions={districtOptions}
          subCountyOptions={subCountyOptions}
          onEpiWeekChange={setSelectedWeekId}
          onRegionChange={setSelectedRegionId}
          onDistrictChange={setSelectedDistrictId}
          onSubCountyChange={setSelectedSubCountyId}
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
          region={selectedRegion?.name ?? ""}
          district={selectedDistrict?.name ?? ""}
          subCounty={selectedSubCounty?.name ?? ""}
          districtStatuses={filteredDistrictStatuses}
          regionStatuses={filteredRegionStatuses}
          nationalStatuses={nationalStatuses}
          facilityMetrics={filteredFacilityMetrics}
          loading={loading}
        />
      </section>
    </div>
  );
}
