import { useEffect, useState } from "react";

import { API, APP_VERSION, BUILD_TIME } from "@moh-sso/config";

type VersionBuildInfo = {
  service?: string;
  version?: string;
  commit?: string;
  buildTime?: string;
};

type VersionEnvelope = {
  success?: boolean;
  data?: VersionBuildInfo;
};

export type VersionInfo = {
  frontend: VersionBuildInfo;
  backend?: VersionBuildInfo;
  isLoading: boolean;
  error?: string;
};

const frontendBuildInfo: VersionBuildInfo = {
  service: "moh-sso-dashboard-frontend",
  version: APP_VERSION,
  buildTime: BUILD_TIME || undefined,
};

let cachedBackendVersion: VersionBuildInfo | undefined;
let cachedError: string | undefined;
let backendVersionRequest: Promise<VersionBuildInfo | undefined> | undefined;

async function fetchBackendVersion() {
  const response = await fetch(API.build.version(), {
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Backend version unavailable");
  }

  const payload = (await response.json()) as VersionEnvelope;

  if (!payload.data?.version) {
    throw new Error("Backend version unavailable");
  }

  return payload.data;
}

function getBackendVersion() {
  if (cachedBackendVersion || cachedError) {
    return Promise.resolve(cachedBackendVersion);
  }

  backendVersionRequest ??= fetchBackendVersion()
    .then((buildInfo) => {
      cachedBackendVersion = buildInfo;
      return buildInfo;
    })
    .catch(() => {
      cachedError = "Backend version unavailable";
      return undefined;
    })
    .finally(() => {
      backendVersionRequest = undefined;
    });

  return backendVersionRequest;
}

export function useVersionInfo(): VersionInfo {
  const [versionInfo, setVersionInfo] = useState<VersionInfo>({
    frontend: frontendBuildInfo,
    backend: cachedBackendVersion,
    isLoading: !cachedBackendVersion && !cachedError,
    error: cachedError,
  });

  useEffect(() => {
    let isMounted = true;

    void getBackendVersion().then((backend) => {
      if (!isMounted) {
        return;
      }

      setVersionInfo({
        frontend: frontendBuildInfo,
        backend,
        isLoading: false,
        error: cachedError,
      });
    });

    return () => {
      isMounted = false;
    };
  }, []);

  return versionInfo;
}
