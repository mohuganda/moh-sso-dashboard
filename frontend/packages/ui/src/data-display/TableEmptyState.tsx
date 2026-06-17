import type { ReactNode } from "react";

import "./TableEmptyState.scss";

type TableEmptyStateProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
};

export function TableEmptyState({ title, description, actions }: TableEmptyStateProps) {
  return (
    <div className="moh-table-empty-state">
      <h4 className="moh-table-empty-state__title">{title}</h4>
      {description && <p className="moh-table-empty-state__description">{description}</p>}
      {actions && <div className="moh-table-empty-state__actions">{actions}</div>}
    </div>
  );
}
