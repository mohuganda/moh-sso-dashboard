import {
  DataTable,
  Pagination,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from "@carbon/react";
import { useEffect, useMemo, useState } from "react";
import type { Alert } from "../../types";

interface DiseaseAlertsTableProps {
  alerts?: Alert[];
  districtId?: string;
  regionId?: string;
  diseaseName?: string;
  loading?: boolean;
  districtNameById?: Record<string, string>;
  regionNameById?: Record<string, string>;
  diseaseNameById?: Record<string, string>;
  epiWeekLabelById?: Record<string, string>;
  districtRegionByDistrictId?: Record<string, string>;
}

const headers = [
  { key: "district", header: "District" },
  { key: "region", header: "Region" },
  { key: "week", header: "Week" },
  { key: "disease", header: "Disease" },
  { key: "status", header: "Status" },
  { key: "narrative", header: "Narrative" },
];

function normalize(value?: unknown) {
  return String(value ?? "")
    .trim()
    .toLowerCase();
}

function formatStatus(value?: string) {
  if (!value) return "--";
  return value.charAt(0).toUpperCase() + value.slice(1).toLowerCase();
}

export function DiseaseAlertsTable({
  alerts = [],
  districtId = "",
  regionId = "",
  diseaseName = "",
  loading = false,
  districtNameById = {},
  regionNameById = {},
  diseaseNameById = {},
  epiWeekLabelById = {},
  districtRegionByDistrictId = {},
}: DiseaseAlertsTableProps) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);

  const rows = useMemo(() => {
    const normalizedDisease = normalize(diseaseName);

    return alerts
      .filter((item) => {
        const matchesDistrict = !districtId || String(item.district_id ?? "") === districtId;

        const alertRegionId = districtRegionByDistrictId[String(item.district_id ?? "")] ?? "";

        const matchesRegion = !regionId || alertRegionId === regionId;

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

        const resolvedRegionId = districtRegionByDistrictId[String(item.district_id ?? "")] ?? "";

        const region = resolvedRegionId ? (regionNameById[resolvedRegionId] ?? "--") : "--";

        const disease = item.disease_id ? (diseaseNameById[String(item.disease_id)] ?? "--") : "--";

        const week = item.epi_week_id ? (epiWeekLabelById[String(item.epi_week_id)] ?? "--") : "--";

        return {
          id: item.id || `${item.district_id ?? "unknown"}-${item.disease_id ?? "none"}-${index}`,
          district,
          region,
          week,
          disease,
          status: formatStatus(item.status),
          narrative: item.narrative || "--",
        };
      });
  }, [
    alerts,
    districtId,
    regionId,
    diseaseName,
    districtNameById,
    regionNameById,
    diseaseNameById,
    epiWeekLabelById,
    districtRegionByDistrictId,
  ]);

  useEffect(() => {
    setPage(1);
  }, [alerts, districtId, regionId, diseaseName]);

  const paginatedRows = useMemo(() => {
    const start = (page - 1) * pageSize;
    return rows.slice(start, start + pageSize);
  }, [rows, page, pageSize]);

  if (loading) {
    return <p>Loading alerts...</p>;
  }

  if (!rows.length) {
    return <p>No alerts available.</p>;
  }

  return (
    <div>
      <DataTable rows={paginatedRows} headers={headers} size="sm">
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

      <Pagination
        backwardText="Previous page"
        forwardText="Next page"
        itemsPerPageText="Items per page:"
        page={page}
        pageSize={pageSize}
        pageSizes={[5, 10, 20, 30, 50]}
        totalItems={rows.length}
        onChange={({ page, pageSize }) => {
          setPage(page);
          setPageSize(pageSize);
        }}
      />
    </div>
  );
}
