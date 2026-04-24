import { Breadcrumb, BreadcrumbItem, Button, InlineLoading } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Upload } from "@carbon/react/icons";
import type { Feature, FeatureCollection, GeoJsonProperties, Geometry } from "geojson";

import SurveillanceFilters from "./surveillance-filters.component";
import SurveillanceTiles from "./surveillance-tiles.component";
import SurveillancePanels from "./surveillance-panel.component";
import SurveillanceUgandaMap from "./surveillance-uganda-map.component";
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
import { useGetGeoJsonQuery } from "../../../store/api/geojson.api";
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";
import { UploadCSVModal } from "./surveillance-csv-upload.component";

type FilterOption = {
  value: string;
  label: string;
};

type GenericFeatureCollection = FeatureCollection<Geometry, GeoJsonProperties>;

function slugifyDiseaseName(name: string) {
  return name.trim().toLowerCase().replace(/\s+/g, "-");
}

function normalize(value: unknown): string {
  if (typeof value !== "string") return "";
  return value.trim().toLowerCase();
}

function asFeatureCollection(value: unknown): GenericFeatureCollection {
  if (
    value &&
    typeof value === "object" &&
    "type" in value &&
    (value as { type?: string }).type === "FeatureCollection" &&
    "features" in value &&
    Array.isArray((value as { features?: unknown[] }).features)
  ) {
    return value as GenericFeatureCollection;
  }

  return {
    type: "FeatureCollection",
    features: [],
  };
}

function getDistrictFeatureName(properties?: GeoJsonProperties | null) {
  if (!properties) return "";

  return normalize(
    properties.District ??
      properties.district ??
      properties.district_name ??
      properties.name ??
      properties.DISTRICT ??
      "",
  );
}

function getSubcountyFeatureName(properties?: GeoJsonProperties | null) {
  if (!properties) return "";

  return normalize(
    properties.Subcounty ??
      properties.sname2019 ??
      properties.subcounty ??
      properties.sub_county ??
      properties.subcounty_name ??
      properties.name ??
      properties.SCOUNTY ??
      "",
  );
}

function filterFeatureCollection(
  collection: GenericFeatureCollection,
  predicate: (feature: Feature<Geometry, GeoJsonProperties>) => boolean,
): GenericFeatureCollection {
  return {
    ...collection,
    features: collection.features.filter(predicate),
  };
}

