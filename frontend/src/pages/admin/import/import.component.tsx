import {
  Button,
  FileUploaderDropContainer,
  InlineNotification,
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
  Loading,
} from "@carbon/react";
import { useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE as string;

type ImportRow = {
  rowNumber: number;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  role: string;
  enabled: boolean;
  clientIds: string[];
  status: string; // valid|invalid|success|failed|skipped
  errors?: string[];
  errorMsg?: string;
};

type PreviewResponse = {
  jobId: string;
  fileName: string;
  total: number;
  valid: number;
  invalid: number;
  rows: ImportRow[];
};

type ExecuteResponse = {
  jobId: string;
  total: number;
  successCount: number;
  failureCount: number;
  status: string;
};

export default function UserBulkImportPage() {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [executed, setExecuted] = useState<ExecuteResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<{
    kind: "error" | "success" | "info";
    title: string;
    subtitle?: string;
  } | null>(null);

  const headers = useMemo(
    () => [
      { key: "rowNumber", header: "Row" },
      { key: "username", header: "Username" },
      { key: "email", header: "Email" },
      { key: "name", header: "Name" },
      { key: "role", header: "Role" },
      { key: "enabled", header: "Enabled" },
      { key: "status", header: "Status" },
      { key: "errors", header: "Errors" },
    ],
    [],
  );

  const rows = useMemo(() => {
    const src = preview?.rows ?? [];
    return src.map((r) => ({
      id: String(r.rowNumber),
      rowNumber: r.rowNumber,
      username: r.username,
      email: r.email,
      name: `${r.firstName} ${r.lastName}`,
      role: r.role,
      enabled: String(r.enabled),
      status: r.status,
      errors: r.errorMsg || (r.errors ? r.errors.join("; ") : ""),
    }));
  }, [preview]);

  const downloadTemplate = () => {
    window.open(`${API_BASE}/admin/users/import/template.csv`, "_blank");
  };

  const downloadErrors = () => {
    if (!preview?.jobId) return;
    window.open(`${API_BASE}/admin/users/import/${preview.jobId}/errors.csv`, "_blank");
  };

  const onPickFile = (f: File) => {
    setFile(f);
    setPreview(null);
    setExecuted(null);
    setNotice(null);
  };

  const previewUpload = async () => {
    if (!file) {
      setNotice({ kind: "error", title: "No file selected" });
      return;
    }

    setLoading(true);
    setNotice(null);

    try {
      const form = new FormData();
      form.append("file", file);

      const res = await fetch(`${API_BASE}/admin/users/import/preview`, {
        method: "POST",
        credentials: "include",
        body: form,
      });

      const json = await res.json();
      if (!res.ok) throw new Error(json?.error || `Preview failed (${res.status})`);

      setPreview(json);
      setNotice({
        kind: "success",
        title: "Preview generated",
        subtitle: `Total: ${json.total} | Valid: ${json.valid} | Invalid: ${json.invalid}`,
      });
    } catch (e: any) {
      setNotice({
        kind: "error",
        title: "Preview failed",
        subtitle: e.message,
      });
    } finally {
      setLoading(false);
    }
  };

  const executeImport = async () => {
    if (!preview?.jobId) return;

    setLoading(true);
    setNotice(null);

    try {
      const res = await fetch(`${API_BASE}/admin/users/import/execute`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ jobId: preview.jobId }),
      });

      const json = await res.json();
      if (!res.ok) throw new Error(json?.error || `Execute failed (${res.status})`);

      setExecuted(json);
      setNotice({
        kind: "success",
        title: "Import completed",
        subtitle: `Success: ${json.successCount} | Failed: ${json.failureCount}`,
      });

      // Refresh status/details
      const statusRes = await fetch(`${API_BASE}/admin/users/import/${preview.jobId}`, {
        method: "GET",
        credentials: "include",
      });
      const statusJson = await statusRes.json();
      if (statusRes.ok) setPreview({ ...preview, rows: statusJson.rows });
    } catch (e: any) {
      setNotice({ kind: "error", title: "Import failed", subtitle: e.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "1rem" }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "0.75rem",
          marginBottom: "1rem",
        }}
      >
        <h3 style={{ margin: 0 }}>Bulk Import Users</h3>
        <Button size="sm" kind="tertiary" onClick={downloadTemplate}>
          Download CSV Template
        </Button>
      </div>

      {notice && (
        <div style={{ marginBottom: "1rem" }}>
          <InlineNotification
            kind={notice.kind}
            title={notice.title}
            subtitle={notice.subtitle}
            lowContrast
            hideCloseButton={false}
            onCloseButtonClick={() => {
              setNotice(null);
            }}
          />
        </div>
      )}

      <div style={{ marginBottom: "1rem" }}>
        <FileUploaderDropContainer
          labelText="Drag and drop CSV here or click to upload"
          multiple={false}
          accept={[".csv"]}
          onAddFiles={({ addedFiles }: any) => {
            const f = addedFiles?.[0];
            if (f) onPickFile(f);
          }}
        />
        {file && (
          <div style={{ marginTop: "0.5rem" }}>
            Selected: <strong>{file.name}</strong>
          </div>
        )}
      </div>

      <div style={{ display: "flex", gap: "0.75rem", marginBottom: "1rem" }}>
        <Button kind="primary" disabled={!file || loading} onClick={previewUpload}>
          Preview
        </Button>
        <Button
          kind="danger"
          disabled={!preview?.jobId || loading || (preview?.invalid ?? 0) > 0}
          onClick={executeImport}
        >
          Execute Import
        </Button>
        <Button kind="secondary" disabled={!preview?.jobId || loading} onClick={downloadErrors}>
          Download Errors CSV
        </Button>
      </div>

      {loading && <Loading withOverlay description="Working..." />}

      {preview && (
        <div
          style={{
            marginBottom: "1rem",
            display: "flex",
            gap: "0.5rem",
            flexWrap: "wrap",
          }}
        >
          <Tag type="gray">Total: {preview.total}</Tag>
          <Tag type="green">Valid: {preview.valid}</Tag>
          <Tag type="red">Invalid: {preview.invalid}</Tag>
          {executed && <Tag type="blue">Status: {executed.status}</Tag>}
        </div>
      )}

      {preview && (
        <DataTable rows={rows} headers={headers} isSortable>
          {({ rows, headers, getHeaderProps, getRowProps }: any) => (
            <TableContainer title="Preview / Results">
              <Table size="md" useZebraStyles>
                <TableHead>
                  <TableRow>
                    {headers.map((h: any) => (
                      <TableHeader key={h.key} {...getHeaderProps({ header: h })}>
                        {h.header}
                      </TableHeader>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {rows.map((r: any) => (
                    <TableRow key={r.id} {...getRowProps({ row: r })}>
                      {r.cells.map((cell: any) => (
                        <TableCell key={cell.id}>
                          {cell.info.header === "status" ? (
                            <Tag
                              type={
                                cell.value === "valid" || cell.value === "success"
                                  ? "green"
                                  : cell.value === "skipped"
                                    ? "gray"
                                    : "red"
                              }
                            >
                              {cell.value}
                            </Tag>
                          ) : (
                            cell.value
                          )}
                        </TableCell>
                      ))}
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
