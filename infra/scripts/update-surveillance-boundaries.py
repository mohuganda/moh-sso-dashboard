#!/usr/bin/env python3
"""Bundle official UBOS NPHC 2024 sub-county geometry for offline maps."""
import concurrent.futures
import hashlib
import json
import math
from pathlib import Path
import time
import urllib.parse
import urllib.request

BASE = "https://statistics.ubos.org/nphc/api/get_geospatial_data.php"
ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "backend/assets/geojson/uganda_subcounties.geojson"
CACHE = Path("/tmp/moh-ubos-boundaries-2024")


def load(params):
    url = BASE + "?" + urllib.parse.urlencode(params)
    cache = CACHE / (hashlib.sha256(url.encode()).hexdigest() + ".json")
    if cache.exists():
        raw = cache.read_text()
    else:
        for attempt in range(3):
            try:
                with urllib.request.urlopen(url, timeout=60) as response:
                    raw = response.read().decode("utf-8")
                break
            except Exception:
                if attempt == 2:
                    raise
                time.sleep(2)
        cache.write_text(raw)
    # Some public API responses prefix valid JSON with PHP warnings.
    start = raw.find('{"geojson"')
    data = json.loads(raw[start:] if start >= 0 else raw)
    geo = data.get("geojson")
    if isinstance(geo, str):
        geo = json.loads(geo)
    if not isinstance(geo, dict) or geo.get("type") != "FeatureCollection":
        raise ValueError(f"Expected GeoJSON FeatureCollection from {url}")
    return geo["features"]


def validate_geometry(geometry):
    if geometry.get("type") not in ("Polygon", "MultiPolygon"):
        raise ValueError("Expected polygon geometry")
    polygons = geometry["coordinates"] if geometry["type"] == "MultiPolygon" else [geometry["coordinates"]]
    if not polygons:
        raise ValueError("Empty geometry")
    for polygon in polygons:
        if not polygon:
            raise ValueError("Empty polygon")
        for ring in polygon:
            if len(ring) < 4 or ring[0] != ring[-1]:
                raise ValueError("Invalid polygon ring")
            for point in ring:
                if len(point) < 2 or not all(math.isfinite(value) for value in point[:2]):
                    raise ValueError("Non-finite coordinate")
                if not (28 <= point[0] <= 36 and -2 <= point[1] <= 5):
                    raise ValueError("Coordinate outside Uganda's expected extent")


def main():
    CACHE.mkdir(exist_ok=True)
    districts = {}
    for feature in load({"level": "district"}):
        props = feature["properties"]
        code, name = str(props["DCode"]), props["District"]
        if code in districts and districts[code] != name:
            raise ValueError(f"Conflicting district name for {code}")
        districts[code] = name
    print(f"Downloading sub-county geometry for {len(districts)} districts/cities", flush=True)

    def district_features(item):
        code, district = item
        features = load({"level": "subcounty", "district_code": code})
        if not features:
            raise ValueError(f"No sub-county geometry for {district}")
        result = []
        for feature in features:
            props = feature["properties"]
            if str(props.get("DCode")) != code or not props.get("Sub_County"):
                raise ValueError(f"Unexpected boundary level or district for {district}")
            validate_geometry(feature["geometry"])
            result.append({
                "type": "Feature",
                "properties": {
                    "District": district,
                    "Subcounty": props["Sub_County"],
                    "DCode": code,
                    "CCode": str(props["CCode"]),
                    "SCode": str(props["SCode"]),
                    "source": "UBOS NPHC 2024 (processed)",
                },
                "geometry": feature["geometry"],
            })
        return result

    all_features = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
        for index, features in enumerate(pool.map(district_features, sorted(districts.items())), 1):
            all_features.extend(features)
            if index % 20 == 0:
                print(f"Validated {index}/{len(districts)} districts", flush=True)
    collection = {
        "type": "FeatureCollection",
        "source": "Processed Uganda Bureau of Statistics NPHC 2024 data; district/subcounty labels normalized, geometry unchanged.",
        "source_url": "https://statistics.ubos.org/nphc/map",
        "features": all_features,
    }
    payload = json.dumps(collection, separators=(",", ":"), ensure_ascii=False) + "\n"
    temporary = OUTPUT.with_suffix(".geojson.tmp")
    temporary.write_text(payload)
    temporary.replace(OUTPUT)
    print(f"Bundled {len(all_features)} features; sha256 {hashlib.sha256(payload.encode()).hexdigest()}", flush=True)


if __name__ == "__main__":
    main()
