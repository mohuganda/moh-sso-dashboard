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
  region?: string;
  district?: string;
  subCounty?: string;
  loading?: boolean;
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

function normalize(value?: string) {
  return (value ?? "").trim().toLowerCase();
}

export function FacilitiesActionTable({
  facilityMetrics = [],
  region = "",
  district = "",
  subCounty = "",
  loading = false,
}: FacilitiesActionTableProps) {
  const rows = useMemo(() => {
    return facilityMetrics
      .filter((item) => {
        const matchesRegion = !region || normalize(item.region_name) === normalize(region);
        const matchesDistrict = !district || normalize(item.district_name) === normalize(district);
        const matchesSubCounty =
          !subCounty || normalize(item.subcounty_name) === normalize(subCounty);

        return matchesRegion && matchesDistrict && matchesSubCounty;
      })
      .map((item, index) => {
        const numericCases = Number(item.value ?? 0);

        return {
          id: item.id ?? item.facility_id ?? `${item.facility_name ?? "facility"}-${index}`,
          facility: item.facility_name ?? "--",
          region: item.region_name ?? "--",
          district: item.district_name ?? "--",
          subcounty: item.subcounty_name ?? "--",
          disease: item.disease_name ?? item.indicator_name ?? "--",
          week: item.week ? String(item.week) : "--",
          cases: String(numericCases),
          numericCases,
        };
      })
      .sort((a, b) => b.numericCases - a.numericCases);
  }, [facilityMetrics, region, district, subCounty]);

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
