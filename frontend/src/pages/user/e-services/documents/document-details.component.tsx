import { useParams, useNavigate } from "react-router-dom";
import {
  Button,
  Tag,
  ProgressBar,
  Tile,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  Stack,
  Breadcrumb,
  BreadcrumbItem,
} from "@carbon/react";

import { useMemo } from "react";
import {
  useDeleteDocumentMutation,
  useGetDocumentProcessesQuery,
  useGetDocumentQuery,
  useLazyDownloadDocumentQuery,
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

export default function DocumentDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const {
    data: document,
    isLoading: docLoading,
    error: docError,
  } = useGetDocumentQuery(id!, { skip: !id });

  const { data: processes } = useGetDocumentProcessesQuery(id!, {
    skip: !id,
    pollingInterval: 3000,
    refetchOnFocus: true,
  });

  const [deleteDocument, { isLoading: deleting }] = useDeleteDocumentMutation();

  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const latestProcess = useMemo(() => {
    if (!processes || processes.length === 0) return undefined;

    return [...processes].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    )[0];
  }, [processes]);

  const handleDelete = async () => {
    if (!id) return;

    await deleteDocument(id);
    navigate(-1);
  };

  const handleDownload = async () => {
    if (!id || !document) return;

    const blob = await triggerDownload(id).unwrap();

    const url = window.URL.createObjectURL(blob);
    const link = window.document.createElement("a");
    link.href = url;
    link.download = document.original_filename;
    link.click();

    window.URL.revokeObjectURL(url);
  };

  if (docLoading) return <div>Loading...</div>;
  if (docError || !document) return <div>Failed to load document</div>;

  const historyRows =
    processes?.map((p, index) => ({
      id: p.id,
      attempt: index + 1,
      status: p.status,
      progress: p.progress,
      started: new Date(p.created_at).toLocaleString(),
      finished:
        p.status === "COMPLETED" || p.status === "FAILED"
          ? new Date(p.created_at).toLocaleString()
          : "-",
    })) ?? [];

  const headers = [
    { key: "attempt", header: "Attempt" },
    { key: "status", header: "Status" },
    { key: "progress", header: "Progress" },
    { key: "started", header: "Started" },
    { key: "finished", header: "Finished" },
  ];

  return (
    <div>
      <Breadcrumb style={{ marginBottom: "1rem" }}>
        <BreadcrumbItem onClick={() => navigate(-1)}>Documents</BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>{document.original_filename}</BreadcrumbItem>
      </Breadcrumb>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "2rem",
        }}
      >
        <div>
          <h2 style={{ marginBottom: "0.5rem" }}>{document.original_filename}</h2>

          {latestProcess && <StatusTag status={latestProcess.status} />}
        </div>
      </div>

      {/* --------------------------------------------
         Metadata Tile
      -------------------------------------------- */}

      <Tile style={{ marginBottom: "2rem" }}>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
            gap: "1rem",
          }}
        >
          <InfoItem label="Content Type" value={document.content_type.String} />
          <InfoItem label="Size" value={`${(document.size_bytes / 1024 / 1024).toFixed(2)} MB`} />
          <InfoItem label="Object Key" value={document.object_key} />
          <InfoItem label="Created At" value={new Date(document.created_at).toLocaleString()} />
        </div>
      </Tile>

      {/* --------------------------------------------
         Latest Process
      -------------------------------------------- */}

      {latestProcess && (
        <Tile style={{ marginBottom: "2rem" }}>
          <h4 style={{ marginBottom: "1rem" }}>Latest Processing Attempt</h4>

          <Stack gap={4}>
            <StatusTag status={latestProcess.status} />

            <div style={{ width: 300 }}>
              <ProgressBar value={latestProcess.progress} max={100} label="" size="small" />
            </div>

            <div>Last Update: {new Date(latestProcess.created_at).toLocaleString()}</div>
          </Stack>
        </Tile>
      )}

      {/* --------------------------------------------
         Action Buttons
      -------------------------------------------- */}

      <Stack orientation="horizontal" gap={4} style={{ marginBottom: "2rem" }}>
        <Button
          kind="primary"
          disabled={!latestProcess || latestProcess.status !== "COMPLETED"}
          onClick={handleDownload}
        >
          Download
        </Button>

        <Button kind="secondary" disabled={latestProcess?.status === "PROCESSING"}>
          Reprocess
        </Button>

        <Button kind="danger--tertiary" onClick={handleDelete} disabled={deleting}>
          Delete
        </Button>
      </Stack>

      {/* --------------------------------------------
         Process History Table
      -------------------------------------------- */}

      <h3 style={{ marginBottom: "1rem" }}>Process History</h3>

      <DataTable rows={historyRows} headers={headers}>
        {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
          <TableContainer>
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
                  <TableRow {...getRowProps({ row })}>
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

                      return <TableCell key={cell.id}>{cell.value}</TableCell>;
                    })}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DataTable>
    </div>
  );
}

/* --------------------------------------------
   Info Item
-------------------------------------------- */

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div style={{ fontSize: "0.75rem", color: "#6f6f6f" }}>{label}</div>
      <div style={{ fontWeight: 500 }}>{value}</div>
    </div>
  );
}
