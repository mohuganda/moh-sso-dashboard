import {
  Breadcrumb,
  BreadcrumbItem,
  Content,
  Grid,
  Column,
  Tile,
  Link,
  InlineLoading,
} from "@carbon/react";
import { useMemo, useState } from "react";
import { useParams, useSearchParams, Link as RouterLink } from "react-router-dom";

import { DiseaseAlertsTable } from "./surveillance-disease-alerts.component";
import { FacilitiesActionTable } from "./surveillance-facilities-actions.component";
import { FacilityTrendModal } from "./surveillance-facility-trend-modal.component";
import {
  useListAlertsQuery,
  useListDiseasesQuery,
  useListEpiWeeksQuery,
  useListFacilityDiseaseMetricsTrendQuery,
  useListFacilityIndicatorMetricsTrendQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
  useListRegionsQuery,
  useListWeeklyStatusesDetailedQuery,
} from "../../../../store/api/surveillance.api";
import WeeklyCasesChart from "./surveillance-weekly-cases.component";
import "./surveillance-details.css";

function formatDiseaseName(value?: string) {
  if (!value) return "Disease";

  return value
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function normalize(value?: string) {
  return String(value ?? "")
    .trim()
    .toLowerCase();
}

type RelevantDocument = {
  label: string;
  href: string;
};

type FacilityTrendSelection = {
  facilityId?: string;
  facilityName: string;
  regionId?: string;
  districtId?: string;
  subCountyId?: string;
  diseaseId?: string;
  indicatorId?: string;
  diseaseName?: string;
  indicatorName?: string;
};

export default function DiseaseDetailsPage() {
  const { diseaseName } = useParams<{ diseaseName: string }>();
  const [searchParams] = useSearchParams();

  const requestedWeekId = searchParams.get("weekId") ?? "";
  const selectedRegionId = searchParams.get("regionId") ?? "";
  const selectedDistrictId = searchParams.get("districtId") ?? "";
  const selectedSubCountyId = searchParams.get("subCountyId") ?? "";

  const [selectedFacilityTrend, setSelectedFacilityTrend] = useState<FacilityTrendSelection | null>(
    null,
  );
  const [isTrendModalOpen, setIsTrendModalOpen] = useState(false);

  const normalizedDiseaseSlug = normalize(diseaseName).replace(/-/g, " ");
  const title = formatDiseaseName(diseaseName);
  const currentYear = new Date().getFullYear();

  const { data: weeksResponse, isLoading: weeksLoading } = useListEpiWeeksQuery(currentYear);
  const { data: diseasesResponse, isLoading: diseasesLoading } = useListDiseasesQuery();
  const { data: regionsResponse, isLoading: regionsLoading } = useListRegionsQuery();

  const weeks = Array.isArray(weeksResponse) ? weeksResponse : [];
  const diseases = Array.isArray(diseasesResponse) ? diseasesResponse : [];
  const regions = Array.isArray(regionsResponse) ? regionsResponse : [];

  const selectedWeek = useMemo(() => {
    if (!weeks.length) return undefined;

    if (requestedWeekId) {
      const requestedWeek = weeks.find((week) => week.id === requestedWeekId);
      if (requestedWeek) {
        return requestedWeek;
      }
    }

    return [...weeks].sort((a, b) => {
      const yearDiff = Number(b.epi_year ?? 0) - Number(a.epi_year ?? 0);
      if (yearDiff !== 0) return yearDiff;
      return Number(b.epi_week ?? 0) - Number(a.epi_week ?? 0);
    })[0];
  }, [weeks, requestedWeekId]);

  const matchedDisease = useMemo(() => {
    return diseases.find((item) => normalize(item.name) === normalizedDiseaseSlug);
  }, [diseases, normalizedDiseaseSlug]);

  const selectedWeekId = selectedWeek?.id ?? "";
  const matchedDiseaseId = matchedDisease?.id ?? "";

  const baseParams = useMemo(() => {
    if (!selectedWeekId) return undefined;

    return {
      epiWeekID: selectedWeekId,
      diseaseID: matchedDiseaseId || undefined,
    };
  }, [selectedWeekId, matchedDiseaseId]);

  const { data: alertsResponse, isFetching: alertsLoading } = useListAlertsQuery(baseParams, {
    skip: !selectedWeekId,
  });

  const { data: weeklyStatusesDetailedResponse, isFetching: weeklyStatusesDetailedLoading } =
    useListWeeklyStatusesDetailedQuery(baseParams, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetricsResponse, isFetching: facilityMetricsLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const alerts = Array.isArray(alertsResponse) ? alertsResponse : [];
  const weeklyStatusesDetailed = Array.isArray(weeklyStatusesDetailedResponse)
    ? weeklyStatusesDetailedResponse
    : [];
  const facilityMetrics = Array.isArray(facilityMetricsResponse) ? facilityMetricsResponse : [];

  const diseaseNameById = useMemo<Record<string, string>>(() => {
    return Object.fromEntries(diseases.map((item) => [item.id, item.name]));
  }, [diseases]);

  const indicatorNameById = useMemo<Record<string, string>>(() => {
    const map = new Map<string, string>();

    facilityMetrics.forEach((item) => {
      const indicatorId = String(item.indicator_id ?? "");
      const indicatorName = item.indicator_name ?? "";

      if (indicatorId && indicatorName) {
        map.set(indicatorId, indicatorName);
      }
    });

    weeklyStatusesDetailed.forEach((item) => {
      const indicatorId = String(item.indicator_id ?? "");
      const indicatorName = item.indicator_name ?? "";

      if (indicatorId && indicatorName) {
        map.set(indicatorId, indicatorName);
      }
    });

    return Object.fromEntries(map.entries());
  }, [facilityMetrics, weeklyStatusesDetailed]);

  const regionNameById = useMemo<Record<string, string>>(() => {
    const map = new Map<string, string>();

    regions.forEach((item) => {
      map.set(item.id, item.name);
    });

    weeklyStatusesDetailed.forEach((item) => {
      const regionId = item.region_id ? String(item.region_id) : "";
      const regionName = item.region_name ?? "";
      if (regionId && regionName) {
        map.set(regionId, regionName);
      }
    });

    facilityMetrics.forEach((item) => {
      const regionId = String(item.region_id ?? "");
      const regionName = item.region_name ?? "";
      if (regionId && regionName) {
        map.set(regionId, regionName);
      }
    });

    return Object.fromEntries(map.entries());
  }, [regions, weeklyStatusesDetailed, facilityMetrics]);

  const districtNameById = useMemo<Record<string, string>>(() => {
    const map = new Map<string, string>();

    weeklyStatusesDetailed.forEach((item) => {
      const districtId = item.district_id ? String(item.district_id) : "";
      const districtName = item.district_name ?? "";

      if (districtId && districtName) {
        map.set(districtId, districtName);
      }
    });

    facilityMetrics.forEach((item) => {
      const districtId = String(item.district_id ?? "");
      const districtName = item.district_name ?? "";

      if (districtId && districtName) {
        map.set(districtId, districtName);
      }
    });

    alerts.forEach((item) => {
      const districtId = String(item.district_id ?? "");
      const districtName = item.district_name ?? "";

      if (districtId && districtName) {
        map.set(districtId, districtName);
      }
    });

    return Object.fromEntries(map.entries());
  }, [weeklyStatusesDetailed, facilityMetrics, alerts]);

  const districtRegionByDistrictId = useMemo<Record<string, string>>(() => {
    const map = new Map<string, string>();

    weeklyStatusesDetailed.forEach((item) => {
      const districtId = item.district_id ? String(item.district_id) : "";
      const regionId = item.region_id ? String(item.region_id) : "";

      if (districtId && regionId) {
        map.set(districtId, regionId);
      }
    });

    return Object.fromEntries(map.entries());
  }, [weeklyStatusesDetailed]);

  const subCountyNameById = useMemo<Record<string, string>>(() => {
    const map = new Map<string, string>();

    weeklyStatusesDetailed.forEach((item) => {
      const subCountyId = item.sub_county_id ? String(item.sub_county_id) : "";
      const subCountyName = item.sub_county_name ?? "";

      if (subCountyId && subCountyName) {
        map.set(subCountyId, subCountyName);
      }
    });

    facilityMetrics.forEach((item) => {
      const subCountyId = String(item.sub_county_id ?? "");
      const subCountyName = item.subcounty_name ?? "";

      if (subCountyId && subCountyName) {
        map.set(subCountyId, subCountyName);
      }
    });

    return Object.fromEntries(map.entries());
  }, [weeklyStatusesDetailed, facilityMetrics]);

  const epiWeekLabelById = useMemo<Record<string, string>>(() => {
    return Object.fromEntries(
      weeks.map((week) => [
        week.id,
        week.epi_year && week.epi_week
          ? `${week.epi_year} - Week ${week.epi_week}`
          : `Week ${week.epi_week ?? "--"}`,
      ]),
    );
  }, [weeks]);

  const loading =
    weeksLoading ||
    diseasesLoading ||
    regionsLoading ||
    alertsLoading ||
    weeklyStatusesDetailedLoading ||
    facilityMetricsLoading;

  const filteredAlerts = useMemo(() => {
    return alerts.filter((item) => {
      if (matchedDiseaseId) {
        return String(item.disease_id ?? "") === matchedDiseaseId;
      }

      const diseaseLabel = normalize(item.disease_name ?? "").replace(/-/g, " ");
      return diseaseLabel.includes(normalizedDiseaseSlug);
    });
  }, [alerts, matchedDiseaseId, normalizedDiseaseSlug]);

  const diseaseOrIndicatorFacilityMetrics = useMemo(() => {
    return facilityMetrics.filter((item) => {
      if (matchedDiseaseId) {
        return String(item.disease_id ?? "") === matchedDiseaseId;
      }

      return normalize(item.disease_name ?? item.indicator_name)
        .replace(/-/g, " ")
        .includes(normalizedDiseaseSlug);
    });
  }, [facilityMetrics, matchedDiseaseId, normalizedDiseaseSlug]);

  const filteredFacilityMetrics = useMemo(() => {
    return diseaseOrIndicatorFacilityMetrics.filter((item) => {
      const matchesRegion = !selectedRegionId || String(item.region_id ?? "") === selectedRegionId;
      const matchesDistrict =
        !selectedDistrictId || String(item.district_id ?? "") === selectedDistrictId;
      const matchesSubCounty =
        !selectedSubCountyId || String(item.sub_county_id ?? "") === selectedSubCountyId;

      return matchesRegion && matchesDistrict && matchesSubCounty;
    });
  }, [
    diseaseOrIndicatorFacilityMetrics,
    selectedRegionId,
    selectedDistrictId,
    selectedSubCountyId,
  ]);

  const relevantDocuments = useMemo<RelevantDocument[]>(() => {
    const documents: Record<string, RelevantDocument[]> = {
      malaria: [
        { label: "Malaria Surveillance Guidelines", href: "/documents/malaria-guidelines" },
        { label: "Malaria Case Investigation Form", href: "/documents/malaria-cif" },
      ],
      measles: [
        { label: "Measles Surveillance Guidelines", href: "/documents/measles-guidelines" },
        { label: "Measles Case Investigation Form", href: "/documents/measles-cif" },
      ],
      mpox: [
        { label: "Mpox Surveillance Guidelines", href: "/documents/mpox-guidelines" },
        { label: "Mpox Case Investigation Form", href: "/documents/mpox-cif" },
      ],
      "yellow fever": [
        {
          label: "Yellow Fever Surveillance Guidelines",
          href: "/documents/yellow-fever-guidelines",
        },
        { label: "Yellow Fever Case Investigation Form", href: "/documents/yellow-fever-cif" },
      ],
      anthrax: [
        { label: "Anthrax Surveillance Guidelines", href: "/documents/anthrax-guidelines" },
        { label: "Anthrax Case Investigation Form", href: "/documents/anthrax-cif" },
      ],
      plague: [
        { label: "Plague Surveillance Guidelines", href: "/documents/plague-guidelines" },
        { label: "Plague Case Investigation Form", href: "/documents/plague-cif" },
      ],
    };

    return (
      documents[normalizedDiseaseSlug] ?? [
        { label: `${title} Surveillance Guidelines`, href: "/documents" },
        { label: `${title} Data Collection Tools`, href: "/documents" },
      ]
    );
  }, [normalizedDiseaseSlug, title]);

  const totalCases = useMemo(() => {
    return diseaseOrIndicatorFacilityMetrics.reduce(
      (sum, item) => sum + Number(item.metric_value ?? 0),
      0,
    );
  }, [diseaseOrIndicatorFacilityMetrics]);

  const facilitiesCount = useMemo(() => {
    const uniqueFacilities = new Set(
      diseaseOrIndicatorFacilityMetrics
        .map((item) => item.facility_name)
        .filter((value): value is string => Boolean(value?.trim())),
    );

    return uniqueFacilities.size;
  }, [diseaseOrIndicatorFacilityMetrics]);

  const weeklyChartData = useMemo(() => {
    const grouped = new Map<string, number>();

    diseaseOrIndicatorFacilityMetrics.forEach((item) => {
      const weekKey = item.epi_week_id
        ? (epiWeekLabelById[String(item.epi_week_id)] ??
          ("week" in item && item.week != null ? String(item.week) : "Unknown"))
        : "week" in item && item.week != null
          ? String(item.week)
          : "Unknown";

      const currentValue = grouped.get(weekKey) ?? 0;
      grouped.set(weekKey, currentValue + Number(item.metric_value ?? 0));
    });

    return Array.from(grouped.entries()).map(([week, value]) => ({
      week,
      label: week,
      value,
    }));
  }, [diseaseOrIndicatorFacilityMetrics, epiWeekLabelById]);

  const shouldLoadDiseaseTrend =
    isTrendModalOpen &&
    Boolean(selectedFacilityTrend?.facilityId) &&
    Boolean(selectedFacilityTrend?.diseaseId);

  const shouldLoadIndicatorTrend =
    isTrendModalOpen &&
    Boolean(selectedFacilityTrend?.facilityId) &&
    Boolean(selectedFacilityTrend?.indicatorId);

  const { data: facilityDiseaseTrendResponse, isFetching: facilityDiseaseTrendLoading } =
    useListFacilityDiseaseMetricsTrendQuery(
      {
        facilityID: selectedFacilityTrend?.facilityId ?? "",
        diseaseID: selectedFacilityTrend?.diseaseId ?? "",
      },
      {
        skip: !shouldLoadDiseaseTrend,
      },
    );

  const { data: facilityIndicatorTrendResponse, isFetching: facilityIndicatorTrendLoading } =
    useListFacilityIndicatorMetricsTrendQuery(
      {
        facilityID: selectedFacilityTrend?.facilityId ?? "",
        indicatorID: selectedFacilityTrend?.indicatorId ?? "",
      },
      {
        skip: !shouldLoadIndicatorTrend,
      },
    );

  const selectedFacilityTrendRows = useMemo(() => {
    if (selectedFacilityTrend?.diseaseId) {
      return Array.isArray(facilityDiseaseTrendResponse) ? facilityDiseaseTrendResponse : [];
    }

    if (selectedFacilityTrend?.indicatorId) {
      return Array.isArray(facilityIndicatorTrendResponse) ? facilityIndicatorTrendResponse : [];
    }

    return [];
  }, [selectedFacilityTrend, facilityDiseaseTrendResponse, facilityIndicatorTrendResponse]);

  const selectedFacilityTrendData = useMemo(() => {
    return selectedFacilityTrendRows
      .map((item) => ({
        week: item.epi_week ?? 0,
        label:
          item.epi_year && item.epi_week
            ? `${item.epi_year} - Week ${item.epi_week}`
            : `Week ${item.epi_week ?? "--"}`,
        value: Number(item.metric_value ?? 0),
      }))
      .filter((item) => Number(item.week) > 0)
      .sort((a, b) => Number(a.week) - Number(b.week));
  }, [selectedFacilityTrendRows]);

  const selectedFacilityTotalCases = useMemo(() => {
    return selectedFacilityTrendRows.reduce((sum, item) => {
      return sum + Number(item.metric_value ?? 0);
    }, 0);
  }, [selectedFacilityTrendRows]);

  const selectedFacilityPeakWeek = useMemo(() => {
    if (!selectedFacilityTrendRows.length) return "--";

    const peak = [...selectedFacilityTrendRows].sort(
      (a, b) => Number(b.metric_value ?? 0) - Number(a.metric_value ?? 0),
    )[0];

    return peak?.epi_week ? `Week ${peak.epi_week}` : "--";
  }, [selectedFacilityTrendRows]);

  const selectedFacilityTrendLoading = facilityDiseaseTrendLoading || facilityIndicatorTrendLoading;

  const trendSubjectLabel =
    selectedFacilityTrend?.diseaseName ?? selectedFacilityTrend?.indicatorName ?? title;

  const handleCloseTrendModal = () => {
    setIsTrendModalOpen(false);
    setSelectedFacilityTrend(null);
  };

  return (
    <Content className="disease-details-page">
      <Breadcrumb className="disease-details-page__breadcrumb" noTrailingSlash>
        <BreadcrumbItem>
          <RouterLink to="/apps/dwh/surveillance">Surveillance</RouterLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>{title}</BreadcrumbItem>
      </Breadcrumb>

      <div className="disease-details-page__header">
        <h1>{title}</h1>
        <p>
          View disease-specific alerts, weekly case trends, relevant documents, and facilities that
          require immediate follow-up action.
        </p>

        {loading ? <InlineLoading description="Loading disease details..." /> : null}
      </div>

      <Grid fullWidth className="disease-details-page__grid">
        <Column lg={8} md={4} sm={4}>
          <Tile className="disease-details-page__tile">
            <div className="disease-details-page__section">
              <h3>Relevant Documents</h3>
              <ul className="disease-details-page__links">
                {relevantDocuments.map((doc) => (
                  <li key={doc.label}>
                    <Link as={RouterLink} to={doc.href}>
                      {doc.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>

            <div className="disease-details-page__section">
              <h3>Disease Alerts</h3>
              <DiseaseAlertsTable
                alerts={filteredAlerts}
                diseaseName={title}
                loading={loading}
                districtNameById={districtNameById}
                regionNameById={regionNameById}
                diseaseNameById={diseaseNameById}
                epiWeekLabelById={epiWeekLabelById}
                districtRegionByDistrictId={districtRegionByDistrictId}
              />
            </div>
          </Tile>
        </Column>

        <Column lg={8} md={4} sm={4}>
          <Tile className="disease-details-page__tile">
            <div className="disease-details-page__section">
              <h3>{title} Weekly Cases</h3>

              <dl className="disease-details-page__stats">
                <div>
                  <dt>Reporting Week</dt>
                  <dd>
                    {selectedWeek
                      ? `Week ${selectedWeek.epi_week}, ${selectedWeek.epi_year}`
                      : "--"}
                  </dd>
                </div>

                <div>
                  <dt>Total Cases</dt>
                  <dd>{totalCases}</dd>
                </div>

                <div>
                  <dt>Facilities Reporting</dt>
                  <dd>{facilitiesCount}</dd>
                </div>
              </dl>

              <WeeklyCasesChart diseaseName={title} data={weeklyChartData} />
            </div>
          </Tile>
        </Column>
      </Grid>

      <div className="disease-details-page__bottom">
        <Tile className="disease-details-page__tile">
          <div className="disease-details-page__section">
            <h3>Facilities requiring Action</h3>
            <FacilitiesActionTable
              facilityMetrics={filteredFacilityMetrics}
              regionId={selectedRegionId}
              districtId={selectedDistrictId}
              subCountyId={selectedSubCountyId}
              regionNameById={regionNameById}
              districtNameById={districtNameById}
              subCountyNameById={subCountyNameById}
              diseaseNameById={diseaseNameById}
              indicatorNameById={indicatorNameById}
              epiWeekLabelById={epiWeekLabelById}
              onViewFacilityTrend={(payload) => {
                setSelectedFacilityTrend(payload);
                setIsTrendModalOpen(true);
              }}
            />
          </div>
        </Tile>
      </div>

      <FacilityTrendModal
        open={isTrendModalOpen}
        selectedFacilityTrend={selectedFacilityTrend}
        trendSubjectLabel={trendSubjectLabel}
        selectedFacilityTotalCases={selectedFacilityTotalCases}
        selectedFacilityPeakWeek={selectedFacilityPeakWeek}
        selectedFacilityTrendData={selectedFacilityTrendData}
        loading={selectedFacilityTrendLoading}
        onClose={handleCloseTrendModal}
      />
    </Content>
  );
}
