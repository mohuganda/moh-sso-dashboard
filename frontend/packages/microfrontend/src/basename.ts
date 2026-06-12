function normalizePath(path: string) {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return normalized.replace(/\/+$/, "") || "/";
}

function hasPathBoundary(pathname: string, start: number, length: number) {
  const before = start === 0 ? "" : pathname[start - 1];
  const after = pathname[start + length] ?? "";

  return (!before || before === "/") && (!after || after === "/" || after === "?" || after === "#");
}

export function resolveRuntimeBasename(basename: string, deploymentBase?: string) {
  const normalizedBasename = normalizePath(basename);
  const normalizedDeploymentBase = deploymentBase ? normalizePath(deploymentBase) : "";

  if (
    normalizedDeploymentBase &&
    normalizedDeploymentBase !== "/" &&
    !normalizedBasename.startsWith(`${normalizedDeploymentBase}/`) &&
    normalizedBasename !== normalizedDeploymentBase
  ) {
    return `${normalizedDeploymentBase}${normalizedBasename}`;
  }

  if (typeof window === "undefined") {
    return normalizedBasename;
  }

  const pathname = normalizePath(window.location.pathname);

  if (
    pathname === normalizedBasename ||
    pathname.startsWith(`${normalizedBasename}/`)
  ) {
    return normalizedBasename;
  }

  const basenameIndex = pathname.indexOf(normalizedBasename);

  if (basenameIndex > 0 && hasPathBoundary(pathname, basenameIndex, normalizedBasename.length)) {
    return `${pathname.slice(0, basenameIndex)}${normalizedBasename}`;
  }

  return normalizedBasename;
}
