import { Breadcrumb, BreadcrumbItem, Button, InlineLoading } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Upload } from "@carbon/react/icons";

import SurveillanceFilters from "./surveillance-filters.component";
import SurveillanceTiles from "./surveillance-tiles.component";
import SurveillancePanels from "./surveillance-panel.component";
import SurveillanceUgandaMap from "./surveillance-uganda-map.component";
import ugandaGeoJson from "../../../assets/maps/uganda_districts.json";
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
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";
import { UploadCSVModal } from "./surveillance-csv-upload.component";

type FilterOption = {
  value: string;
  label: string;
};

function slugifyDiseaseName(name: string) {
  return name.trim().toLowerCase().replace(/\s+/g, "-");
}

export default function SurveillanceDashboardPage() {
  const navigate = useNavigate();
  const { openPanel, closePanel } = useHeaderPanel();

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

    return selectedWeek.epi_year && selectedWeek.epi_week
      ? `${selectedWeek.epi_year} / Week ${selectedWeek.epi_week}`
      : `Week ${selectedWeek.epi_week ?? "--"}`;
  }, [selectedWeek]);

  const epiWeekOptions: FilterOption[] = useMemo(
    () => [
      { value: "", label: "Select..." },
      ...weeks.map((week) => ({
        value: week.id,
        label:
          week.epi_year && week.epi_week
            ? `${week.epi_year} - Week ${week.epi_week}`
            : `Week ${week.epi_week ?? ""}`,
      })),
    ],
    [weeks],
  );

  const regionOptions: FilterOption[] = useMemo(
    () => [
      { value: "", label: "Select..." },
      ...regions.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ],
    [regions],
  );

  const districtOptions: FilterOption[] = useMemo(
    () => [
      { value: "", label: "Select..." },
      ...districts.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ],
    [districts],
  );

  const subCountyOptions: FilterOption[] = useMemo(
    () => [
      { value: "", label: "Select..." },
      ...subcounties.map((item) => ({
        value: item.id,
        label: item.name,
      })),
    ],
    [subcounties],
  );

  const filteredRegionStatuses = useMemo(() => {
    if (!selectedRegion?.name) return regionStatuses;

    const regionName = selectedRegion.name.trim().toLowerCase();
    return regionStatuses.filter((item) => item.region_name?.trim().toLowerCase() === regionName);
  }, [regionStatuses, selectedRegion]);

  const filteredDistrictStatuses = useMemo(() => {
    let result = districtStatuses;

    if (selectedRegion?.name) {
      const regionName = selectedRegion.name.trim().toLowerCase();
      result = result.filter((item) => item.region_name?.trim().toLowerCase() === regionName);
    }

    if (selectedDistrict?.name) {
      const districtName = selectedDistrict.name.trim().toLowerCase();
      result = result.filter((item) => item.district_name?.trim().toLowerCase() === districtName);
    }

    return result;
  }, [districtStatuses, selectedRegion, selectedDistrict]);

  const filteredFacilityMetrics = useMemo(() => {
    let result = facilityMetrics;

    if (selectedRegion?.name) {
      const regionName = selectedRegion.name.trim().toLowerCase();
      result = result.filter((item) => item.region_name?.trim().toLowerCase() === regionName);
    }

    if (selectedDistrict?.name) {
      const districtName = selectedDistrict.name.trim().toLowerCase();
      result = result.filter((item) => item.district_name?.trim().toLowerCase() === districtName);
    }

    if (selectedSubCounty?.name) {
      const subCountyName = selectedSubCounty.name.trim().toLowerCase();
      result = result.filter((item) => item.subcounty_name?.trim().toLowerCase() === subCountyName);
    }

    return result;
  }, [facilityMetrics, selectedRegion, selectedDistrict, selectedSubCounty]);

  const handleOpenDisease = (diseaseSlug: string) => {
    navigate(`/apps/dwh/surveillance/${encodeURIComponent(diseaseSlug)}`);
  };

  const immediateActionItems = useMemo(
    () =>
      diseases.slice(0, 4).map((disease) => ({
        label: disease.name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(disease.name)),
      })),
    [diseases],
  );

  const takeActionItems = useMemo(
    () =>
      diseases.slice(4, 8).map((disease) => ({
        label: disease.name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(disease.name)),
      })),
    [diseases],
  );

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

  const handleOpenUpload = () => {
    openPanel({
      title: "Import Surveillance File",
      content: <UploadCSVModal onClose={closePanel} />,
      size: "lg",
    });
  };

  return (
    <div className="surveillance-dashboard-page">
      <div className="surveillance-dashboard-page__header">
        <Breadcrumb className="surveillance-dashboard-page__breadcrumb" noTrailingSlash>
          <BreadcrumbItem href="/portal/apps/dwh/data-visualizer">
            Data &amp; Statistics
          </BreadcrumbItem>

          <BreadcrumbItem isCurrentPage>
            <span>National Surveillance Dashboard</span>
          </BreadcrumbItem>
        </Breadcrumb>

        <div className="surveillance-dashboard-page__hero">
          <div className="surveillance-dashboard-page__hero-main">
            <div className="surveillance-dashboard-page__hero-copy">
              <p className="surveillance-dashboard-page__eyebrow">
                Epi Week: {currentEpiWeekLabel}
              </p>

              <h1 className="surveillance-dashboard-page__title">
                National Surveillance Reporting Dashboard
              </h1>

              <p className="surveillance-dashboard-page__subtitle">
                Monitor surveillance signals, reporting trends, alerts, and priority conditions
                across regions, districts, and sub-counties.
              </p>
            </div>

            <div className="surveillance-dashboard-page__hero-actions">
              <Button renderIcon={Upload} onClick={handleOpenUpload}>
                Import File
              </Button>
            </div>
          </div>

          {loading ? (
            <div className="surveillance-dashboard-page__loading">
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
          selectedWeekId={selectedWeek?.id}
          selectedWeek={{
            id: selectedWeek?.id,
            year: selectedWeek?.epi_year,
            week: selectedWeek?.epi_week,
          }}
          region={selectedRegion?.name ?? ""}
          district={selectedDistrict?.name ?? ""}
          subCounty={selectedSubCounty?.name ?? ""}
          districtStatuses={filteredDistrictStatuses}
          regionStatuses={filteredRegionStatuses}
          nationalStatuses={nationalStatuses}
          facilityMetrics={filteredFacilityMetrics}
          loading={loading}
          mapContent={
            <SurveillanceUgandaMap
              geoJson={ugandaGeoJson}
              districtStatuses={filteredDistrictStatuses}
              regionStatuses={filteredRegionStatuses}
              nationalStatuses={nationalStatuses}
              selectedWeek={{
                year: selectedWeek?.epi_year,
                week: selectedWeek?.epi_week,
              }}
            />
          }
        />
      </section>
    </div>
  );
}
