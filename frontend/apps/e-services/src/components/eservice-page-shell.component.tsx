import { PageShell } from "@moh-sso/ui";
import { Button, Tile } from "@carbon/react";
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
    <PageShell
      title={title}
      subtitle={description}
      actions={
        actionLabel ? (
          <Button renderIcon={Add} onClick={onActionClick}>
            {actionLabel}
          </Button>
        ) : undefined
      }
    >
      <Tile>{children}</Tile>
    </PageShell>
  );
}
