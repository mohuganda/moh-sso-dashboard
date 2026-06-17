import type { System } from "./auth.types";

export const SYSTEMS = {
  dashboardWeb: "dashboard-web",
  integratedOutbreak: "integrated-outbreak-system",
  reportBrowser: "report-browser",
} as const satisfies Record<string, System>;

export const SYSTEM_ROLES = {
  dashboardWeb: {
    access: "dashboard-web_access",
    admin: "portal_admin",
    manager: "portal_manager",
    user: "portal_user",
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
    admin: "report_admin",
    manager: "report_manager",
    analyst: "report_analyst",
    viewer: "report_viewer",
  },
} as const;
