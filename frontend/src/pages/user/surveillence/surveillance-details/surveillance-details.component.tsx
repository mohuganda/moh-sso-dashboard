import { Breadcrumb, BreadcrumbItem, Content, Grid, Column, Tile, Link } from "@carbon/react";
import { useParams, Link as RouterLink } from "react-router-dom";
import "../surveillance-details/surveillance-details.css";
import { DiseaseAlertsTable } from "./surveillance-disease-alerts.component";
import { FacilitiesActionTable } from "./surveillance-facilities-actions.component";

function formatDiseaseName(value?: string) {
  if (!value) return "Disease";
  return value
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export default function DiseaseDetailsPage() {
  const { diseaseName } = useParams();
  const title = formatDiseaseName(diseaseName);

  return (
    <Content className="disease-details-page">
      <Breadcrumb noTrailingSlash>
        <BreadcrumbItem>
          <RouterLink to="/surveillance">Surveillance</RouterLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>{title}</BreadcrumbItem>
      </Breadcrumb>

      <div className="disease-details-page__header">
        <h1>Details for {title}</h1>
        <p>
          View disease-specific alerts, weekly case trends, relevant documents, and facilities that
          require immediate follow-up action.
        </p>
      </div>

      <Grid fullWidth className="disease-details-page__grid">
        <Column lg={8} md={4} sm={4}>
          <Tile className="disease-details-page__tile">
            <div className="disease-details-page__section">
              <h3>Relevant Documents</h3>
              <ul className="disease-details-page__links">
                <li>
                  <Link href="#">Outbreak Guidelines</Link>
                </li>
                <li>
                  <Link href="#">Data Collection Tools</Link>
                </li>
              </ul>
            </div>

            <div className="disease-details-page__section">
              <h3>Disease Alerts</h3>
              <DiseaseAlertsTable />
            </div>
          </Tile>
        </Column>

        <Column lg={8} md={4} sm={4}>
          <Tile className="disease-details-page__tile">
            <div className="disease-details-page__section">
              <h3>{title} Weekly Cases</h3>
            </div>
          </Tile>
        </Column>
      </Grid>

      <div className="disease-details-page__bottom">
        <Tile className="disease-details-page__tile">
          <div className="disease-details-page__section">
            <h3>Facilities requiring Action</h3>
            <FacilitiesActionTable />
          </div>
        </Tile>
      </div>
    </Content>
  );
}
