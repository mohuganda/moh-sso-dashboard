import { useState, useCallback } from "react";
import * as XLSX from "xlsx";
import {
  Button,
  Modal,
  FileUploaderDropContainer,
  FileUploaderItem,
  InlineLoading,
  InlineNotification,
  Stack,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from "@carbon/react";
import { Download } from "@carbon/react/icons";
import { useCreateIssueMutation } from "../api";
import { IMPORT_TEMPLATE_HEADERS } from "../lib/constants.ts";
import type { IssuePayload } from "../types";

type RowData = Record<string, unknown>;

type ImportSummary = {
  total: number;
  success: number;
  failed: number;
  errors: { row: number; message: string }[];
};

export const ImportIssuesModal = ({ onClose }: { onClose: () => void }) => {
  const [file, setFile] = useState<File | null>(null);
  const [parsedRows, setParsedRows] = useState<RowData[]>([]);
  const [isImporting, setIsImporting] = useState(false);
  const [summary, setSummary] = useState<ImportSummary | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [createIssue] = useCreateIssueMutation();

  const handleDrop = useCallback(
    (event: { target: { files: FileList } }) => {
      const selectedFile = event.target.files[0];
      if (!selectedFile) return;

      const ext = selectedFile.name.split(".").pop()?.toLowerCase();
      if (ext !== "xlsx" && ext !== "xls") {
        setError("Please upload an Excel file (.xlsx or .xls).");
        return;
      }

      setFile(selectedFile);
      setError(null);
      setSummary(null);
      parseExcelFile(selectedFile);
    },
    [],
  );

  const parseExcelFile = (excelFile: File) => {
    const reader = new FileReader();

    reader.onerror = () => {
      setError("Failed to read the Excel file.");
    };

    reader.onload = (event) => {
      try {
        const binary = event.target?.result;
        if (!binary) {
          setError("Invalid Excel file.");
          return;
        }

        const workbook = XLSX.read(binary, { type: "binary" });
        const firstSheetName = workbook.SheetNames[0];
        if (!firstSheetName) {
          setError("Excel file has no sheets.");
          return;
        }

        const firstSheet = workbook.Sheets[firstSheetName];
        const rows = XLSX.utils
          .sheet_to_json<RowData>(firstSheet, { defval: "" })
          .filter((row) =>
            Object.values(row).some(
              (value) => String(value ?? "").trim() !== "",
            ),
          );

        if (!rows.length) {
          setError("The Excel sheet contains no data rows.");
          return;
        }

        setParsedRows(rows);
      } catch (err) {
        setError("Failed to parse Excel file: " + (err as Error).message);
      }
    };

    reader.readAsBinaryString(excelFile);
  };

  const mapRowToPayload = (row: RowData): Partial<IssuePayload> => {
    return {
      dataset: String(row["dataset"] ?? ""),
      data_element: String(row["data_element"] ?? ""),
      org_unit: String(row["org_unit"] ?? ""),
      region: String(row["region"] ?? ""),
      district: String(row["district"] ?? ""),
      issue: String(row["issue"] ?? ""),
      issue_type: String(row["issue_type"] ?? ""),
      priority: String(row["priority"] ?? ""),
      severity: String(row["severity"] ?? ""),
      time_period: String(row["time_period"] ?? ""),
    };
  };

  const handleImport = async () => {
    if (!parsedRows.length) return;

    setIsImporting(true);
    setSummary(null);
    setError(null);

    const result: ImportSummary = {
      total: parsedRows.length,
      success: 0,
      failed: 0,
      errors: [],
    };

    for (let i = 0; i < parsedRows.length; i++) {
      try {
        const payload = mapRowToPayload(parsedRows[i]);
        await createIssue(payload as IssuePayload).unwrap();
        result.success++;
      } catch (err) {
        result.failed++;
        result.errors.push({
          row: i + 2, // +2 because row 1 is header
          message: (err as Error)?.message || "Unknown error",
        });
      }
    }

    setSummary(result);
    setIsImporting(false);
  };

  const downloadTemplate = () => {
    const wb = XLSX.utils.book_new();
    const ws = XLSX.utils.aoa_to_sheet([
      IMPORT_TEMPLATE_HEADERS,
      [
        "Example Dataset",
        "Example Data Element",
        "Example Org Unit",
        "Example issue description",
        "Outliers",
        "High",
        "Moderate",
        "2025Q1",
      ],
    ]);

    // Set column widths
    ws["!cols"] = IMPORT_TEMPLATE_HEADERS.map(() => ({ wch: 22 }));

    XLSX.utils.book_append_sheet(wb, ws, "Issues Template");
    XLSX.writeFile(wb, "issue-import-template.xlsx");
  };

  const previewHeaders =
    parsedRows.length > 0 ? Object.keys(parsedRows[0]) : [];
  const previewRows = parsedRows.slice(0, 10);

  const resetFile = () => {
    setFile(null);
    setParsedRows([]);
    setSummary(null);
    setError(null);
  };

  return (
    <Modal
      open
      modalHeading="Import Issues from Excel"
      primaryButtonText={summary ? "Close" : "Import"}
      secondaryButtonText={summary ? "Import Another" : "Cancel"}
      onRequestClose={onClose}
      onRequestSubmit={() => {
        if (summary) {
          onClose();
        } else {
          handleImport();
        }
      }}
      onSecondarySubmit={summary ? resetFile : onClose}
      size="lg"
    >
      <Stack gap={5}>
        {error && !file && (
          <InlineNotification
            kind="error"
            title="Error"
            subtitle={error}
            lowContrast
            onClose={() => setError(null)}
          />
        )}

        {/* Template Download */}
        <div className="import-template-section">
          <p className="cds--file--label">ISSUE IMPORT TEMPLATE</p>
          <p className="cds--label-description">
            Download the template to see the expected column format.
          </p>
          <Button
            kind="tertiary"
            size="sm"
            renderIcon={Download}
            onClick={downloadTemplate}
            style={{ marginTop: "0.5rem" }}
          >
            Download Template
          </Button>
        </div>

        {/* File Upload */}
        <div>
          <p className="cds--file--label">UPLOAD EXCEL FILE</p>
          <p className="cds--label-description">
            Supported file type: .xlsx, .xls
          </p>

          {!file ? (
            <FileUploaderDropContainer
              labelText="Drag and drop Excel file here or click to upload"
              accept={[".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={(event: any) => {
                if (event.target?.files?.length) {
                  handleDrop(event);
                }
              }}
            />
          ) : (
            <FileUploaderItem
              name={file.name}
              status="edit"
              onDelete={resetFile}
            />
          )}
        </div>

        {/* Parse Error */}
        {error && file && (
          <InlineNotification
            kind="error"
            title="Parse Error"
            subtitle={error}
            lowContrast
            onClose={() => setError(null)}
          />
        )}

        {/* Import Summary */}
        {summary && (
          <div className="import-summary">
            <InlineNotification
              kind={
                summary.failed === 0 ? "success" : "warning"
              }
              title="Import Complete"
              subtitle={`${summary.success} of ${summary.total} issues created successfully.${summary.failed > 0 ? ` ${summary.failed} failed.` : ""}`}
              lowContrast
            />
            {summary.errors.length > 0 && (
              <div className="import-errors">
                <p style={{ fontWeight: 600, marginBottom: "0.5rem" }}>
                  Row Errors:
                </p>
                <ul>
                  {summary.errors.slice(0, 10).map((e, idx) => (
                    <li key={idx}>
                      Row {e.row}: {e.message}
                    </li>
                  ))}
                  {summary.errors.length > 10 && (
                    <li>...and {summary.errors.length - 10} more errors.</li>
                  )}
                </ul>
              </div>
            )}
          </div>
        )}

        {/* Importing Progress */}
        {isImporting && (
          <InlineLoading description="Importing issues..." />
        )}

        {/* Preview Table */}
        {parsedRows.length > 0 && !summary && (
          <div className="import-preview">
            <p className="cds--file--label" style={{ marginBottom: "0.5rem" }}>
              PREVIEW: {file?.name}
            </p>
            <p className="cds--label-description">
              Showing first {previewRows.length} rows of {parsedRows.length}{" "}
              total rows.
            </p>
            <div className="import-preview-table-wrapper">
              <Table size="sm" aria-label="Import preview table">
                <TableHead>
                  <TableRow>
                    {previewHeaders.map((header) => (
                      <TableHeader key={header}>
                        {header}
                      </TableHeader>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {previewRows.map((row, rowIndex) => (
                    <TableRow key={rowIndex}>
                      {previewHeaders.map((header) => (
                        <TableCell key={`${rowIndex}-${header}`}>
                          {String(row[header] ?? "")}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        )}
      </Stack>
    </Modal>
  );
};

