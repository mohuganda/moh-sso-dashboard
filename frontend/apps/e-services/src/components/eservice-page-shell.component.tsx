import { Button, Column, Grid, Stack, Tile } from "@carbon/react";
import { Add } from "@carbon/react/icons";
import type { ReactNode } from "react";

type EservicePageShellProps = {
  title: string;
  description: string;
  actionLabel?: string;
  onActionClick?: () => void;
  children?: ReactNode;
};

export function EservicePageShell({
  title,
  description,
  actionLabel,
  onActionClick,
  children,
}: EservicePageShellProps) {
  return (
    <Grid fullWidth style={{ padding: 16 }}>
      <Column lg={16} md={8} sm={4}>
        <Stack gap={5}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              gap: 16,
              alignItems: "flex-start",
            }}
          >
            <div>
              <h3 style={{ margin: 0 }}>{title}</h3>
              <p style={{ marginTop: 6, opacity: 0.8 }}>{description}</p>
            </div>

            {actionLabel && (
              <Button renderIcon={Add} onClick={onActionClick}>
                {actionLabel}
              </Button>
            )}
          </div>

          <Tile>{children}</Tile>
        </Stack>
      </Column>
    </Grid>
  );
}
