import type { GlobalProvider } from "@ladle/react";

import { MohThemeProvider } from "../src/theme";

import "../src/theme/index";

export const Provider: GlobalProvider = ({ children }) => {
  return <MohThemeProvider theme="white">{children}</MohThemeProvider>;
};
