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
import type { DistrictWeeklyStatus } from "../../../../store/types/surveillance.types";

interface DiseaseAlertsTableProps {
  districtStatuses?: DistrictWeeklyStatus[];
  district?: string;
  region?: string;
  diseaseName?: string;
  loading?: boolean;
}

const headers = [
  { key: "district", header: "District" },
  { key: "region", header: "Region" },
  { key: "week", header: "Week" },
  { key: "disease", header: "Disease" },
  { key: "severity", header: "Severity" },
];

function normalize(value?: string) {
  return (value ?? "").trim().toLowerCase();
}

function formatSeverity(level: string) {
  return level.charAt(0).toUpperCase() + level.slice(1);
}

export function DiseaseAlertsTable({
  districtStatuses = [],
  district = "",
  region = "",
  diseaseName = "",
  loading = false,
}: DiseaseAlertsTableProps) {
  const rows = useMemo(() => {
    const normalizedDisease = normalize(diseaseName);

    const expandedRows = districtStatuses.flatMap((item, index) => {
      const alertDiseases = [
        ...(item.maroon ?? []).map((disease) => ({ disease, level: "maroon" })),
        ...(item.red ?? []).map((disease) => ({ disease, level: "red" })),
        ...(item.yellow ?? []).map((disease) => ({ disease, level: "yellow" })),
      ];

      return alertDiseases.map((entry, diseaseIndex) => ({
        id: `${item.id ?? item.district_id ?? index}-${entry.disease}-${diseaseIndex}`,
        district: item.district_name ?? "--",
        region: item.region_name ?? "--",
        week: item.week != null ? String(item.week) : "--",
        disease: entry.disease ?? "--",
        severity: formatSeverity(entry.level),
      }));
    });

    return expandedRows.filter((row) => {
      const matchesDistrict = !district || normalize(row.district) === normalize(district);
      const matchesRegion = !region || normalize(row.region) === normalize(region);
      const matchesDisease =
        !normalizedDisease || normalize(row.disease).includes(normalizedDisease);

      return matchesDistrict && matchesRegion && matchesDisease;
    });
  }, [districtStatuses, district, region, diseaseName]);

  if (loading) {
    return <p>Loading disease alerts...</p>;
  }

  if (!rows.length) {
    return <p>No disease alerts available.</p>;
  }

  return (
    <DataTable rows={rows} headers={headers} size="sm">
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
