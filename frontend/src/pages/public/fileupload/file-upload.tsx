import { useState } from "react";
import Papa from "papaparse";
import * as XLSX from "xlsx";
import "./file-upload.css";
import {Button, FileUploaderDropContainer, FileUploaderItem, FormItem} from "@carbon/react";

const FileUpload = () => {
    const [data, setData] = useState([]);
    const [headers, setHeaders] = useState([]);
    const [fileName, setFileName] = useState("");
    const [showUploader, setShowUploader] = useState(true);
    const [uploadStatus, setUploadStatus] = useState<"edit"|"complete"|"uploading">("uploading");

    const handleFileUpload = (event) => {
        setShowUploader(false);
        const file = event?.type === "drop" ? event?.dataTransfer?.files[0] : event?.target?.files[0];
        if (!file) return;

        setFileName(file.name);
        const fileExtension = file.name.split(".").pop().toLowerCase();

        if (fileExtension === "csv") {
            parseCSV(file);
            setUploadStatus("complete");
        } else if (["xlsx", "xls"].includes(fileExtension)) {
            parseExcel(file);
            setUploadStatus("complete");
        } else {
            alert("Please upload a CSV or Excel file.");
        }

        setTimeout(() => {
            setUploadStatus("edit");
        }, 2000);

    };

    const handleFileDelete = () => {
        setData([]);
        setHeaders([]);
        setFileName("");
        setShowUploader(true);
    }

    // Process CSV Files
    const parseCSV = (file) => {
        Papa.parse(file, {
            header: true,
            skipEmptyLines: true,
            complete: (results) => {
                setHeaders(Object?.keys(results?.data[0]));
                setData(results?.data);
            },
        });
    };

    // Process Excel Files
    const parseExcel = (file) => {
        const reader = new FileReader();
        reader.onload = (e) => {
            const binaryStr = e?.target?.result;
            const workbook = XLSX.read(binaryStr, { type: 'binary' });
            const sheetName = workbook.SheetNames[0];
            const sheet = workbook.Sheets[sheetName];
            const json = XLSX.utils.sheet_to_json(sheet);

            if (json.length > 0) {
                setHeaders(Object?.keys(json[0]));
                setData(json);
            }
        };
        reader.readAsBinaryString(file);
    };

    const handleDataUpload = () => {
        console.log("Data to send to Backend:", data);
    };

    return (
        <>
            <FormItem className={`uploader-form-item`}>
                <p className="cds--file--label">
                    UPLOAD A FILE & PREVIEW DATA
                </p>
                <p className="cds--label-description">
                    Supported file types are .csv .xlsx and .xls.
                </p>

                {showUploader && (<FileUploaderDropContainer
                    accept={['.csv','.xlsx','.xls']}
                    labelText="Drag and drop a file here or click to upload"
                    name=""
                    onAddFiles={handleFileUpload}
                    onChange={handleFileUpload}
                    tabIndex={0}
                />)}

                {!showUploader && (<FileUploaderItem
                    iconDescription="Delete file"
                    name={fileName}
                    onDelete={handleFileDelete}
                    size="lg"
                    status={uploadStatus}
                />)}
            </FormItem>

            {data.length > 0 && (
                <div className={`preview-container`}>
                    <p className="cds--file--label">
                        PREVIEW: {fileName}
                    </p>
                    <div style={{ overflowX: 'auto', maxHeight: '400px', border: '1px solid #ddd' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <thead style={{ backgroundColor: '#f4f4f4', position: 'sticky', top: 0 }}>
                            <tr>
                                {headers.map((h) => (
                                    <th key={h} style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left' }}>{h}</th>
                                ))}
                            </tr>
                            </thead>
                            <tbody>
                            {data.slice(0, 10).map((row, i) => (
                                <tr key={i}>
                                    {headers.map((h) => (
                                        <td key={h} style={{ border: '1px solid #ddd', padding: '8px' }}>{row[h]}</td>
                                    ))}
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </div>
                    <p><i>Showing first 10 rows of {data.length} total rows.</i></p>


                    <div className={`file-upload-btn-container`}>
                        <Button size={`md`} kind="tertiary" onClick={handleDataUpload}>Upload Data</Button>
                        <Button size={`md`} kind="secondary" onClick={handleFileDelete}>Cancel</Button>
                    </div>
                </div>
            )}
        </>
    );
};

export default FileUpload;