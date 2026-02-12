import { Information } from "@carbon/react/icons";
import { Button, Tile } from "@carbon/react";

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
    <Tile
      style={{
        padding: "3rem",
        textAlign: "center",
        maxWidth: 520,
        margin: "2rem auto",
      }}
    >
      <Information size={32} style={{ marginBottom: "1rem", opacity: 0.6 }} />

      <h4 style={{ marginBottom: "0.5rem" }}>{title}</h4>
      <p style={{ opacity: 0.8, marginBottom: "1.5rem" }}>{description}</p>

      <div style={{ display: "flex", gap: "0.75rem", justifyContent: "center" }}>
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
