import {
  Button,
  DataTable,
  Pagination,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from "@carbon/react";
import { ChartLine } from "@carbon/icons-react";
import { useEffect, useMemo, useState } from "react";

type FacilityWeeklyMetricRow = {
  id: string;
  source_record_id?: string | null;
  facility_id: string;
  facility_name: string;
  sub_county_id?: string | null;
  sub_county_name?: string | null;
  district_id?: string | null;
  district_name?: string | null;
  region_id?: string | null;
  region_name?: string | null;
  disease_id?: string | null;
  disease_name?: string | null;
  indicator_id?: string | null;
  indicator_name?: string | null;
  epi_week_id: string;
  metric_value: number;
  source_name?: string | null;
  imported_at?: string;
  created_at?: string;
  epi_year?: number;
  epi_week?: number;
};

interface FacilityTrendActionPayload {
  facilityId?: string;
  facilityName: string;
  regionId?: string;
  districtId?: string;
  subCountyId?: string;
  diseaseId?: string;
  indicatorId?: string;
  diseaseName?: string;
  indicatorName?: string;
}

interface FacilitiesActionTableProps {
  facilityMetrics?: FacilityWeeklyMetricRow[];
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
  onViewFacilityTrend?: (payload: FacilityTrendActionPayload) => void;
}

const facilityHeaders = [
  { key: "facility", header: "Facility" },
  { key: "region", header: "Region" },
  { key: "district", header: "District" },
  { key: "subcounty", header: "Subcounty" },
  { key: "disease", header: "Disease" },
  { key: "week", header: "Week" },
  { key: "cases", header: "Cases" },
  { key: "actions", header: "Actions" },
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
  onViewFacilityTrend,
}: FacilitiesActionTableProps) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const rows = useMemo(() => {
    return facilityMetrics
      .filter((item) => {
        const matchesRegion = !regionId || String(item.region_id ?? "") === regionId;
        const matchesDistrict = !districtId || String(item.district_id ?? "") === districtId;
        const matchesSubCounty = !subCountyId || String(item.sub_county_id ?? "") === subCountyId;

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

        const subcounty = item.sub_county_id
          ? (subCountyNameById[String(item.sub_county_id)] ?? item.sub_county_name ?? "--")
          : (item.sub_county_name ?? "--");

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
          ? (epiWeekLabelById[String(item.epi_week_id)] ??
            (item.epi_week ? `Week ${item.epi_week}` : "--"))
          : item.epi_week
            ? `Week ${item.epi_week}`
            : "--";

        return {
          id: String(item.id ?? item.facility_id ?? `${item.facility_name ?? "facility"}-${index}`),
          facility: item.facility_name ?? "--",
          region,
          district,
          subcounty,
          disease,
          week,
          cases: String(numericCases),
          actions: "view-trend",
          numericCases,
          raw: item,
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

  useEffect(() => {
    setPage(1);
  }, [facilityMetrics, regionId, districtId, subCountyId]);

  const paginatedRows = useMemo(() => {
    const start = (page - 1) * pageSize;
    return rows.slice(start, start + pageSize);
  }, [rows, page, pageSize]);

  if (loading) {
    return <p>Loading facility action data...</p>;
  }

  if (!rows.length) {
    return <p>No facility action data available.</p>;
  }

  return (
    <div>
      <DataTable
        rows={paginatedRows.map(({ numericCases, raw, ...row }) => row)}
        headers={facilityHeaders}
        size="sm"
      >
        {({ rows: tableRows, headers, getHeaderProps, getRowProps, getTableProps }) => (
          <Table {...getTableProps()} size="sm">
            <TableHead>
              <TableRow>
                {headers.map((header) => (
                  <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                ))}
              </TableRow>
            </TableHead>

            <TableBody>
              {tableRows.map((row) => {
                const sourceRow = paginatedRows.find((item) => item.id === row.id);

                return (
                  <TableRow {...getRowProps({ row })}>
                    {row.cells.map((cell) => {
                      if (cell.info.header === "actions") {
                        const raw = sourceRow?.raw;

                        return (
                          <TableCell key={cell.id}>
                            <Button
                              kind="ghost"
                              size="sm"
                              renderIcon={ChartLine}
                              iconDescription="View facility trend"
                              disabled={!onViewFacilityTrend}
                              onClick={() => {
                                if (!raw || !onViewFacilityTrend) return;

                                onViewFacilityTrend({
                                  facilityId: raw.facility_id ? String(raw.facility_id) : undefined,
                                  facilityName: raw.facility_name ?? "--",
                                  regionId: raw.region_id ? String(raw.region_id) : undefined,
                                  districtId: raw.district_id ? String(raw.district_id) : undefined,
                                  subCountyId: raw.sub_county_id
                                    ? String(raw.sub_county_id)
                                    : undefined,
                                  diseaseId: raw.disease_id ? String(raw.disease_id) : undefined,
                                  indicatorId: raw.indicator_id
                                    ? String(raw.indicator_id)
                                    : undefined,
                                  diseaseName: raw.disease_name ?? undefined,
                                  indicatorName: raw.indicator_name ?? undefined,
                                });
                              }}
                            >
                              View trends
                            </Button>
                          </TableCell>
                        );
                      }

                      return <TableCell key={cell.id}>{cell.value}</TableCell>;
                    })}
                  </TableRow>
                );
              })}
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
        pageSizes={[10, 20, 30, 50]}
        totalItems={rows.length}
        onChange={({ page, pageSize }) => {
          setPage(page);
          setPageSize(pageSize);
        }}
      />
    </div>
  );
}
