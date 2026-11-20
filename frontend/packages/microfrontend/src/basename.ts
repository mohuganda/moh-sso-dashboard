function normalizePath(path: string) {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return normalized.replace(/\/+$/, "") || "/";
}

function containsPath(pathname: string, basename: string) {
  return pathname === basename || pathname.startsWith(`${basename}/`);
}

export function resolveRuntimeBasename(basename: string, deploymentBase?: string) {
  const normalizedBasename = normalizePath(basename);
  const deployment = deploymentBase ? normalizePath(deploymentBase) : "";

  if (deployment && deployment !== "/") {
    if (normalizedBasename === "/") {
      return deployment;
    }
    return containsPath(normalizedBasename, deployment)
      ? normalizedBasename
      : `${deployment}${normalizedBasename}`;
  }

  if (typeof window === "undefined" || normalizedBasename === "/") {
    return normalizedBasename;
  }

  const pathname = normalizePath(window.location.pathname);
  // The leading slash already establishes the start of a path segment.
  let index = pathname.indexOf(normalizedBasename);
  while (index !== -1) {
    const end = index + normalizedBasename.length;
    if (end === pathname.length || pathname[end] === "/") {
      return pathname.slice(0, end);
    }
    index = pathname.indexOf(normalizedBasename, index + 1);
  }

  return normalizedBasename;
}
