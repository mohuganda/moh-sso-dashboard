import { Warning } from "@carbon/react/icons";
import { Button, Tile } from "@carbon/react";

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
    <Tile
      style={{
        padding: "3rem",
        textAlign: "center",
        maxWidth: 520,
        margin: "2rem auto",
      }}
    >
      <Warning size={32} style={{ marginBottom: "1rem", color: "#da1e28" }} />

      <h4 style={{ marginBottom: "0.5rem" }}>{title}</h4>

      <p style={{ opacity: 0.8, marginBottom: "1.5rem" }}>{description}</p>

      <div style={{ display: "flex", gap: "0.75rem", justifyContent: "center" }}>
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
