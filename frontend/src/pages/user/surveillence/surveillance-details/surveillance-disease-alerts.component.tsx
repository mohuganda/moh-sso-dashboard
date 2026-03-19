import { DataTable } from "@carbon/react";

const rows = [
  {
    id: "1",
    createdAt: "2024-09-09",
    district: "Bunyabula",
    week: "37",
    disease: "Mpox",
    region: "Tooro",
  },
  {
    id: "2",
    createdAt: "2024-09-09",
    district: "Kampala",
    week: "37",
    disease: "Mpox",
    region: "Kampala",
  },
  {
    id: "3",
    createdAt: "2024-09-09",
    district: "Kicoga",
    week: "37",
    disease: "Measles",
    region: "North Central",
  },
];

const headers = [
  { key: "createdAt", header: "createdAt" },
  { key: "district", header: "district" },
  { key: "week", header: "week" },
  { key: "disease", header: "disease" },
  { key: "region", header: "region" },
];

export function DiseaseAlertsTable() {
  return (
    <DataTable rows={rows} headers={headers} size="sm">
      {({ rows, headers, getHeaderProps, getRowProps, getTableProps }) => (
        <table {...getTableProps()} className="cds--data-table cds--data-table--sm">
          <thead>
            <tr>
              {headers.map((header) => (
                <th {...getHeaderProps({ header })}>{header.header}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr {...getRowProps({ row })}>
                {row.cells.map((cell) => (
                  <td key={cell.id}>{cell.value}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </DataTable>
  );
}
