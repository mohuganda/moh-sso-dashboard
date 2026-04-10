import { useEffect, useMemo, useRef } from "react";
import { GeoJSON, MapContainer, Pane, useMap } from "react-leaflet";
import type { GeoJSON as GeoJSONType, Feature, Geometry } from "geojson";
import type { Layer, PathOptions } from "leaflet";
import L from "leaflet";

import type { WeeklyStatus } from "../../../store/types/surveillance.types";

type StatusKey = "maroon" | "red" | "yellow" | "green" | "default";

type UgandaMapFeatureProperties = {
  district?: string;
  district_name?: string;
  name?: string;
  ADM2_NAME?: string;
  ADM1_NAME?: string;
  region?: string;
  subregion?: string;
  district_id?: string;
  region_id?: string;
  sub_county_id?: string;
  [key: string]: unknown;
};

interface SurveillanceUgandaMapProps {
  geoJson: GeoJSONType;
  weeklyStatuses?: WeeklyStatus[];
  selectedWeek?: {
    year?: number;
    week?: number;
  };
  title?: string;
}

const STATUS_COLORS: Record<StatusKey, string> = {
  maroon: "#7d1733",
  red: "#e68b7d",
  yellow: "#e4d64e",
  green: "#a7e2b6",
  default: "#c7dced",
};

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

function mapRiskLevelToStatusKey(status?: WeeklyStatus["status"]): StatusKey {
  switch (status) {
    case "MAROON":
      return "maroon";
    case "RED":
      return "red";
    case "YELLOW":
      return "yellow";
    case "GREEN":
      return "green";
    default:
      return "default";
  }
}

function getStatusPriority(status: StatusKey): number {
  switch (status) {
    case "maroon":
      return 4;
    case "red":
      return 3;
    case "yellow":
      return 2;
    case "green":
      return 1;
    default:
      return 0;
  }
}

function FitMapToGeoJson({ geoJson }: { geoJson: GeoJSONType }) {
  const map = useMap();

  useEffect(() => {
    const layer = L.geoJSON(geoJson as any);
    const bounds = layer.getBounds();

    if (bounds.isValid()) {
      map.fitBounds(bounds, {
        padding: [24, 24],
      });
    }
  }, [geoJson, map]);

  return null;
}

