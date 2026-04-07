import { useMemo, useRef } from "react";
import { GeoJSON, MapContainer, TileLayer } from "react-leaflet";
import type { GeoJSON as GeoJSONType, Feature, Geometry } from "geojson";
import type { Layer } from "leaflet";
import L from "leaflet";

import type {
  DistrictWeeklyStatus,
  RegionWeeklyStatus,
  NationalWeeklyStatus,
} from "../../../store/types/surveillance.types";

type StatusKey = "maroon" | "red" | "yellow" | "green" | "default";

type UgandaMapFeatureProperties = {
  district?: string;
  district_name?: string;
  name?: string;
  ADM2_NAME?: string;
  ADM1_NAME?: string;
  region?: string;
  subregion?: string;
  [key: string]: unknown;
};

interface SurveillanceUgandaMapProps {
  geoJson: GeoJSONType;
  districtStatuses?: DistrictWeeklyStatus[];
  regionStatuses?: RegionWeeklyStatus[];
  nationalStatuses?: NationalWeeklyStatus[];
  selectedWeek?: {
    year?: number;
    week?: number;
  };
}

const UGANDA_CENTER: [number, number] = [1.3733, 32.2903];
const DEFAULT_ZOOM = 7;

const STATUS_COLORS: Record<StatusKey, string> = {
  maroon: "#8b0000",
  red: "#ff5a36",
  yellow: "#e2b600",
  green: "#008a00",
  default: "#d9d9d9",
};

function normalize(value?: string) {
  return (value ?? "").trim().toLowerCase();
}

function getFeatureDistrictName(properties?: UgandaMapFeatureProperties) {
  return (
    properties?.district_name ||
    properties?.district ||
    properties?.ADM2_NAME ||
    properties?.name ||
    ""
  );
}

function getFeatureRegionName(properties?: UgandaMapFeatureProperties) {
  return properties?.region || properties?.ADM1_NAME || properties?.subregion || "";
}

function getStatusFromDistrictItem(item: DistrictWeeklyStatus): StatusKey {
  const maroonCount = item.maroon?.length ?? 0;
  const redCount = item.red?.length ?? 0;
  const yellowCount = item.yellow?.length ?? 0;
  const greenCount = item.green?.length ?? 0;

  if (maroonCount > 0) return "maroon";
  if (redCount > 0) return "red";
  if (yellowCount > 0) return "yellow";
  if (greenCount > 0) return "green";

  return "default";
}

function getStatusFromRegionItem(item: RegionWeeklyStatus): StatusKey {
  const maroonCount = item.maroon?.length ?? 0;
  const redCount = item.red?.length ?? 0;
  const yellowCount = item.yellow?.length ?? 0;
  const greenCount = item.green?.length ?? 0;

  if (maroonCount > 0) return "maroon";
  if (redCount > 0) return "red";
  if (yellowCount > 0) return "yellow";
  if (greenCount > 0) return "green";

  return "default";
}

