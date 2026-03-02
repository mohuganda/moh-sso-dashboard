import { useState, useMemo } from "react";
import {
  Button,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableContainer,
  TableToolbar,
  TableToolbarContent,
  TableToolbarSearch,
  Select,
  SelectItem,
  Loading,
} from "@carbon/react";
import { Upload } from "@carbon/react/icons";
import { EmptyState } from "../../../../components/emptystate/EmptyState";
import { useHeaderPanel } from "../../../../components/header-panel/header-panel.context";
import { UploadDocumentModal } from "./UploadDocumentModal";
import { useListDocumentsQuery } from "../../../../store/api/document.api";
import { DocumentRow } from "./document-row.component";

export default function DocumentPage() {
  const { openPanel } = useHeaderPanel();

  const [statusFilter, setStatusFilter] = useState<string>("ALL");

  // Load documents only
  const { data: documents = [], isLoading } = useListDocumentsQuery();

  // Filtered documents
  const filteredDocs = useMemo(() => {
    if (statusFilter === "ALL") return documents;

    // Filtering will be handled inside row via process polling,
    // so for now we keep full list (optional optimization later)
    return documents;
  }, [documents, statusFilter]);

  const headers = [
    { key: "filename", header: "Filename" },
    { key: "type", header: "Type" },
    { key: "size", header: "Size" },
    { key: "status", header: "Status" },
    { key: "progress", header: "Progress" },
    { key: "actions", header: "" },
  ];

  if (isLoading) {
    return <Loading description="Loading documents..." />;
  }

  return (
    <div>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "2rem",
        }}
      >
        <div>
          <h2 style={{ marginBottom: "0.5rem" }}>Document Management</h2>
          <p style={{ color: "#6f6f6f" }}>Upload and monitor document import processes.</p>
        </div>

        <Button
          renderIcon={Upload}
          onClick={() =>
            openPanel({
              title: "Upload Document",
              content: <UploadDocumentModal onClose={() => {}} />,
              size: "lg",
            })
          }
        >
          Upload Document
        </Button>
      </div>

      {filteredDocs.length === 0 ? (
        <EmptyState
          title="No documents found"
          description="Upload a document to begin processing."
        />
      ) : (
        <DataTable rows={filteredDocs} headers={headers}>
          {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
            <TableContainer>
              <TableToolbar>
                <TableToolbarContent>
                  <TableToolbarSearch />

                  <Select
                    id="status-filter"
                    size="sm"
                    labelText=""
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                    style={{ width: 180 }}
                  >
                    <SelectItem value="ALL" text="All Status" />
                    <SelectItem value="PENDING" text="Pending" />
                    <SelectItem value="PROCESSING" text="Processing" />
                    <SelectItem value="COMPLETED" text="Completed" />
                    <SelectItem value="FAILED" text="Failed" />
                  </Select>
                </TableToolbarContent>
              </TableToolbar>

              <Table {...getTableProps()}>
                <TableHead>
                  <TableRow>
                    {headers.map((header) => (
                      <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                    ))}
                  </TableRow>
                </TableHead>

                <TableBody>
                  {rows.map((row, index) => {
                    const document = filteredDocs[index];

                    return (
                      <DocumentRow
                        key={document.id}
                        document={document}
                        getRowProps={getRowProps}
                      />
                    );
                  })}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DataTable>
      )}
    </div>
  );
}
