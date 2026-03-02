import { useMemo, useState } from "react";
import Papa from "papaparse";
import * as XLSX from "xlsx";
import "./file-upload.css";
import {
  Button,
  FileUploaderDropContainer,
  FileUploaderItem,
  FormItem,
  InlineNotification,
  Stack,
  Select,
  SelectItem,
  Loading,
} from "@carbon/react";
import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
} from "../../../store/api/document.api";

type UploadStatus = "edit" | "complete" | "uploading";
type Row = Record<string, unknown>;

const ACCEPTED_EXTS = [".csv", ".xlsx", ".xls"] as const;

function getExtension(name: string): string {
  const idx = name.lastIndexOf(".");
  return idx >= 0 ? name.slice(idx).toLowerCase() : "";
}

const FileUpload = () => {
  const [data, setData] = useState<Row[]>([]);
  const [headers, setHeaders] = useState<string[]>([]);
  const [fileName, setFileName] = useState("");
  const [file, setFile] = useState<File | null>(null);

  const [showUploader, setShowUploader] = useState(true);
  const [uploadStatus, setUploadStatus] = useState<UploadStatus>("edit");
  const [error, setError] = useState<string | null>(null);

  // storage location selection
  const [storageLocation, setStorageLocation] = useState<string>("");

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
    isError: isLocationsError,
  } = useListStorageLocationsQuery();

  const activeLocations = useMemo(() => locations.filter((l) => l.is_active), [locations]);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const previewRows = useMemo(() => data.slice(0, 10), [data]);

  const reset = () => {
    setData([]);
    setHeaders([]);
    setFileName("");
    setFile(null);
    setShowUploader(true);
    setUploadStatus("edit");
    setError(null);
    setStorageLocation("");
  };

  const parseCSV = (f: File) =>
    new Promise<void>((resolve, reject) => {
      Papa.parse<Row>(f, {
        header: true,
        skipEmptyLines: true,
        complete: (results) => {
          const rows = results.data || [];
          if (!rows.length) {
            reject(new Error("CSV file contains no rows."));
            return;
          }
          const cols = Object.keys(rows[0] || {});
          setHeaders(cols);
          setData(rows);
          resolve();
        },
        error: (err) => reject(err),
      });
    });

  const parseExcel = (f: File) =>
    new Promise<void>((resolve, reject) => {
      const reader = new FileReader();
      reader.onerror = () => reject(new Error("Failed to read Excel file."));
      reader.onload = (e) => {
        try {
          const binaryStr = e?.target?.result;
          if (!binaryStr) throw new Error("Invalid Excel file.");

          const workbook = XLSX.read(binaryStr, { type: "binary" });
          const sheetName = workbook.SheetNames[0];
          if (!sheetName) throw new Error("Excel file has no sheets.");

          const sheet = workbook.Sheets[sheetName];
          const json = XLSX.utils.sheet_to_json<Row>(sheet, { defval: "" });

          if (!json.length) throw new Error("Excel sheet contains no rows.");

          const cols = Object.keys(json[0] || {});
          setHeaders(cols);
          setData(json);
          resolve();
        } catch (err: any) {
          reject(err);
        }
      };
      reader.readAsBinaryString(f);
    });

  // Carbon FileUploaderDropContainer calls onAddFiles({ addedFiles })
  const handleFileUpload = async (evt: { addedFiles?: File[] } | any) => {
    const selected: File | undefined =
      evt?.addedFiles?.[0] ?? evt?.target?.files?.[0] ?? evt?.dataTransfer?.files?.[0];

    if (!selected) return;

    setError(null);
    setUploadStatus("uploading");
    setShowUploader(false);
    setFile(selected);
    setFileName(selected.name);

    const ext = getExtension(selected.name);
    if (!ACCEPTED_EXTS.includes(ext as any)) {
      setError("Please upload a CSV or Excel file (.csv, .xlsx, .xls).");
      setUploadStatus("edit");
      setShowUploader(true);
      setFile(null);
      setFileName("");
      return;
    }

    try {
      if (ext === ".csv") await parseCSV(selected);
      else await parseExcel(selected);

      setUploadStatus("complete");
      window.setTimeout(() => setUploadStatus("edit"), 1500);
    } catch (e: any) {
      setError(e?.message || "Failed to parse file. Please try again.");
      setUploadStatus("edit");
      setShowUploader(true);
      setFile(null);
      setFileName("");
    }
  };

  const handleFileDelete = () => reset();

  const handleUploadToBackend = async () => {
    if (!file) {
      setError("Please select a file first.");
      return;
    }
    if (!storageLocation) {
      setError("Please select a storage location.");
      return;
    }

    try {
      setError(null);

      await createDocument({
        file,
        storageLocation, // IMPORTANT: this is what your API expects (string)
      }).unwrap();

      // Reset after successful upload
      reset();
    } catch (err: any) {
      setError(err?.data?.message || "Upload failed. Please try again.");
    }
  };

  return (
    <>
      <FormItem className="uploader-form-item">
        <p className="cds--file--label">UPLOAD A FILE & PREVIEW DATA</p>
        <p className="cds--label-description">Supported file types are .csv .xlsx and .xls.</p>

        <Stack gap={4}>
          {error && (
            <InlineNotification
              kind="error"
              title="Upload Error"
              subtitle={error}
              lowContrast
              onCloseButtonClick={() => setError(null)}
            />
          )}

          {showUploader ? (
            <FileUploaderDropContainer
              accept={[".csv", ".xlsx", ".xls"]}
              labelText="Drag and drop a file here or click to upload"
              onAddFiles={handleFileUpload}
              tabIndex={0}
            />
          ) : (
            <FileUploaderItem
              iconDescription="Delete file"
              name={fileName}
              onDelete={handleFileDelete}
              size="lg"
              status={uploadStatus}
            />
          )}

          {isLocationsLoading ? (
            <Loading description="Loading storage locations..." />
          ) : isLocationsError ? (
            <InlineNotification
              kind="error"
              title="Storage Locations Error"
              subtitle="Failed to load storage locations."
              lowContrast
            />
          ) : (
            <Select
              id="storage-location"
              labelText="Storage Location"
              value={storageLocation}
              onChange={(e) => setStorageLocation(e.target.value)}
              disabled={activeLocations.length === 0}
            >
              <SelectItem value="" text="Select location" />
              {activeLocations.map((loc) => (
                <SelectItem
                  key={loc.id}
                  value={loc.id}
                  text={`${loc.name} (${loc.provider.toUpperCase()})`}
                />
              ))}
            </Select>
          )}
        </Stack>
      </FormItem>

      {data.length > 0 && (
        <div className="preview-container">
          <p className="cds--file--label">PREVIEW: {fileName}</p>

          <div style={{ overflowX: "auto", maxHeight: 400, border: "1px solid #ddd" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead style={{ backgroundColor: "#f4f4f4", position: "sticky", top: 0 }}>
                <tr>
                  {headers.map((h) => (
                    <th key={h} style={{ border: "1px solid #ddd", padding: 8, textAlign: "left" }}>
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>

              <tbody>
                {previewRows.map((row, i) => (
                  <tr key={i}>
                    {headers.map((h) => (
                      <td key={h} style={{ border: "1px solid #ddd", padding: 8 }}>
                        {String(row[h] ?? "")}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <p>
            <i>Showing first 10 rows of {data.length} total rows.</i>
          </p>

          <div className="file-upload-btn-container">
            <Button
              size="md"
              kind="tertiary"
              onClick={handleUploadToBackend}
              disabled={isUploading || isLocationsLoading || !storageLocation}
            >
              {isUploading ? "Uploading..." : "Upload to Backend"}
            </Button>

            <Button size="md" kind="secondary" onClick={handleFileDelete} disabled={isUploading}>
              Cancel
            </Button>
          </div>
        </div>
      )}
    </>
  );
};

export default FileUpload;
