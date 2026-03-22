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

import "../surveillance-details/surveillance-details.css";
import { DiseaseAlertsTable } from "./surveillance-disease-alerts.component";
import { FacilitiesActionTable } from "./surveillance-facilities-actions.component";
import {
  useListEpiWeeksQuery,
  useListDistrictWeeklyStatusesByWeekQuery,
  useListFacilityWeeklyMetricsByWeekQuery,
} from "../../../../store/api/surveillance.api";

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
  const { diseaseName } = useParams();
  const title = formatDiseaseName(diseaseName);

  const { data: weeks = [], isLoading: weeksLoading } = useListEpiWeeksQuery();

  const selectedWeekId = useMemo(() => {
    return weeks.length > 0 ? weeks[0].id : "";
  }, [weeks]);

  const selectedWeek = useMemo(() => {
    return weeks.find((week) => week.id === selectedWeekId);
  }, [weeks, selectedWeekId]);

  const { data: districtStatuses = [], isFetching: districtStatusesLoading } =
    useListDistrictWeeklyStatusesByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const { data: facilityMetrics = [], isFetching: facilityMetricsLoading } =
    useListFacilityWeeklyMetricsByWeekQuery(selectedWeekId, {
      skip: !selectedWeekId,
    });

  const loading = weeksLoading || districtStatusesLoading || facilityMetricsLoading;

  const filteredFacilityMetrics = useMemo(() => {
    const disease = normalize(title);

    return facilityMetrics.filter((item) => {
      const diseaseLabel = normalize(item.disease_name ?? item.indicator_name);
      return diseaseLabel.includes(disease);
    });
  }, [facilityMetrics, title]);

  const relevantDocuments = useMemo<RelevantDocument[]>(() => {
    const disease = normalize(title);

    const documents: Record<string, RelevantDocument[]> = {
      malaria: [
        { label: "Malaria Surveillance Guidelines", href: "#" },
        { label: "Malaria Case Investigation Form", href: "#" },
      ],
      measles: [
        { label: "Measles Surveillance Guidelines", href: "#" },
        { label: "Measles Case Investigation Form", href: "#" },
      ],
      mpox: [
        { label: "Mpox Surveillance Guidelines", href: "#" },
        { label: "Mpox Case Investigation Form", href: "#" },
      ],
      "yellow fever": [
        { label: "Yellow Fever Surveillance Guidelines", href: "#" },
        { label: "Yellow Fever Case Investigation Form", href: "#" },
      ],
      anthrax: [
        { label: "Anthrax Surveillance Guidelines", href: "#" },
        { label: "Anthrax Case Investigation Form", href: "#" },
      ],
      plague: [
        { label: "Plague Surveillance Guidelines", href: "#" },
        { label: "Plague Case Investigation Form", href: "#" },
      ],
    };

    return (
      documents[disease] ?? [
        { label: `${title} Surveillance Guidelines`, href: "#" },
        { label: `${title} Data Collection Tools`, href: "#" },
      ]
    );
  }, [title]);

  const totalCases = useMemo(() => {
    return filteredFacilityMetrics.reduce((sum, item) => sum + Number(item.value ?? 0), 0);
  }, [filteredFacilityMetrics]);

  const facilitiesCount = useMemo(() => {
    const uniqueFacilities = new Set(
      filteredFacilityMetrics
        .map((item) => item.facility_name)
        .filter((value): value is string => Boolean(value)),
    );

    return uniqueFacilities.size;
  }, [filteredFacilityMetrics]);

  return (
    <Content className="disease-details-page">
      <Breadcrumb noTrailingSlash>
        <BreadcrumbItem>
          <RouterLink to="/portal/surveillance">Surveillance</RouterLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>{title}</BreadcrumbItem>
      </Breadcrumb>

      <div className="disease-details-page__header">
        <h1>Details for {title}</h1>
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
                    <Link href={doc.href}>{doc.label}</Link>
                  </li>
                ))}
              </ul>
            </div>

            <div className="disease-details-page__section">
              <h3>Disease Alerts</h3>
              <DiseaseAlertsTable
                districtStatuses={districtStatuses}
                diseaseName={title}
                district=""
                region=""
                loading={loading}
              />
            </div>
          </Tile>
        </Column>

        <Column lg={8} md={4} sm={4}>
          <Tile className="disease-details-page__tile">
            <div className="disease-details-page__section">
              <h3>{title} Weekly Cases</h3>

              <div className="disease-details-page__stats">
                <p>
                  <strong>Reporting Week:</strong>{" "}
                  {selectedWeek ? `Week ${selectedWeek.week}, ${selectedWeek.year}` : "--"}
                </p>
                <p>
                  <strong>Total Cases:</strong> {totalCases}
                </p>
                <p>
                  <strong>Facilities Reporting:</strong> {facilitiesCount}
                </p>
              </div>

              <p className="disease-details-page__placeholder">
                Weekly trend chart will appear here once the chart component is connected.
              </p>
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
              region=""
              district=""
              subCounty=""
              loading={loading}
            />
          </div>
        </Tile>
      </div>
    </Content>
  );
}