export default function SurveillanceDashboardPage() {
  const navigate = useNavigate();
  const { openPanel, closePanel } = useHeaderPanel();

  const [selectedWeekId, setSelectedWeekId] = useState("");
  const [selectedRegionId, setSelectedRegionId] = useState("");
  const [selectedDistrictId, setSelectedDistrictId] = useState("");
  const [selectedSubCountyId, setSelectedSubCountyId] = useState("");
  const currentYear = new Date().getFullYear();

  const { data: weeksResponse, isLoading: weeksLoading } = useListEpiWeeksQuery(currentYear);
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

  const { data: districtsGeoJsonResponse, isLoading: districtsGeoJsonLoading } =
    useGetGeoJsonQuery("geo_districts");

  const { data: subcountiesGeoJsonResponse, isLoading: subcountiesGeoJsonLoading } =
    useGetGeoJsonQuery("geo_subcounties");

  const weeks = Array.isArray(weeksResponse) ? weeksResponse : [];
  const diseases = Array.isArray(diseasesResponse) ? diseasesResponse : [];
  const regions = Array.isArray(regionsResponse) ? regionsResponse : [];
  const districts = Array.isArray(districtsResponse) ? districtsResponse : [];
  const subcounties = Array.isArray(subcountiesResponse) ? subcountiesResponse : [];

  const districtGeoJson = useMemo(
    () => asFeatureCollection(districtsGeoJsonResponse),
    [districtsGeoJsonResponse],
  );

  const subcountyGeoJson = useMemo(
    () => asFeatureCollection(subcountiesGeoJsonResponse),
    [subcountiesGeoJsonResponse],
  );

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

  console.log("districtGeoJson features", districtGeoJson.features.length);
  console.log("subcountyGeoJson features", subcountyGeoJson.features.length);

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
    districtsGeoJsonLoading ||
    subcountiesGeoJsonLoading ||
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
    const params = new URLSearchParams();

    if (selectedWeekId) {
      params.set("weekId", selectedWeekId);
    }

    if (selectedRegionId) {
      params.set("regionId", selectedRegionId);
    }

    if (selectedDistrictId) {
      params.set("districtId", selectedDistrictId);
    }

    if (selectedSubCountyId) {
      params.set("subCountyId", selectedSubCountyId);
    }

    navigate({
      pathname: `/apps/dwh/surveillance/${encodeURIComponent(diseaseSlug)}`,
      search: params.toString() ? `?${params.toString()}` : "",
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
    [
      filteredWeeklyStatuses,
      diseaseNameById,
      selectedWeekId,
      selectedRegionId,
      selectedDistrictId,
      selectedSubCountyId,
    ],
  );

  const takeActionItems = useMemo(
    () => createStatusItems("RED"),
    [
      filteredWeeklyStatuses,
      diseaseNameById,
      selectedWeekId,
      selectedRegionId,
      selectedDistrictId,
      selectedSubCountyId,
    ],
  );

  const alertItems = useMemo(
    () => createStatusItems("YELLOW"),
    [
      filteredWeeklyStatuses,
      diseaseNameById,
      selectedWeekId,
      selectedRegionId,
      selectedDistrictId,
      selectedSubCountyId,
    ],
  );

  const watchItems = useMemo(
    () => createStatusItems("GREEN"),
    [
      filteredWeeklyStatuses,
      diseaseNameById,
      selectedWeekId,
      selectedRegionId,
      selectedDistrictId,
      selectedSubCountyId,
    ],
  );

  const regionDistrictNames = useMemo(() => {
    return new Set(districts.map((district) => normalize(district.name)).filter(Boolean));
  }, [districts]);

  const districtSubcountyNames = useMemo(() => {
    return new Set(subcounties.map((subcounty) => normalize(subcounty.name)).filter(Boolean));
  }, [subcounties]);

  const activeMapGeoJson = useMemo<GenericFeatureCollection>(() => {
    if (selectedDistrictId) {
      return filterFeatureCollection(subcountyGeoJson, (feature) => {
        const subcountyName = getSubcountyFeatureName(feature.properties);
        return districtSubcountyNames.has(subcountyName);
      });
    }

    if (selectedRegionId) {
      return filterFeatureCollection(districtGeoJson, (feature) => {
        const districtName = getDistrictFeatureName(feature.properties);
        return regionDistrictNames.has(districtName);
      });
    }

    return districtGeoJson;
  }, [
    selectedDistrictId,
    selectedRegionId,
    districtGeoJson,
    subcountyGeoJson,
    districtSubcountyNames,
    regionDistrictNames,
  ]);

  const activeMapLevel = selectedDistrictId ? "subcounty" : "district";

  const handleMapDistrictSelect = (districtName: string) => {
    const match = districts.find((item) => normalize(item.name) === normalize(districtName));
    if (match) {
      setSelectedDistrictId(match.id);
    }
  };

  const handleMapSubcountySelect = (subcountyName: string) => {
    const match = subcounties.find((item) => normalize(item.name) === normalize(subcountyName));
    if (match) {
      setSelectedSubCountyId(match.id);
    }
  };

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
          immediateAction={{ title: "Immediate Action", items: immediateActionItems }}
          takeAction={{ title: "Take Action", items: takeActionItems }}
          alert={{ title: "Alert", items: alertItems }}
          watch={{ title: "Watch", items: watchItems }}
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
          weeklyStatusesDetailed={filteredWeeklyStatusesDetailed}
          facilityMetrics={filteredFacilityMetrics}
          alerts={filteredAlerts}
          loading={loading}
          mapContent={
            <SurveillanceUgandaMap
              geoJson={activeMapGeoJson}
              mapLevel={activeMapLevel}
              weeklyStatuses={filteredWeeklyStatusesDetailed}
              selectedWeek={{
                year: selectedWeek?.epi_year,
                week: selectedWeek?.epi_week,
              }}
              onDistrictSelect={handleMapDistrictSelect}
              onSubCountySelect={handleMapSubcountySelect}
            />
          }
        />
      </section>
    </div>
  );
}
