import { Button } from "@carbon/react";

import "./eservice-components.scss";

type EserviceEmptyStateProps = {
  title: string;
  description: string;
  actionLabel?: string;
  onActionClick?: () => void;
};

export function EserviceEmptyState({
  title,
  description,
  actionLabel,
  onActionClick,
}: EserviceEmptyStateProps) {
  return (
    <div className="eservice-empty-state">
      <div>
        <h4 className="eservice-empty-state__title">{title}</h4>
        <p className="eservice-empty-state__description">
          {description}
        </p>

        {actionLabel && (
          <Button kind="secondary" onClick={onActionClick}>
            {actionLabel}
          </Button>
        )}
      </div>
    </div>
  );
}
