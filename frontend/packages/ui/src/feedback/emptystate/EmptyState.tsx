import { Information } from "@carbon/react/icons";
import { Button, Tile } from "@carbon/react";

import "./EmptyState.scss";

interface EmptyStateProps {
  title: string;
  description: string;
  primaryAction?: {
    label: string;
    onClick: () => void;
    icon?: React.ElementType;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
}

export function EmptyState({
  title,
  description,
  primaryAction,
  secondaryAction,
}: EmptyStateProps) {
  return (
    <Tile className="moh-empty-state">
      <Information size={32} className="moh-empty-state__icon" />

      <h4 className="moh-empty-state__title">{title}</h4>
      <p className="moh-empty-state__description">{description}</p>

      <div className="moh-empty-state__actions">
        {primaryAction && (
          <Button kind="primary" renderIcon={primaryAction.icon} onClick={primaryAction.onClick}>
            {primaryAction.label}
          </Button>
        )}

        {secondaryAction && (
          <Button kind="tertiary" onClick={secondaryAction.onClick}>
            {secondaryAction.label}
          </Button>
        )}
      </div>
    </Tile>
  );
}
