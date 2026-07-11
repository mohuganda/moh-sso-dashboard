export type MicrofrontendRoute = {
  appName: string;
  path: string;
  paths?: string[];
  excludedPaths?: string[];
  requiredPermissions?: string[];
  requiredAnyPermissions?: string[];
  requiredSystems?: string[];
  requiredSystemRoles?: Array<{
    system: string;
    role: string;
  }>;
};
