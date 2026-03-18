import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  Button,
  DataTable,
  Loading,
  Select,
  SelectItem,
  Table,
  TableBody,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  TableToolbar,
  TableToolbarContent,
  TableToolbarSearch,
} from "@carbon/react";
import { Upload } from "@carbon/react/icons";

import { EmptyState } from "../../../../components/emptystate/EmptyState";
import { useHeaderPanel } from "../../../../components/header-panel/header-panel.context";
import { UploadDocumentModal } from "./UploadDocumentModal";
import { useListDocumentsQuery } from "../../../../store/api/document.api";
import { DocumentRow } from "./document-row.component";

const headers = [
  { key: "filename", header: "Filename" },
  { key: "type", header: "Type" },
  { key: "size", header: "Size" },
  { key: "status", header: "Status" },
  { key: "progress", header: "Progress" },
  { key: "actions", header: "" },
];

function getDocumentSearchText(document: any) {
  return [
    document?.filename,
    document?.original_name,
    document?.name,
    document?.type,
    document?.mime_type,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}

function getDocumentStatus(document: any) {
  return (
    document?.status || document?.process_status || document?.latest_process?.status || "UNKNOWN"
  );
}

export default function DocumentPage() {
  const { openPanel } = useHeaderPanel();

  const [statusFilter, setStatusFilter] = useState("ALL");
  const [searchTerm, setSearchTerm] = useState("");

  const { data: documents = [], isLoading } = useListDocumentsQuery();

  const filteredDocs = useMemo(() => {
    return documents.filter((document: any) => {
      const matchesSearch =
        searchTerm.trim() === "" ||
        getDocumentSearchText(document).includes(searchTerm.trim().toLowerCase());

      const documentStatus = getDocumentStatus(document);
      const matchesStatus = statusFilter === "ALL" || documentStatus === statusFilter;

      return matchesSearch && matchesStatus;
    });
  }, [documents, searchTerm, statusFilter]);

  const handleOpenUpload = () => {
    openPanel({
      title: "Upload Document",
      content: <UploadDocumentModal onClose={() => {}} />,
      size: "lg",
    });
  };

  if (isLoading) {
    return <Loading description="Loading documents..." withOverlay={false} />;
  }

  return (
    <div>
      <div style={{ marginBottom: "1rem" }}>
        <Breadcrumb noTrailingSlash>
          <BreadcrumbItem>
            <RouterLink to="/">Home</RouterLink>
          </BreadcrumbItem>
          <BreadcrumbItem isCurrentPage>Documents</BreadcrumbItem>
        </Breadcrumb>
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: "1rem",
          marginBottom: "2rem",
          flexWrap: "wrap",
        }}
      >
        <div>
          <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Document Management</h2>
          <p style={{ margin: 0, color: "#6f6f6f" }}>
            Upload documents and monitor import progress.
          </p>
        </div>

        <Button renderIcon={Upload} onClick={handleOpenUpload}>
          Upload Document
        </Button>
      </div>

      {documents.length === 0 ? (
        <EmptyState
          title="No documents found"
          description="Upload a document to begin processing."
        />
      ) : (
        <DataTable rows={filteredDocs} headers={headers}>
          {({ headers, getTableProps, getHeaderProps, getRowProps }) => (
            <TableContainer title="Documents">
              <TableToolbar>
                <TableToolbarContent>
                  <TableToolbarSearch
                    persistent
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />

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

              {filteredDocs.length === 0 ? (
                <div style={{ padding: "2rem 0" }}>
                  <EmptyState
                    title="No matching documents"
                    description="Try changing your search or status filter."
                  />
                </div>
              ) : (
                <Table {...getTableProps()} aria-label="Documents table">
                  <TableHead>
                    <TableRow>
                      {headers.map((header) => (
                        <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {filteredDocs.map((document: any) => (
                      <DocumentRow
                        key={document.id}
                        document={document}
                        getRowProps={getRowProps}
                      />
                    ))}
                  </TableBody>
                </Table>
              )}
            </TableContainer>
          )}
        </DataTable>
      )}
    </div>
  );
}
