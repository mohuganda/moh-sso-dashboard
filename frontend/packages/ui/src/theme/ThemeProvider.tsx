import type { ReactNode } from "react";
import { Theme } from "@carbon/react";

import "./moh-theme.scss";

type MohTheme = "white" | "g10" | "g90" | "g100";

type MohThemeProviderProps = {
  children: ReactNode;
  theme?: MohTheme;
};

export function MohThemeProvider({ children, theme = "white" }: MohThemeProviderProps) {
  return (
    <Theme theme={theme}>
      <div className="moh-theme-root">{children}</div>
    </Theme>
  );
}
