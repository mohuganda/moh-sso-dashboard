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
import type { WeeklyStatus } from "../../../../store/types/surveillance.types";

interface DiseaseAlertsTableProps {
  weeklyStatuses?: WeeklyStatus[];
  districtId?: string;
  regionId?: string;
  diseaseName?: string;
  loading?: boolean;
  districtNameById?: Record<string, string>;
  regionNameById?: Record<string, string>;
  diseaseNameById?: Record<string, string>;
  epiWeekLabelById?: Record<string, string>;
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
  return level.charAt(0).toUpperCase() + level.slice(1).toLowerCase();
}

export function DiseaseAlertsTable({
  weeklyStatuses = [],
  districtId = "",
  regionId = "",
  diseaseName = "",
  loading = false,
  districtNameById = {},
  regionNameById = {},
  diseaseNameById = {},
  epiWeekLabelById = {},
}: DiseaseAlertsTableProps) {
  const rows = useMemo(() => {
    const normalizedDisease = normalize(diseaseName);

    return weeklyStatuses
      .filter((item) => item.status !== "GREEN")
      .filter((item) => {
        const matchesDistrict = !districtId || item.district_id === districtId;
        const matchesRegion = !regionId || item.region_id === regionId;

        const resolvedDiseaseName = item.disease_id
          ? (diseaseNameById[String(item.disease_id)] ?? "")
          : "";

        const matchesDisease =
          !normalizedDisease || normalize(resolvedDiseaseName).includes(normalizedDisease);

        return matchesDistrict && matchesRegion && matchesDisease;
      })
      .map((item, index) => {
        const district = item.district_id
          ? (districtNameById[String(item.district_id)] ?? "--")
          : "--";

        const region = item.region_id ? (regionNameById[String(item.region_id)] ?? "--") : "--";

        const disease = item.disease_id ? (diseaseNameById[String(item.disease_id)] ?? "--") : "--";

        const week = epiWeekLabelById[item.epi_week_id] ?? "--";

        return {
          id: item.id || `${item.district_id ?? "unknown"}-${item.disease_id ?? "none"}-${index}`,
          district,
          region,
          week,
          disease,
          severity: formatSeverity(item.status),
        };
      });
  }, [
    weeklyStatuses,
    districtId,
    regionId,
    diseaseName,
    districtNameById,
    regionNameById,
    diseaseNameById,
    epiWeekLabelById,
  ]);

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
