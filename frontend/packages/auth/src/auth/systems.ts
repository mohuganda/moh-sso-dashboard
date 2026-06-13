import type { System } from "./auth.types";

export const SYSTEMS = {
  dashboardWeb: "dashboard-web",
  integratedOutbreak: "integrated-outbreak-system",
  reportBrowser: "report-browser",
} as const satisfies Record<string, System>;

export const SYSTEM_ROLES = {
  dashboardWeb: {
    access: "dashboard-web_access",
    integratedOutbreakAccess: "integrated-outbreak-system_access",
  },
  integratedOutbreak: {
    access: "integrated-outbreak-system_access",
    superAdmin: "super_admin",
    admin: "admin",
    manager: "manager",
    viewer: "viewer",
    dataEntry: "data_entry",
    labTechnician: "lab_technician",
    surveillanceOfficer: "surveillance_officer",
  },
  reportBrowser: {
    access: "report-browser_access",
  },
} as const;

