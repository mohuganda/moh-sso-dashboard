import { DataTable } from "@carbon/react";

const facilityRows = [
  {
    id: "1",
    facility: "Ngoleriet Health Centre II",
    region: "Karamoja",
    district: "NAPAK",
    subcounty: "NGOLERIET",
    disease: "Malaria",
    weeks: "36",
    flag: "Red",
    cases: "216",
  },
  {
    id: "2",
    facility: "Oriajini Hospital",
    region: "West Nile",
    district: "TEREGO",
    subcounty: "KATRINI",
    disease: "Malaria",
    weeks: "36",
    flag: "Red",
    cases: "47",
  },
];

const facilityHeaders = [
  { key: "facility", header: "facility" },
  { key: "region", header: "region" },
  { key: "district", header: "district" },
  { key: "subcounty", header: "subcounty" },
  { key: "disease", header: "disease" },
  { key: "weeks", header: "weeks" },
  { key: "flag", header: "flag" },
  { key: "cases", header: "cases" },
];

export function FacilitiesActionTable() {
  return (
    <DataTable rows={facilityRows} headers={facilityHeaders} size="sm">
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
