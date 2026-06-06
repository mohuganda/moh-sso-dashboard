// src/layouts/BaseLayout.tsx
import { Content } from "@carbon/react";
import type { ReactNode } from "react";

type Props = {
  header?: ReactNode;
  footer?: ReactNode;
  children: ReactNode;
};

export function BaseLayout({ header, footer, children }: Props) {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
      }}
    >
      {header}

      <Content style={{ flex: 1 }}>{children}</Content>

      {footer}
    </div>
  );
}
