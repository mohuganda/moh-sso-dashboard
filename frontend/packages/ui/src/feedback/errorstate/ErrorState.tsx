import { Warning } from "@carbon/react/icons";
import { Button, Tile } from "@carbon/react";

import "./ErrorState.scss";

interface ErrorStateProps {
  title?: string;
  description: string;
  primaryAction?: {
    label: string;
    onClick: () => void;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
}

export function ErrorState({
  title = "Something went wrong",
  description,
  primaryAction,
  secondaryAction,
}: ErrorStateProps) {
  return (
    <Tile className="moh-error-state">
      <Warning size={32} className="moh-error-state__icon" />

      <h4 className="moh-error-state__title">{title}</h4>

      <p className="moh-error-state__description">{description}</p>

      <div className="moh-error-state__actions">
        {primaryAction && (
          <Button kind="primary" onClick={primaryAction.onClick}>
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
