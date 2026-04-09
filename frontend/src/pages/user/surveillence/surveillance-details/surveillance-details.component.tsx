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
import { useMemo } from "react";
import { useParams, Link as RouterLink } from "react-router-dom";

import { DiseaseAlertsTable } from "./surveillance-disease-alerts.component";
import { FacilitiesActionTable } from "./surveillance-facilities-actions.component";
import {
  useListEpiWeeksQuery,
  useListDistrictWeeklyStatusesByWeekQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
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
  return (value ?? "").trim().toLowerCase();
}

type RelevantDocument = {
  label: string;
  href: string;
};

export default function DiseaseDetailsPage() {
  const { diseaseName } = useParams<{ diseaseName: string }>();

  const normalizedDiseaseSlug = normalize(diseaseName).replace(/-/g, " ");
  const title = formatDiseaseName(diseaseName);
  const currentYear = new Date().getFullYear();

  const { data: weeksResponse, isLoading: weeksLoading } = useListEpiWeeksQuery(currentYear);
  const weeks = Array.isArray(weeksResponse) ? weeksResponse : [];

  const selectedWeek = useMemo(() => {
    if (weeks.length === 0) return undefined;

    return [...weeks].sort((a, b) => {
      const yearDiff = Number(b.epi_year ?? 0) - Number(a.epi_year ?? 0);
      if (yearDiff !== 0) return yearDiff;
      return Number(b.epi_week ?? 0) - Number(a.epi_week ?? 0);
    })[0];
  }, [weeks]);

  const selectedWeekId = selectedWeek?.id ?? "";

  const { data: districtStatusesResponse, isFetching: districtStatusesLoading } =
    useListDistrictWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetricsResponse, isFetching: facilityMetricsLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const districtStatuses = Array.isArray(districtStatusesResponse) ? districtStatusesResponse : [];
  const facilityMetrics = Array.isArray(facilityMetricsResponse) ? facilityMetricsResponse : [];

  const loading = weeksLoading || districtStatusesLoading || facilityMetricsLoading;

  const filteredFacilityMetrics = useMemo(() => {
    return facilityMetrics.filter((item) => {
      const diseaseLabel = normalize(item.disease_name ?? item.indicator_name).replace(/-/g, " ");
      return diseaseLabel.includes(normalizedDiseaseSlug);
    });
  }, [facilityMetrics, normalizedDiseaseSlug]);

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
    return filteredFacilityMetrics.reduce((sum, item) => sum + Number(item.value ?? 0), 0);
  }, [filteredFacilityMetrics]);

  const facilitiesCount = useMemo(() => {
    const uniqueFacilities = new Set(
      filteredFacilityMetrics
        .map((item) => item.facility_name)
        .filter((value): value is string => Boolean(value?.trim())),
    );

    return uniqueFacilities.size;
  }, [filteredFacilityMetrics]);

  const weeklyChartData = useMemo(() => {
    const grouped = new Map<string, number>();

    filteredFacilityMetrics.forEach((item) => {
      const weekKey =
        item.week != null
          ? String(item.week)
          : item.epi_week != null
            ? String(item.epi_week)
            : "Unknown";

      const currentValue = grouped.get(weekKey) ?? 0;
      grouped.set(weekKey, currentValue + Number(item.value ?? 0));
    });

    return Array.from(grouped.entries())
      .map(([week, value]) => ({
        week,
        value,
      }))
      .sort((a, b) => Number(a.week) - Number(b.week));
  }, [filteredFacilityMetrics]);

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
                districtStatuses={districtStatuses}
                diseaseName={title}
                loading={loading}
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
            <FacilitiesActionTable facilityMetrics={filteredFacilityMetrics} loading={loading} />
          </div>
        </Tile>
      </div>
    </Content>
  );
}
