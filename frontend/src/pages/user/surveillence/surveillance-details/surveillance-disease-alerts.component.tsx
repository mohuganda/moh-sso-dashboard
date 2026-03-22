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
  { key: "createdAt", header: "Created At" },
  { key: "district", header: "District" },
  { key: "week", header: "Week" },
  { key: "disease", header: "Disease" },
  { key: "region", header: "Region" },
];

function normalize(value?: string) {
  return (value ?? "").trim().toLowerCase();
}

function formatDate(value?: string) {
  if (!value) return "--";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toISOString().slice(0, 10);
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
        // createdAt: formatDate(item.created_at),
        district: item.district_name ?? "--",
        // week: item.week != null ? String(item.week) : "--",
        disease: entry.disease,
        // region: item.region_name ?? "--",
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
