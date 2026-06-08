import { Button } from "@carbon/react";

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
    <div
      style={{
        minHeight: 280,
        display: "grid",
        placeItems: "center",
        textAlign: "center",
        padding: "2rem",
      }}
    >
      <div>
        <h4 style={{ margin: 0 }}>{title}</h4>
        <p
          style={{
            marginTop: 8,
            marginBottom: 20,
            opacity: 0.75,
            maxWidth: 520,
          }}
        >
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
