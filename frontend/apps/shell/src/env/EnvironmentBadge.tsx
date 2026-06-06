// src/components/env/EnvironmentBadge.tsx
import { Tag } from "@carbon/react";

import { APP_ENV } from "@moh-sso/config";

const ENV_COLORS: Record<string, "red" | "purple" | "blue" | "gray"> = {
  development: "purple",
  dev: "purple",
  staging: "blue",
  production: "gray",
};

export function EnvironmentBadge() {
  if (APP_ENV === "production") return null;

  const type = ENV_COLORS[APP_ENV] ?? "gray";

  return (
    <Tag size="sm" type={type}>
      {APP_ENV.toUpperCase()}
    </Tag>
  );
}
