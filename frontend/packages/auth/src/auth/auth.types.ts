export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  firstName?: string;
  lastName?: string;
  fullName?: string;
  isAdmin: boolean;
  isUser: boolean;
  realmRoles: string[];
  clientRoles: Record<string, string[]>;
  permissions: Permission[];
  systems: System[];
  accessibleSystems: SystemAccess[];
  enabled: boolean;
  emailVerified: boolean;
  requirePwdChange: boolean;
  lastLoginAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface SystemAccess {
  clientId: string;
  displayName: string;
  launchUrl?: string;
  icon?: string;
  category?: string;
  roles: string[];
}

export type Role = "admin" | "user" | "manager";

export type System = "dashboard-web" | "integrated-outbreak-system" | "report-browser";

export type Permission =
  | "portal:access"
  | "systems:read"
  | "systems:launch"
  | "users:read"
  | "users:write"
  | "users:roles:write"
  | "clients:read"
  | "clients:write"
  | "clients:roles:write"
  | "announcements:read"
  | "announcements:write"
  | "announcements:publish"
  | "documents:read"
  | "documents:write"
  | "documents:process"
  | "document_templates:read"
  | "document_templates:write"
  | "document_templates:publish"
  | "surveillance:read"
  | "surveillance:import"
  | "surveillance:manage_locations"
  | "surveillance:manage_alerts"
  | "email:read"
  | "email:send"
  | "email:manage"
  | "storage_locations:read"
  | "storage_locations:write"
  | "data_quality:read"
  | "data_quality:write"
  | "data_quality:resolve"
  | "report_browser:read"
  | "outbreak:access"
  | "outbreak:manage"
  | "metrics:read"
  | "audit:read"
  | "notifications:read"
  | "notifications:write";
