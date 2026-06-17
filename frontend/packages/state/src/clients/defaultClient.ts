import type { Client } from "@moh-sso/types";

export const DEFAULT_CLIENT_ID = "__default__";
export const UTILITIES_CLIENT_ID = "__utilities__";
export const SETTINGS_CLIENT_ID = "__settings__";

export const defaultClient: Client = {
  id: DEFAULT_CLIENT_ID,
  clientId: DEFAULT_CLIENT_ID,
  name: "Data & Statistics",
  enabled: true,
  publicClient: false,

  // Redirects are handled centrally (auth callback + router)
  redirectUris: [],

  attributes: {
    // UI metadata
    "ui.icon": "home",

    // User landing entry (NOT a deep link)
    "ui.home": "/apps/dwh",

    // UserLayout sidenav
    "ui.sidenav": JSON.stringify([
      {
        id: "data-visualizer",
        label: "Data Visualizer",
        path: "/apps/dwh/data-visualizer",
        permission: "data_quality:read",
      },
      {
        id: "dashboards",
        label: "Dashboards",
        path: "/apps/dwh/dashboards",
        permission: "data_quality:read",
      },
      {
        id: "reports",
        label: "Reports",
        path: "/apps/dwh/reports",
        permission: "report_browser:read",
      },
      {
        id: "data-exports",
        label: "Data Exports",
        path: "/apps/dwh/exports",
        permission: "data_quality:read",
      },
      {
        id: "file-svr",
        label: "File Upload",
        path: "/apps/dwh/filesvr",
        permission: "documents:read",
      },
      {
        id: "surveillance",
        label: "Surveillance",
        path: "/apps/dwh/surveillance",
        permission: "surveillance:read",
      },
      {
        id: "issue-tracker",
        label: "Issue Tracking",
        path: "/apps/dwh/issue-tracker",
        permission: "data_quality:read",
      },
    ]),
  },
};

export const utilitiesClient: Client = {
  id: UTILITIES_CLIENT_ID,
  clientId: UTILITIES_CLIENT_ID,
  name: "Utilities",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "tools",
    "ui.home": "/apps/utilities",

    "ui.sidenav": JSON.stringify([
      {
        id: "self-service",
        label: "Self Service",
        children: [
          {
            id: "my-timesheet",
            label: "My Timesheet",
            path: "/apps/utilities/self-service/timesheet",
          },
          {
            id: "elearning",
            label: "eLearning",
            path: "/apps/utilities/self-service/elearning",
          },
          {
            id: "leave-plan",
            label: "Leave Plan",
            path: "/apps/utilities/self-service/leave-plan",
          },
          {
            id: "absence-requests",
            label: "Absence Requests",
            path: "/apps/utilities/self-service/absence-requests",
          },
          {
            id: "absence-dashboard",
            label: "My Absence Dashboard",
            path: "/apps/utilities/self-service/absence-dashboard",
          },
          {
            id: "eservice-requests",
            label: "eService Requests",
            children: [
              {
                id: "document-upload",
                label: "Document Upload",
                path: "/apps/utilities/self-service/eservice/document-upload",
              },
              {
                id: "service-access",
                label: "Service Access",
                path: "/apps/utilities/self-service/eservice/service-access",
              },
              {
                id: "equipment-request",
                label: "Equipment Request",
                path: "/apps/utilities/self-service/eservice/equipment-request",
              },
            ],
          },
        ],
      },
    ]),
  },
};

export const settingsClient: Client = {
  id: SETTINGS_CLIENT_ID,
  clientId: SETTINGS_CLIENT_ID,
  name: "Settings",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "settings",
    "ui.home": "/apps/settings",

    "ui.sidenav": JSON.stringify([
      {
        id: "profile",
        label: "My Profile",
        path: "/apps/settings/profile",
      },
      {
        id: "sessions",
        label: "Active Sessions",
        path: "/apps/settings/sessions",
      },
      {
        id: "security",
        label: "Security",
        path: "/apps/settings/security",
      },
    ]),
  },
};

export const caseRegistersClient: Client = {
  id: "case-registers",
  clientId: "case-registers",
  name: "Case Registers",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "document",
    "ui.home": "/apps/case-registers",

    "ui.sidenav": JSON.stringify([
      {
        id: "external-referrals",
        label: "External Referrals",
        path: "/apps/case-registers/external-referrals",
      },
      {
        id: "disease-registers",
        label: "Disease Registers",
        path: "/apps/case-registers/disease-registers",
      },
    ]),
  },
};

