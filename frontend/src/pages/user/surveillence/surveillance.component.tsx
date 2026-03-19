import { Breadcrumb, BreadcrumbItem } from "@carbon/react";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import SurveillanceFilters from "./surveillance-filters.component";
import SurveillanceTiles from "./surveillance-tiles.component";
import SurveillancePanels from "./surveillance-panel.component";
import "./surveillance.css";

const epiWeekOptions = [
  { value: "", label: "Select..." },
  { value: "11", label: "11" },
  { value: "10", label: "10" },
  { value: "9", label: "9" },
];

const regionOptions = [
  { value: "", label: "Select..." },
  { value: "central", label: "Central" },
  { value: "eastern", label: "Eastern" },
  { value: "northern", label: "Northern" },
  { value: "western", label: "Western" },
];

const districtOptions = [{ value: "", label: "Select..." }];
const subCountyOptions = [{ value: "", label: "Select..." }];

export default function SurveillanceDashboardPage() {
  const [epiWeek, setEpiWeek] = useState("11");
  const [region, setRegion] = useState("");
  const [district, setDistrict] = useState("");
  const [subCounty, setSubCounty] = useState("");

  const currentEpiWeek = useMemo(() => epiWeek || "11", [epiWeek]);
  const navigate = useNavigate();

  const handleOpenDisease = (diseaseName: string) => {
    navigate(`/surveillance/${diseaseName}`);
  };

  const immediateActionItems = [
    {
      label: "Acute Flaccid Paralysis",
      onClick: () => handleOpenDisease("acute-flaccid-paralysis"),
    },
    {
      label: "Anthrax",
      onClick: () => handleOpenDisease("anthrax"),
    },
    {
      label: "Plague",
      onClick: () => handleOpenDisease("plague"),
    },
    {
      label: "Yellow Fever",
      onClick: () => handleOpenDisease("yellow-fever"),
    },
  ];

  const takeActionItems = [
    {
      label: "Malaria",
      onClick: () => handleOpenDisease("malaria"),
    },
    {
      label: "Maternal Death",
      onClick: () => handleOpenDisease("maternal-death"),
    },
    {
      label: "Measles",
      onClick: () => handleOpenDisease("measles"),
    },
  ];

  return (
    <div className="surveillance-dashboard-page">
      <div className="surveillance-dashboard-page__header">
        <Breadcrumb className="surveillance-dashboard-page__breadcrumb" noTrailingSlash>
          <BreadcrumbItem>
            <a href="#">Data &amp; Statistics</a>
          </BreadcrumbItem>

          <BreadcrumbItem isCurrentPage>
            <span>National Surveillance Reporting Dashboard: Epi-Week: {currentEpiWeek}</span>
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
        </div>
      </div>

      <section className="surveillance-dashboard-page__section">
        <SurveillanceFilters
          epiWeek={epiWeek}
          region={region}
          district={district}
          subCounty={subCounty}
          epiWeekOptions={epiWeekOptions}
          regionOptions={regionOptions}
          districtOptions={districtOptions}
          subCountyOptions={subCountyOptions}
          onEpiWeekChange={setEpiWeek}
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
            items: [],
          }}
          watch={{
            title: "Watch",
            items: [],
          }}
        />
      </section>

      <section className="surveillance-dashboard-page__section">
        <SurveillancePanels />
      </section>
    </div>
  );
}
