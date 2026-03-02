import { useState, useMemo } from "react";
import {
  Button,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  TableToolbar,
  TableToolbarContent,
  TableToolbarSearch,
  Tag,
  ProgressBar,
  Select,
  SelectItem,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import { Upload } from "@carbon/icons-react";
import { EmptyState } from "../../../../components/emptystate/EmptyState";
import { useNavigate } from "react-router-dom";
import { useHeaderPanel } from "../../../../components/header-panel/header-panel.context";
import { UploadDocumentModal } from "./UploadDocumentModal";
import {
  useGetDocumentProcessesQuery,
  useListDocumentsQuery,
} from "../../../../store/api/document.api";

/* --------------------------------------------
   Status Tag
-------------------------------------------- */

function StatusTag({ status }: { status: string }) {
  const colorMap: Record<string, any> = {
    PENDING: "gray",
    PROCESSING: "blue",
    COMPLETED: "green",
    FAILED: "red",
    CANCELLED: "magenta",
  };

  return <Tag type={colorMap[status] || "gray"}>{status}</Tag>;
}

/* --------------------------------------------
   Page
-------------------------------------------- */

export default function DocumentPage() {
  const { openPanel } = useHeaderPanel();
  const navigate = useNavigate();

  const [statusFilter, setStatusFilter] = useState<string>("ALL");

  // 🔥 Load documents
  const { data: documents, isLoading } = useListDocumentsQuery();

  // 🔥 Poll processes for each document
  const documentProcesses = documents?.map((doc) =>
    useGetDocumentProcessesQuery(doc.id, {
      pollingInterval: 3000,
      skip: !doc.id,
    }),
  );

  /**
   * Merge document + latest process
   */
  const rows = useMemo(() => {
    if (!documents) return [];

    return documents.map((doc, index) => {
      const processes = documentProcesses?.[index]?.data ?? [];

      const latest =
        processes.length > 0
          ? [...processes].sort(
              (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
            )[0]
          : undefined;

      return {
        id: doc.id,
        filename: doc.original_filename,
        type: doc.content_type,
        size: `${(doc.size_bytes / 1024 / 1024).toFixed(2)} MB`,
        status: latest?.status ?? "PENDING",
        progress: latest?.progress ?? 0,
      };
    });
  }, [documents, documentProcesses]);

  const filteredDocs = useMemo(() => {
    if (statusFilter === "ALL") return rows;
    return rows.filter((d) => d.status === statusFilter);
  }, [rows, statusFilter]);

  const headers = [
    { key: "filename", header: "Filename" },
    { key: "type", header: "Type" },
    { key: "size", header: "Size" },
    { key: "status", header: "Status" },
    { key: "progress", header: "Progress" },
    { key: "actions", header: "" },
  ];

  if (isLoading) return <div>Loading...</div>;

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
          onClick={() => {
            openPanel({
              title: "Upload Document",
              content: <UploadDocumentModal onClose={() => {}} />,
              size: "lg",
            });
          }}
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
                  {rows.map((row) => (
                    <TableRow style={{ cursor: "pointer" }} {...getRowProps({ row })}>
                      {row.cells.map((cell) => {
                        if (cell.info.header === "status") {
                          return (
                            <TableCell key={cell.id}>
                              <StatusTag status={cell.value as string} />
                            </TableCell>
                          );
                        }

                        if (cell.info.header === "progress") {
                          return (
                            <TableCell key={cell.id}>
                              <div style={{ width: 150 }}>
                                <ProgressBar
                                  value={cell.value as number}
                                  max={100}
                                  label=""
                                  size="small"
                                />
                              </div>
                            </TableCell>
                          );
                        }

                        if (cell.info.header === "actions") {
                          return (
                            <TableCell key={cell.id}>
                              <OverflowMenu size="sm">
                                <OverflowMenuItem
                                  itemText="View Details"
                                  onClick={() =>
                                    navigate(
                                      `/apps/utilities/self-service/eservice/document-upload/${row.id}`,
                                    )
                                  }
                                />
                              </OverflowMenu>
                            </TableCell>
                          );
                        }

                        return <TableCell key={cell.id}>{cell.value}</TableCell>;
                      })}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DataTable>
      )}
    </div>
  );
}