export const outbreakManagementClient: Client = {
  id: "outbreak-management",
  clientId: "outbreak-management",
  name: "Outbreak Management",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "warning-alt",
    "ui.home": "/apps/outbreak-management",

    "ui.sidenav": JSON.stringify([
      {
        id: "signals-alerts",
        label: "Signals & Alerts",
        path: "/apps/outbreak-management/signals-alerts",
      },
      {
        id: "poe-management",
        label: "PoE Management",
        path: "/apps/outbreak-management/poe-management",
      },
      {
        id: "case-management",
        label: "Case Management",
        path: "/apps/outbreak-management/case-management",
      },
    ]),
  },
};

export const referenceRegistersClient: Client = {
  id: "reference-registers",
  clientId: "reference-registers",
  name: "Reference Registers",
  enabled: true,
  publicClient: false,
  redirectUris: [],
  attributes: {
    "ui.icon": "catalog",
    "ui.home": "/apps/reference-registers",

    "ui.sidenav": JSON.stringify([
      {
        id: "facility-register",
        label: "Facility Register",
        path: "/apps/reference-registers/facility-register",
      },
      {
        id: "terminology-service",
        label: "Terminology Service",
        children: [
          {
            id: "test-menu",
            label: "Test Menu",
            path: "/apps/reference-registers/terminology/test-menu",
          },
          {
            id: "pharmaceuticals",
            label: "Pharmaceuticals",
            path: "/apps/reference-registers/terminology/pharmaceuticals",
          },
          {
            id: "procedures",
            label: "Procedures",
            path: "/apps/reference-registers/terminology/procedures",
          },
          {
            id: "equipment",
            label: "Equipment",
            path: "/apps/reference-registers/terminology/equipment",
          },
        ],
      },
    ]),
  },
};

export const eServicesClient: Client = {
  id: "eservices",
  clientId: "eservices",
  name: "eServices",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "application",
    "ui.home": "/apps/eservices",

    "ui.sidenav": JSON.stringify([
      {
        id: "ihris",
        label: "iHRIS",
        path: "/apps/eservices/ihris",
      },
      {
        id: "meeting-manager",
        label: "Meeting Manager",
        path: "/apps/eservices/meeting-manager",
      },
      {
        id: "action-tracker",
        label: "Action Tracker",
        path: "/apps/eservices/action-tracker",
      },
      {
        id: "clinician-outputs",
        label: "Clinician Outputs",
        path: "/apps/eservices/clinician-outputs",
      },
      {
        id: "leave-absence",
        label: "Leave & Absence Management",
        path: "/apps/eservices/leave-absence",
      },
      {
        id: "workplans",
        label: "Workplans",
        path: "/apps/eservices/workplans",
      },
      {
        id: "budget-tracker",
        label: "Budget Tracker",
        path: "/apps/eservices/budget-tracker",
      },
      {
        id: "activity-reporting",
        label: "Activity Reporting",
        path: "/apps/eservices/activity-reporting",
      },
      {
        id: "partner-management",
        label: "Partner Management",
        path: "/apps/eservices/partner-management",
      },
      {
        id: "observatory-uploads",
        label: "Observatory Uploads",
        path: "/apps/eservices/observatory-uploads",
      },
    ]),
  },
};

export const researchStudiesClient: Client = {
  id: "research-studies",
  clientId: "research-studies",
  name: "Research & Studies",
  enabled: true,
  publicClient: false,

  redirectUris: [],

  attributes: {
    "ui.icon": "microscope",
    "ui.home": "/apps/research-studies",

    "ui.sidenav": JSON.stringify([
      {
        id: "studies",
        label: "Studies",
        path: "/apps/research-studies/studies",
      },
      {
        id: "datasets",
        label: "Datasets",
        path: "/apps/research-studies/datasets",
      },
      {
        id: "ethics-approvals",
        label: "Ethics & Approvals",
        path: "/apps/research-studies/ethics",
      },
      {
        id: "publications",
        label: "Publications",
        path: "/apps/research-studies/publications",
      },
    ]),
  },
};

export const DEFAULT_CLIENTS: Client[] = [
  defaultClient,
  eServicesClient,
  researchStudiesClient,
  caseRegistersClient,
  outbreakManagementClient,
  referenceRegistersClient,
  utilitiesClient,
  settingsClient,
];