export default function SurveillanceUgandaMap({
  geoJson,
  districtStatuses = [],
  regionStatuses = [],
  nationalStatuses = [],
  selectedWeek,
}: SurveillanceUgandaMapProps) {
  const mapRef = useRef<L.Map | null>(null);
  const geoJsonRef = useRef<L.GeoJSON | null>(null);

  const summary = nationalStatuses[0];

  const districtStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of districtStatuses) {
      const districtName = normalize(item.district_name);
      if (!districtName) continue;

      map.set(districtName, getStatusFromDistrictItem(item));
    }

    return map;
  }, [districtStatuses]);

  const regionStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of regionStatuses) {
      const regionName = normalize(item.region_name);
      if (!regionName) continue;

      map.set(regionName, getStatusFromRegionItem(item));
    }

    return map;
  }, [regionStatuses]);

  const getFeatureStatus = (feature?: Feature<Geometry, UgandaMapFeatureProperties>): StatusKey => {
    if (!feature?.properties) {
      return "default";
    }

    const districtName = normalize(getFeatureDistrictName(feature.properties));
    const regionName = normalize(getFeatureRegionName(feature.properties));

    if (districtName && districtStatusMap.has(districtName)) {
      return districtStatusMap.get(districtName) ?? "default";
    }

    if (regionName && regionStatusMap.has(regionName)) {
      return regionStatusMap.get(regionName) ?? "default";
    }

    return "default";
  };

  const getFeatureLabel = (feature?: Feature<Geometry, UgandaMapFeatureProperties>) => {
    if (!feature?.properties) return "Unknown area";

    return (
      getFeatureDistrictName(feature.properties) ||
      getFeatureRegionName(feature.properties) ||
      "Unknown area"
    );
  };

  const styleFeature = (feature?: Feature<Geometry, UgandaMapFeatureProperties>) => {
    const status = getFeatureStatus(feature);
    const fillColor = STATUS_COLORS[status];

    return {
      fillColor,
      weight: 1,
      opacity: 1,
      color: "#ffffff",
      dashArray: "0",
      fillOpacity: 0.85,
    };
  };

  const highlightFeature = (layer: Layer) => {
    const vectorLayer = layer as L.Path;

    vectorLayer.setStyle({
      weight: 2,
      color: "#161616",
      fillOpacity: 1,
    });

    if (!L.Browser.ie && !L.Browser.opera && !L.Browser.edge) {
      vectorLayer.bringToFront();
    }
  };

  const resetHighlight = (layer: Layer) => {
    if (geoJsonRef.current) {
      geoJsonRef.current.resetStyle(layer);
    }
  };

  const handleReset = () => {
    if (mapRef.current) {
      mapRef.current.setView(UGANDA_CENTER, DEFAULT_ZOOM);
    }
  };

  const onEachFeature = (feature: Feature<Geometry, UgandaMapFeatureProperties>, layer: Layer) => {
    const label = getFeatureLabel(feature);
    const status = getFeatureStatus(feature);

    layer.on({
      mouseover: () => highlightFeature(layer),
      mouseout: () => resetHighlight(layer),
      click: () => {
        const boundsLayer = layer as L.FeatureGroup;
        if (mapRef.current && "getBounds" in boundsLayer) {
          mapRef.current.fitBounds(boundsLayer.getBounds(), {
            padding: [20, 20],
          });
        }
      },
    });

    layer.bindTooltip(
      `
        <div class="surveillance-map__tooltip">
          <strong>${label}</strong><br/>
          Status: ${status === "default" ? "No data" : status}
        </div>
      `,
      {
        sticky: true,
      },
    );
  };

  return (
    <>
      <div className="surveillance-map__toolbar">
        <button type="button" className="surveillance-map__tool-button" onClick={handleReset}>
          Reset
        </button>
      </div>

      <div className="surveillance-map__canvas">
        <MapContainer
          center={UGANDA_CENTER}
          zoom={DEFAULT_ZOOM}
          scrollWheelZoom={true}
          className="surveillance-map__leaflet"
          ref={mapRef}
        >
          <TileLayer
            attribution="&copy; OpenStreetMap contributors"
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />

          <GeoJSON
            data={geoJson}
            style={styleFeature}
            onEachFeature={onEachFeature}
            ref={geoJsonRef}
          />
        </MapContainer>
      </div>

      <div className="surveillance-map__legend">
        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--alert" />
          Yellow: Alert {summary?.yellow ? `(${summary.yellow})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--watch" />
          Green: Watch {summary?.green ? `(${summary.green})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--take-action" />
          Red: Take Action {summary?.red ? `(${summary.red})` : ""}
        </span>

        <span className="surveillance-map__legend-item">
          <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--immediate-action" />
          Maroon: Immediate Action {summary?.maroon ? `(${summary.maroon})` : ""}
        </span>
      </div>

      <div className="surveillance-map__meta">
        <p>
          {selectedWeek?.year && selectedWeek?.week
            ? `Week ${selectedWeek.week}, ${selectedWeek.year}`
            : "Current reporting week"}
        </p>
      </div>
    </>
  );
}
