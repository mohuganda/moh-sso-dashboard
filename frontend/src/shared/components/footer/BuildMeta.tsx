// src/components/footer/BuildMeta.tsx
import { APP_VERSION, BUILD_TIME } from "@/config/appMeta";

export function BuildMeta() {
  return (
    <div style={{ fontSize: "0.75rem", opacity: 0.7 }}>
      v{APP_VERSION}
      {BUILD_TIME && (
        <>
          {" · "}
          built {new Date(BUILD_TIME).toLocaleString()}
        </>
      )}
    </div>
  );
}