export default function SurveillanceUgandaMap({
  geoJson,
  weeklyStatuses = [],
  selectedWeek,
  title,
}: SurveillanceUgandaMapProps) {
  const mapRef = useRef<L.Map | null>(null);
  const geoJsonRef = useRef<L.GeoJSON | null>(null);

  const summary = useMemo(() => {
    return weeklyStatuses.reduce(
      (acc, item) => {
        const status = mapRiskLevelToStatusKey(item.status);

        if (status === "maroon") acc.maroon += 1;
        if (status === "red") acc.red += 1;
        if (status === "yellow") acc.yellow += 1;
        if (status === "green") acc.green += 1;

        return acc;
      },
      { maroon: 0, red: 0, yellow: 0, green: 0 },
    );
  }, [weeklyStatuses]);

  const districtStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of weeklyStatuses) {
      if (!item.district_id) continue;

      const districtId = String(item.district_id);
      const nextStatus = mapRiskLevelToStatusKey(item.status);
      const currentStatus = map.get(districtId) ?? "default";

      if (getStatusPriority(nextStatus) > getStatusPriority(currentStatus)) {
        map.set(districtId, nextStatus);
      }
    }

    return map;
  }, [weeklyStatuses]);

  const regionStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of weeklyStatuses) {
      if (!item.region_id || item.district_id) continue;

      const regionId = String(item.region_id);
      const nextStatus = mapRiskLevelToStatusKey(item.status);
      const currentStatus = map.get(regionId) ?? "default";

      if (getStatusPriority(nextStatus) > getStatusPriority(currentStatus)) {
        map.set(regionId, nextStatus);
      }
    }

    return map;
  }, [weeklyStatuses]);

  const getFeatureStatus = (feature?: Feature<Geometry, UgandaMapFeatureProperties>): StatusKey => {
    if (!feature?.properties) {
      return "default";
    }

    const districtId = String(feature.properties.district_id ?? "");
    const regionId = String(feature.properties.region_id ?? "");

    if (districtId && districtStatusMap.has(districtId)) {
      return districtStatusMap.get(districtId) ?? "default";
    }

    if (regionId && regionStatusMap.has(regionId)) {
      return regionStatusMap.get(regionId) ?? "default";
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

  const styleFeature = (feature?: Feature<Geometry, UgandaMapFeatureProperties>): PathOptions => {
    const status = getFeatureStatus(feature);

    return {
      fillColor: STATUS_COLORS[status],
      fillOpacity: 1,
      color: "#6f6f6f",
      weight: 0.8,
      opacity: 1,
    };
  };

  const highlightFeature = (layer: Layer) => {
    const vectorLayer = layer as L.Path;

    vectorLayer.setStyle({
      weight: 1.5,
      color: "#1f1f1f",
      fillOpacity: 1,
    });
  };

  const resetHighlight = (layer: Layer) => {
    if (geoJsonRef.current) {
      geoJsonRef.current.resetStyle(layer);
    }
  };

  const handleReset = () => {
    if (!mapRef.current) return;

    const layer = L.geoJSON(geoJson as any);
    const bounds = layer.getBounds();

    if (bounds.isValid()) {
      mapRef.current.fitBounds(bounds, {
        padding: [24, 24],
      });
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
            maxZoom: 9,
          });
        }
      },
    });

    layer.bindTooltip(
      `
        <div class="surveillance-map__tooltip">
          <strong>${label}</strong><br/>
          Status: ${status === "default" ? "No data" : status.toUpperCase()}
        </div>
      `,
      { sticky: true },
    );
  };

  const displayTitle =
    title ||
    (selectedWeek?.year
      ? `Week ${selectedWeek.week ?? ""} ${selectedWeek.year}`.trim()
      : "Uganda surveillance map");

  return (
    <div className="surveillance-map">
      <div className="surveillance-map__header">
        <h3>{displayTitle}</h3>
      </div>

      <div className="surveillance-map__canvas">
        <MapContainer
          ref={mapRef}
          zoomControl={false}
          attributionControl={false}
          dragging={false}
          doubleClickZoom={false}
          boxZoom={false}
          keyboard={false}
          scrollWheelZoom={false}
          className="surveillance-map__leaflet"
        >
          <FitMapToGeoJson geoJson={geoJson} />

          <Pane name="districts" style={{ zIndex: 400 }}>
            <GeoJSON
              ref={geoJsonRef}
              data={geoJson}
              style={styleFeature}
              onEachFeature={onEachFeature}
            />
          </Pane>
        </MapContainer>
      </div>

      <div className="surveillance-map__footer">
        <div className="surveillance-map__legend">
          <span className="surveillance-map__legend-item">
            <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--yellow" />
            Yellow {summary.yellow ? `(${summary.yellow})` : ""}
          </span>

          <span className="surveillance-map__legend-item">
            <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--green" />
            Green {summary.green ? `(${summary.green})` : ""}
          </span>

          <span className="surveillance-map__legend-item">
            <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--red" />
            Red {summary.red ? `(${summary.red})` : ""}
          </span>

          <span className="surveillance-map__legend-item">
            <span className="surveillance-map__legend-swatch surveillance-map__legend-swatch--maroon" />
            Maroon {summary.maroon ? `(${summary.maroon})` : ""}
          </span>
        </div>

        <div className="surveillance-map__actions">
          <button type="button" className="surveillance-map__tool-button" onClick={handleReset}>
            Reset
          </button>
        </div>
      </div>
    </div>
  );
}
