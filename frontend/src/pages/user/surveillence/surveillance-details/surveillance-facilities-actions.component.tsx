import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from "@carbon/react";
import { useMemo } from "react";
import type { FacilityWeeklyMetric } from "../../../../store/types/surveillance.types";

interface FacilitiesActionTableProps {
  facilityMetrics?: FacilityWeeklyMetric[];
  regionId?: string;
  districtId?: string;
  subCountyId?: string;
  loading?: boolean;
  regionNameById?: Record<string, string>;
  districtNameById?: Record<string, string>;
  subCountyNameById?: Record<string, string>;
  diseaseNameById?: Record<string, string>;
  indicatorNameById?: Record<string, string>;
  epiWeekLabelById?: Record<string, string>;
}

const facilityHeaders = [
  { key: "facility", header: "Facility" },
  { key: "region", header: "Region" },
  { key: "district", header: "District" },
  { key: "subcounty", header: "Subcounty" },
  { key: "disease", header: "Disease" },
  { key: "week", header: "Week" },
  { key: "cases", header: "Cases" },
];

export function FacilitiesActionTable({
  facilityMetrics = [],
  regionId = "",
  districtId = "",
  subCountyId = "",
  loading = false,
  regionNameById = {},
  districtNameById = {},
  subCountyNameById = {},
  diseaseNameById = {},
  indicatorNameById = {},
  epiWeekLabelById = {},
}: FacilitiesActionTableProps) {
  const rows = useMemo(() => {
    return facilityMetrics
      .filter((item) => {
        const matchesRegion = !regionId || String(item.region_id ?? "") === regionId;

        const matchesDistrict = !districtId || String(item.district_id ?? "") === districtId;

        const matchesSubCounty = !subCountyId || String(item.subcounty_id ?? "") === subCountyId;

        return matchesRegion && matchesDistrict && matchesSubCounty;
      })
      .map((item, index) => {
        const numericCases = Number(item.metric_value ?? 0);

        const region = item.region_id
          ? (regionNameById[String(item.region_id)] ?? item.region_name ?? "--")
          : (item.region_name ?? "--");

        const district = item.district_id
          ? (districtNameById[String(item.district_id)] ?? item.district_name ?? "--")
          : (item.district_name ?? "--");

        const subcounty =
          "sub_county_id" in item && item.sub_county_id
            ? (subCountyNameById[String(item.sub_county_id)] ?? item.subcounty_name ?? "--")
            : (item.subcounty_name ?? "--");

        const disease = item.disease_id
          ? (diseaseNameById[String(item.disease_id)] ??
            item.disease_name ??
            item.indicator_name ??
            "--")
          : item.indicator_id
            ? (indicatorNameById[String(item.indicator_id)] ??
              item.indicator_name ??
              item.disease_name ??
              "--")
            : (item.disease_name ?? item.indicator_name ?? "--");

        const week = item.epi_week_id
          ? (epiWeekLabelById[String(item.epi_week_id)] ?? item.week?.toString() ?? "--")
          : (item.week?.toString() ?? "--");

        return {
          id: item.id ?? item.facility_id ?? `${item.facility_name ?? "facility"}-${index}`,
          facility: item.facility_name ?? "--",
          region,
          district,
          subcounty,
          disease,
          week,
          cases: String(numericCases),
          numericCases,
        };
      })
      .sort((a, b) => b.numericCases - a.numericCases);
  }, [
    facilityMetrics,
    regionId,
    districtId,
    subCountyId,
    regionNameById,
    districtNameById,
    subCountyNameById,
    diseaseNameById,
    indicatorNameById,
    epiWeekLabelById,
  ]);

  if (loading) {
    return <p>Loading facility action data...</p>;
  }

  if (!rows.length) {
    return <p>No facility action data available.</p>;
  }

  return (
    <DataTable
      rows={rows.map(({ numericCases, ...row }) => row)}
      headers={facilityHeaders}
      size="sm"
    >
      {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
        <Table {...getTableProps()} size="sm">
          <TableHead>
            <TableRow>
              {headers.map((header) => (
                <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
              ))}
            </TableRow>
          </TableHead>

          <TableBody>
            {rows.map((row) => (
              <TableRow {...getRowProps({ row })}>
                {row.cells.map((cell) => (
                  <TableCell key={cell.id}>{cell.value}</TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </DataTable>
  );
}
