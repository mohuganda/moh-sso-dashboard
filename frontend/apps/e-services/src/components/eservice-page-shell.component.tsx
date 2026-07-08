import { Button, Column, Grid, Stack, Tile } from "@carbon/react";
import { Add } from "@carbon/react/icons";
import type { ReactNode } from "react";

import "./eservice-components.scss";

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
    <Grid fullWidth className="eservice-page-shell">
      <Column lg={16} md={8} sm={4}>
        <Stack gap={5}>
          <div className="eservice-page-shell__header">
            <div>
              <h3 className="eservice-page-shell__title">{title}</h3>
              <p className="eservice-page-shell__description">{description}</p>
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
