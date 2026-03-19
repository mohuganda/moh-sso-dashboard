type SurveillanceRankItem = {
  label: string;
  value: number | string;
};

interface SurveillancePanelsProps {
  facilitiesReporting?: SurveillanceRankItem[];
  eidsrAlerts?: SurveillanceRankItem[];
  mapContent?: React.ReactNode;
}

interface RankingPanelProps {
  title: string;
  items: SurveillanceRankItem[];
}

function RankingPanel({ title, items }: RankingPanelProps) {
  return (
    <section className="surveillance-panel surveillance-panel--side">
      <h3 className="surveillance-panel__title">{title}</h3>

      {items.length > 0 ? (
        <ol className="surveillance-panel__ranking-list">
          {items.map((item, index) => (
            <li
              key={`${title}-${item.label}-${index}`}
              className="surveillance-panel__ranking-item"
            >
              <span className="surveillance-panel__ranking-label">
                {item.label}: {item.value}
              </span>
            </li>
          ))}
        </ol>
      ) : (
        <p className="surveillance-panel__empty">No data available.</p>
      )}
    </section>
  );
}

function DefaultMapPlaceholder() {
  return (
    <div className="surveillance-map">
      <div className="surveillance-map__toolbar">
        <button type="button" className="surveillance-map__tool-button">
          Reset
        </button>
      </div>

      <div className="surveillance-map__canvas">
        <div className="surveillance-map__placeholder-shape">Uganda Map</div>
      </div>

      <div className="surveillance-map__legend">
        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--alert" />
          Yellow: Alert
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--watch" />
          Green: Watch
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--take-action" />
          Orange: Take Action
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--immediate-action" />
          Maroon: Immediate Action
        </span>
      </div>
    </div>
  );
}

const defaultFacilitiesReporting: SurveillanceRankItem[] = [
  { label: "Kampala", value: 930 },
  { label: "Wakiso", value: 285 },
  { label: "Kasese", value: 141 },
  { label: "Luwero", value: 111 },
  { label: "Mukono", value: 97 },
  { label: "Rukungiri", value: 94 },
  { label: "Kyotera", value: 83 },
  { label: "Tororo", value: 77 },
  { label: "Isingiro", value: 74 },
  { label: "Mityana", value: 71 },
];

const defaultEidsrAlerts: SurveillanceRankItem[] = [
  { label: "Tororo", value: 53 },
  { label: "Mubende", value: 50 },
  { label: "Namisindwa", value: 48 },
  { label: "Kibuku", value: 40 },
  { label: "Kasese", value: 38 },
  { label: "Mbale", value: 30 },
  { label: "Lira", value: 27 },
  { label: "Moroto", value: 22 },
  { label: "Manafwa", value: 20 },
  { label: "Amudat", value: 19 },
];

export default function SurveillancePanels({
  facilitiesReporting = defaultFacilitiesReporting,
  eidsrAlerts = defaultEidsrAlerts,
  mapContent,
}: SurveillancePanelsProps) {
  return (
    <div className="surveillance-panels">
      <div className="surveillance-panels__grid">
        <RankingPanel title="Facilities Reporting" items={facilitiesReporting} />

        <section className="surveillance-panel surveillance-panel--map">
          {mapContent ?? <DefaultMapPlaceholder />}
        </section>

        <RankingPanel title="EIDSR Alerts" items={eidsrAlerts} />
      </div>
    </div>
  );
}
