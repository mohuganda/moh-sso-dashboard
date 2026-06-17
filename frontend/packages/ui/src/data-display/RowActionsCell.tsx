import { TableCell } from "@carbon/react";
import type { ReactNode } from "react";

import "./RowActionsCell.scss";

type RowActionsCellProps = {
  children: ReactNode;
  align?: "start" | "end";
};

export function RowActionsCell({ children, align = "end" }: RowActionsCellProps) {
  return (
    <TableCell>
      <div className={`moh-row-actions-cell moh-row-actions-cell--${align}`}>{children}</div>
    </TableCell>
  );
}
