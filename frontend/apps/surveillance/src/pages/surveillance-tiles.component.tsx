type SurveillanceTileItem = {
  label: string;
  href?: string;
  onClick?: () => void;
};

type SurveillanceTile = {
  title: string;
  items: SurveillanceTileItem[];
};

interface SurveillanceTilesProps {
  immediateAction?: SurveillanceTile;
  takeAction?: SurveillanceTile;
  alert?: SurveillanceTile;
  watch?: SurveillanceTile;
}

function TileList({
  title,
  items,
  tone,
}: {
  title: string;
  items: SurveillanceTileItem[];
  tone: "immediate-action" | "take-action" | "alert" | "watch";
}) {
  return (
    <section className={`surveillance-tile surveillance-tile--${tone}`}>
      <h3 className="surveillance-tile__title">{title}</h3>

      {items.length > 0 ? (
        <ul className="surveillance-tile__list">
          {items.map((item, index) => (
            <li key={`${title}-${item.label}-${index}`} className="surveillance-tile__list-item">
              {item.onClick ? (
                <button type="button" className="surveillance-tile__button" onClick={item.onClick}>
                  {item.label}
                </button>
              ) : item.href ? (
                <a href={item.href} className="surveillance-tile__link">
                  {item.label}
                </a>
              ) : (
                <span className="surveillance-tile__text">{item.label}</span>
              )}
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}

const defaultImmediateAction: SurveillanceTile = {
  title: "Immediate Action",
  items: [],
};

const defaultTakeAction: SurveillanceTile = {
  title: "Take Action",
  items: [],
};

const defaultAlert: SurveillanceTile = {
  title: "Alert",
  items: [],
};

const defaultWatch: SurveillanceTile = {
  title: "Watch",
  items: [],
};

export default function SurveillanceTiles({
  immediateAction = defaultImmediateAction,
  takeAction = defaultTakeAction,
  alert = defaultAlert,
  watch = defaultWatch,
}: SurveillanceTilesProps) {
  return (
    <div className="surveillance-tiles">
      <div className="surveillance-tiles__grid">
        <TileList
          title={immediateAction.title}
          items={immediateAction.items}
          tone="immediate-action"
        />
        <TileList title={takeAction.title} items={takeAction.items} tone="take-action" />
        <TileList title={alert.title} items={alert.items} tone="alert" />
        <TileList title={watch.title} items={watch.items} tone="watch" />
      </div>
    </div>
  );
}
