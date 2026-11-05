# Surveillance map boundaries

`uganda_subcounties.geojson` is processed Uganda Bureau of Statistics (UBOS)
National Population and Housing Census 2024 data, retrieved on 2026-11-05.
Source: https://statistics.ubos.org/nphc/map

The official map uses the public endpoint
`https://statistics.ubos.org/nphc/api/get_geospatial_data.php`.
The district query supplies district codes/names; each
`?level=subcounty&district_code=<code>` supplies that district's sub-county
geometry. The nationwide `level=subcounty` query without a district code returns
county-level shapes and must not be used as a sub-county layer.

The bundled snapshot contains 2,206 Polygon/MultiPolygon features covering 146
unique district/city codes. Hoima City (438) includes Hoima East Division (438101)
and Hoima West Division (438201). Geometry is unchanged from the source. Processing
removes population fields and normalizes label properties to `District` and
`Subcounty`, retaining `DCode`, `CCode`, and `SCode`. This is a processed dataset,
not an official UBOS product; source geometry may contain overlaps or gaps.

Snapshot SHA-256:
`e35141ba72d506d1cb3ef8fed8c9356bb7b1aec23c43501ca09fc3ddabfaa429`

To refresh, run `python3 infra/scripts/update-surveillance-boundaries.py` from the
repository root. It caches source responses under `/tmp/moh-ubos-boundaries-2024`,
validates administrative labels, polygon rings, finite WGS84 coordinates, and the
expected Uganda extent, and replaces the bundled file only after every district
succeeds. Use an empty cache for a fresh snapshot. Update this provenance note,
snapshot hash, and the sub-county version parameter in `geojson.api.ts` when
publishing a new snapshot. The version parameter bypasses previously cached
legacy boundary responses.

The frontend shows every geographic sub-county in the selected district even if
reporting references use an aggregated council name. It does not create database
locations or allocate district/council reports to geographic divisions. Only
sub-county status records with matching labels colour the sub-county layer;
unmatched shapes remain in the no-status-data colour. District flags remain in
the tiles. Matching recognizes the explicit Luwero/Luweero and Sembabule/Ssembabule
spelling variants between the old district layer and the 2024 source; displayed
source labels are preserved. The map links to UBOS for attribution.

`uganda_districts.geojson` and `uga_water.geojson` remain the existing layers;
they are not replaced by this refresh.
