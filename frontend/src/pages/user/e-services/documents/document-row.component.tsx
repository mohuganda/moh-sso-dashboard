import {
  TableRow,
  TableCell,
  Tag,
  ProgressBar,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import { useNavigate } from "react-router-dom";
import { useGetDocumentProcessesQuery } from "../../../../store/api/document.api";

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

type Props = {
  document: any;
  getRowProps: any;
};

export function DocumentRow({ document, getRowProps }: Props) {
  const navigate = useNavigate();

  const { data: processes = [] } = useGetDocumentProcessesQuery(document.id, {
    pollingInterval: 3000,
  });

  const latest =
    processes.length > 0
      ? [...processes].sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        )[0]
      : undefined;

  const status = latest?.status ?? "PENDING";
  const progress = latest?.progress ?? 0;

  return (
    <TableRow {...getRowProps({ row: { id: document.id } })}>
      <TableCell>{document.original_filename}</TableCell>
      <TableCell>{document.content_type.String}</TableCell>
      <TableCell>{(document.size_bytes / 1024 / 1024).toFixed(2)} MB</TableCell>

      <TableCell>
        <StatusTag status={status} />
      </TableCell>

      <TableCell>
        <div style={{ width: 150 }}>
          <ProgressBar value={progress} max={100} label="" size="small" />
        </div>
      </TableCell>

      <TableCell>
        <OverflowMenu size="sm">
          <OverflowMenuItem
            itemText="View Details"
            onClick={() =>
              navigate(`/apps/utilities/self-service/eservice/document-upload/${document.id}`)
            }
          />
        </OverflowMenu>
      </TableCell>
    </TableRow>
  );
}
