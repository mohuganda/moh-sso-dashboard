import { Breadcrumb, BreadcrumbItem, Button, InlineLoading } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Upload } from "@carbon/react/icons";

import SurveillanceFilters from "./surveillance-filters.component";
import SurveillanceTiles from "./surveillance-tiles.component";
import SurveillancePanels from "./surveillance-panel.component";
import SurveillanceUgandaMap from "./surveillance-uganda-map.component";
import rawUgandaGeoJson from "../../../assets/maps/uganda_districts.json";
import type { GeoJSON as GeoJSONType } from "geojson";

const ugandaGeoJson = rawUgandaGeoJson as GeoJSONType;
import "./surveillance.css";

import {
  useListAlertsQuery,
  useListDiseasesQuery,
  useListDistrictsByRegionQuery,
  useListEpiWeeksQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
  useListRegionsQuery,
  useListSubcountiesByDistrictQuery,
  useListWeeklyStatusesDetailedQuery,
  useListWeeklyStatusesQuery,
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

function normalize(value: unknown): string {
  if (typeof value !== "string") return "";
  return value.trim().toLowerCase();
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

  const weeklyStatusParams = useMemo(() => {
    if (!selectedWeekId) return undefined;

    return {
      epiWeekID: selectedWeekId,
      regionID: selectedRegionId || undefined,
      districtID: selectedDistrictId || undefined,
      subCountyID: selectedSubCountyId || undefined,
    };
  }, [selectedWeekId, selectedRegionId, selectedDistrictId, selectedSubCountyId]);

  const { data: weeklyStatusesResponse, isFetching: weeklyStatusesLoading } =
    useListWeeklyStatusesQuery(weeklyStatusParams, {
      skip: !selectedWeekId,
    });

  const { data: weeklyStatusesDetailedResponse, isFetching: weeklyStatusesDetailedLoading } =
    useListWeeklyStatusesDetailedQuery(weeklyStatusParams, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetricsResponse, isFetching: facilitiesLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: alertsResponse, isFetching: alertsLoading } = useListAlertsQuery();

  const weeklyStatuses = Array.isArray(weeklyStatusesResponse) ? weeklyStatusesResponse : [];
  const weeklyStatusesDetailed = Array.isArray(weeklyStatusesDetailedResponse)
    ? weeklyStatusesDetailedResponse
    : [];
  const facilityMetrics = Array.isArray(facilityMetricsResponse) ? facilityMetricsResponse : [];
  const alerts = Array.isArray(alertsResponse) ? alertsResponse : [];

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
    weeklyStatusesLoading ||
    weeklyStatusesDetailedLoading ||
    facilitiesLoading ||
    alertsLoading;

  const selectedWeek = useMemo(
    () => weeks.find((week) => week.id === selectedWeekId),
    [weeks, selectedWeekId],
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

  const filteredWeeklyStatuses = useMemo(() => {
    return weeklyStatuses.filter((item) => {
      const matchesWeek = !selectedWeekId || item.epi_week_id === selectedWeekId;
      const matchesRegion = !selectedRegionId || item.region_id === selectedRegionId;
      const matchesDistrict = !selectedDistrictId || item.district_id === selectedDistrictId;
      const matchesSubCounty = !selectedSubCountyId || item.sub_county_id === selectedSubCountyId;

      return matchesWeek && matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [weeklyStatuses, selectedWeekId, selectedRegionId, selectedDistrictId, selectedSubCountyId]);

  const filteredWeeklyStatusesDetailed = useMemo(() => {
    return weeklyStatusesDetailed.filter((item) => {
      const matchesWeek = !selectedWeekId || item.epi_week_id === selectedWeekId;
      const matchesRegion = !selectedRegionId || item.region_id === selectedRegionId;
      const matchesDistrict = !selectedDistrictId || item.district_id === selectedDistrictId;
      const matchesSubCounty = !selectedSubCountyId || item.sub_county_id === selectedSubCountyId;

      return matchesWeek && matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [
    weeklyStatusesDetailed,
    selectedWeekId,
    selectedRegionId,
    selectedDistrictId,
    selectedSubCountyId,
  ]);

  const filteredFacilityMetrics = useMemo(() => {
    return facilityMetrics.filter((item) => {
      const matchesRegion =
        !selectedRegionId ||
        ("region_id" in item && String(item.region_id ?? "") === selectedRegionId);

      const matchesDistrict =
        !selectedDistrictId ||
        ("district_id" in item && String(item.district_id ?? "") === selectedDistrictId);

      const matchesSubCounty =
        !selectedSubCountyId ||
        ("sub_county_id" in item && String(item.sub_county_id ?? "") === selectedSubCountyId);

      return matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [facilityMetrics, selectedRegionId, selectedDistrictId, selectedSubCountyId]);

  const filteredAlerts = useMemo(() => {
    return alerts.filter((item) => {
      const matchesWeek = !selectedWeekId || item.epi_week_id === selectedWeekId;

      const selectedDistrictName = districts.find(
        (district) => district.id === selectedDistrictId,
      )?.name;

      const matchesDistrict =
        !selectedDistrictId ||
        String(item.district_id ?? "") === selectedDistrictId ||
        normalize(item.district_name) === normalize(selectedDistrictName);

      return matchesWeek && matchesDistrict;
    });
  }, [alerts, selectedWeekId, selectedDistrictId, districts]);

  const diseaseNameById = useMemo(() => {
    return new Map(diseases.map((disease) => [disease.id, disease.name]));
  }, [diseases]);

  const handleOpenDisease = (diseaseSlug: string) => {
    navigate({
      pathname: `/apps/dwh/surveillance/${encodeURIComponent(diseaseSlug)}`,
      search: selectedWeekId ? `?weekId=${encodeURIComponent(selectedWeekId)}` : "",
    });
  };

  const createStatusItems = (status: "MAROON" | "RED" | "YELLOW" | "GREEN") => {
    return Array.from(
      new Set(
        filteredWeeklyStatuses
          .filter((item) => item.status === status && item.disease_id)
          .map((item) => diseaseNameById.get(String(item.disease_id)))
          .filter((name): name is string => Boolean(name?.trim())),
      ),
    )
      .slice(0, 6)
      .map((name) => ({
        label: name,
        onClick: () => handleOpenDisease(slugifyDiseaseName(name)),
      }));
  };

  const immediateActionItems = useMemo(
    () => createStatusItems("MAROON"),
    [filteredWeeklyStatuses, diseaseNameById],
  );

  const takeActionItems = useMemo(
    () => createStatusItems("RED"),
    [filteredWeeklyStatuses, diseaseNameById],
  );

  const alertItems = useMemo(
    () => createStatusItems("YELLOW"),
    [filteredWeeklyStatuses, diseaseNameById],
  );

  const watchItems = useMemo(
    () => createStatusItems("GREEN"),
    [filteredWeeklyStatuses, diseaseNameById],
  );

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
          regionId={selectedRegionId}
          districtId={selectedDistrictId}
          subCountyId={selectedSubCountyId}
          weeklyStatuses={filteredWeeklyStatuses}
          weeklyStatusesDetailed={filteredWeeklyStatusesDetailed}
          facilityMetrics={filteredFacilityMetrics}
          alerts={filteredAlerts}
          loading={loading}
          mapContent={
            <SurveillanceUgandaMap
              geoJson={ugandaGeoJson}
              weeklyStatuses={filteredWeeklyStatuses}
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
