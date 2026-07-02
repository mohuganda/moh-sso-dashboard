import { useEffect, useMemo, useRef } from "react";
import { GeoJSON, MapContainer, useMap } from "react-leaflet";
import type { GeoJSON as GeoJSONType, Feature, Geometry } from "geojson";
import type { Layer, PathOptions } from "leaflet";
import L from "leaflet";

import type { WeeklyStatus, WeeklyStatusDetailed } from "../types";

type StatusKey = "maroon" | "red" | "yellow" | "green" | "default";
type MapLevel = "district" | "subcounty";

type UgandaMapFeatureProperties = {
  District?: string;
  Region?: string;
  regions?: string;
  Subcounty?: string;
  sname2019?: string;
  district?: string;
  district_name?: string;
  region?: string;
  subregion?: string;
  subcounty?: string;
  sub_county?: string;
  subcounty_name?: string;
  name?: string;
  ADM1_NAME?: string;
  ADM2_NAME?: string;
  [key: string]: unknown;
};

interface SurveillanceUgandaMapProps {
  geoJson: GeoJSONType;
  weeklyStatuses?: WeeklyStatusDetailed[];
  selectedWeek?: {
    year?: number;
    week?: number;
  };
  title?: string;
  mapLevel?: MapLevel;
  onDistrictSelect?: (districtName: string) => void;
  onSubCountySelect?: (subcountyName: string) => void;
}

const STATUS_COLORS: Record<StatusKey, string> = {
  maroon: "#7d1733",
  red: "#e68b7d",
  yellow: "#e4d64e",
  green: "#a7e2b6",
  default: "#c7dced",
};

function normalize(value: unknown): string {
  if (typeof value !== "string") return "";
  return value.trim().toLowerCase();
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

function getFeatureDistrictName(properties?: UgandaMapFeatureProperties) {
  return (
    properties?.District ||
    properties?.district_name ||
    properties?.district ||
    properties?.ADM2_NAME ||
    properties?.name ||
    ""
  );
}

function getFeatureRegionName(properties?: UgandaMapFeatureProperties) {
  return (
    properties?.Region || properties?.regions || properties?.region || properties?.ADM1_NAME || ""
  );
}

function getFeatureSubCountyName(properties?: UgandaMapFeatureProperties) {
  return (
    properties?.Subcounty ||
    properties?.sname2019 ||
    properties?.subcounty_name ||
    properties?.subcounty ||
    properties?.sub_county ||
    properties?.name ||
    ""
  );
}

function RefreshMapSize() {
  const map = useMap();

  useEffect(() => {
    const t1 = window.setTimeout(() => map.invalidateSize(), 0);
    const t2 = window.setTimeout(() => map.invalidateSize(), 150);

    return () => {
      window.clearTimeout(t1);
      window.clearTimeout(t2);
    };
  }, [map]);

  return null;
}

function FitMapToGeoJson({ geoJson }: { geoJson: GeoJSONType }) {
  const map = useMap();

  useEffect(() => {
    const layer = L.geoJSON(geoJson as any);
    const bounds = layer.getBounds();

    if (bounds.isValid()) {
      map.fitBounds(bounds, { padding: [24, 24] });
      map.invalidateSize();
    }
  }, [geoJson, map]);

  return null;
}

export default function SurveillanceUgandaMap({
  geoJson,
  weeklyStatuses = [],
  selectedWeek,
  title,
  mapLevel = "district",
  onDistrictSelect,
  onSubCountySelect,
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
      const districtName = normalize(item.district_name);
      if (!districtName) continue;
      map.set(districtName, mapRiskLevelToStatusKey(item.status));
    }

    return map;
  }, [weeklyStatuses]);

  const subCountyStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of weeklyStatuses) {
      const subCountyName = normalize(item.sub_county_name);
      if (!subCountyName) continue;
      map.set(subCountyName, mapRiskLevelToStatusKey(item.status));
    }

    return map;
  }, [weeklyStatuses]);

  const regionStatusMap = useMemo(() => {
    const map = new Map<string, StatusKey>();

    for (const item of weeklyStatuses) {
      const regionName = normalize(item.region_name);
      if (!regionName) continue;
      map.set(regionName, mapRiskLevelToStatusKey(item.status));
    }

    return map;
  }, [weeklyStatuses]);

  const getFeatureStatus = (feature?: Feature<Geometry, UgandaMapFeatureProperties>): StatusKey => {
    if (!feature?.properties) return "default";

    const districtName = normalize(getFeatureDistrictName(feature.properties));
    const regionName = normalize(getFeatureRegionName(feature.properties));
    const subCountyName = normalize(getFeatureSubCountyName(feature.properties));

    if (mapLevel === "subcounty") {
      if (subCountyName && subCountyStatusMap.has(subCountyName)) {
        return subCountyStatusMap.get(subCountyName) ?? "default";
      }

      if (districtName && districtStatusMap.has(districtName)) {
        return districtStatusMap.get(districtName) ?? "default";
      }

      return "default";
    }

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

    if (mapLevel === "subcounty") {
      return (
        getFeatureSubCountyName(feature.properties) ||
        getFeatureDistrictName(feature.properties) ||
        "Unknown area"
      );
    }

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
    (layer as L.Path).setStyle({
      weight: 1.5,
      color: "#1f1f1f",
      fillOpacity: 1,
    });
  };

  const resetHighlight = (layer: Layer) => {
    geoJsonRef.current?.resetStyle(layer);
  };

  const handleReset = () => {
    if (!mapRef.current) return;

    const layer = L.geoJSON(geoJson as any);
    const bounds = layer.getBounds();

    if (bounds.isValid()) {
      mapRef.current.fitBounds(bounds, { padding: [24, 24] });
      mapRef.current.invalidateSize();
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
            maxZoom: mapLevel === "subcounty" ? 10 : 9,
          });
        }

        if (mapLevel === "subcounty") {
          const subcountyName = getFeatureSubCountyName(feature.properties);
          if (subcountyName) onSubCountySelect?.(subcountyName);
          return;
        }

        const districtName = getFeatureDistrictName(feature.properties);
        if (districtName) onDistrictSelect?.(districtName);
      },
    });

    layer.bindTooltip(
      `
        <div class="surveillance-map__tooltip">
          <strong>${label}</strong><br/>
          Status: ${status === "default" ? "No data" : status.toUpperCase()}
        </div>
      `,
      { sticky: true, direction: "top", opacity: 1, className: "surveillance-map__tooltip" },
    );
  };

  const displayTitle =
    title ||
    (selectedWeek?.year
      ? `Week ${selectedWeek.week ?? ""} ${selectedWeek.year}`.trim()
      : "Uganda surveillance map");

  const geoJsonKey = `${mapLevel}-${(geoJson as any)?.features?.length ?? 0}`;

  return (
    <div className="surveillance-map">
      <div className="surveillance-map__header">
        <h3>{displayTitle}</h3>
      </div>

      <div className="surveillance-map__canvas">
        <MapContainer
          ref={mapRef}
          center={[1.3733, 32.2903]}
          zoom={7}
          zoomControl={false}
          attributionControl={false}
          dragging={false}
          doubleClickZoom={false}
          boxZoom={false}
          keyboard={false}
          scrollWheelZoom={false}
          className="surveillance-map__leaflet"
        >
          <RefreshMapSize />
          <FitMapToGeoJson geoJson={geoJson} />

          <GeoJSON
            key={geoJsonKey}
            ref={geoJsonRef}
            data={geoJson}
            style={styleFeature}
            onEachFeature={onEachFeature}
          />
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
