type RuntimeAppConfig = {
  API_BASE_URL?: string;
};

function getRuntimeApiBaseUrl() {
  if (typeof globalThis === "undefined" || !("window" in globalThis)) {
    return undefined;
  }

  return (globalThis.window as Window & { __APP_CONFIG__?: RuntimeAppConfig }).__APP_CONFIG__
    ?.API_BASE_URL;
}

const API_BASE_URL =
  getRuntimeApiBaseUrl() ||
  import.meta.env.VITE_API_BASE_URL ||
  "https://dashboards.health.go.ug/ssobackend";

const API_ROOT = `${API_BASE_URL}/api`;
const API_VERSION = "v1";
const API_BASE = `${API_ROOT}/${API_VERSION}`;

export const API = {
  root: API_ROOT,
  version: API_VERSION,
  base: API_BASE,

  // --------------------------------------------------
  // Auth
  // --------------------------------------------------
  auth: {
    base: `${API_BASE}/auth`,
    login: () => `${API_BASE}/auth/login`,
    callback: () => `${API_BASE}/auth/callback`,
    refresh: () => `${API_BASE}/auth/refresh`,
    logout: () => `${API_BASE}/auth/logout`,
    me: () => `${API_BASE}/auth/me`,
  },

  // --------------------------------------------------
  // Clients
  // --------------------------------------------------
  clients: {
    base: `${API_BASE}/clients`,
    list: () => `${API_BASE}/clients`,
    byId: (id: string) => `${API_BASE}/clients/${id}`,
    create: () => `${API_BASE}/clients`,
    delete: (id: string) => `${API_BASE}/clients/${id}`,
    update: (id: string) => `${API_BASE}/clients/${id}`,

    // -------- Client roles --------
    roles: {
      // READ (authenticated)
      list: (clientId: string) => `${API_BASE}/clients/${clientId}/roles`,

      // WRITE (admin)
      create: (clientId: string) => `${API_BASE}/admin/clients/${clientId}/roles`,
      delete: (clientId: string, role: string) =>
        `${API_BASE}/admin/clients/${clientId}/roles/${encodeURIComponent(role)}`,
    },
  },

  // --------------------------------------------------
  // Users (authenticated)
  // --------------------------------------------------
  users: {
    base: `${API_BASE}/users`,
    list: () => `${API_BASE}/users`,
    byId: (id: string) => `${API_BASE}/users/${id}`,
    create: () => `${API_BASE}/users`,
    delete: (id: string) => `${API_BASE}/users/${id}`,
    update: (id: string) => `${API_BASE}/users/${id}`,
    resetPassword: (id: string) => `${API_BASE}/users/${id}/reset-password`,
  },

  // --------------------------------------------------
  // Visualizer
  // --------------------------------------------------
  visualizer: {
    themes: () => `${API_BASE}/visualizer/themes`,
    theme: () => `${API_BASE}/visualizer/dataelements/theme`,
    hierarchy: () => `${API_BASE}/visualizer/adminunits/hierarchy`,
    dataValues: () => `${API_BASE}/visualizer/datavalues`,
    datasets: () => `${API_BASE}/visualizer/datasets`,
    dataElements: () => `${API_BASE}/visualizer/dataelements`
  },
  // --------------------------------------------------
  // Visualizer
  // --------------------------------------------------
  issue: {
    list: () => `${API_BASE}/issues`,
  },

  // --------------------------------------------------
  // Admin
  // --------------------------------------------------
  admin: {
    base: `${API_BASE}/admin`,

    // -------- Users --------
    users: {
      list: () => `${API_BASE}/admin/users`,
      byId: (id: string) => `${API_BASE}/admin/users/${id}`,
      create: () => `${API_BASE}/admin/users`,
      update: (id: string) => `${API_BASE}/admin/users/${id}`,
      delete: (id: string) => `${API_BASE}/admin/users/${id}`,
      toggle: (id: string) => `${API_BASE}/admin/users/${id}/toggle`,
      resetPassword: (id: string) => `${API_BASE}/admin/users/${id}/reset-password`,
      passwordResetEmail: (id: string) => `${API_BASE}/admin/users/${id}/password-reset`,

      import: {
        preview: () => `${API_BASE}/admin/users/import/preview`,
        execute: () => `${API_BASE}/admin/users/import/execute`,
        job: (jobId: string) => `${API_BASE}/admin/users/import/${jobId}`,
        errorsCsv: (jobId: string) => `${API_BASE}/admin/users/import/${jobId}/errors.csv`,
        templateCsv: () => `${API_BASE}/admin/users/import/template.csv`,
      },

      // -------- User ↔ Client roles --------
      clientRoles: {
        // GET roles user has for a client
        list: (userId: string, clientId: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles`,

        // PUT diff-based update
        update: (userId: string) => `${API_BASE}/admin/users/${userId}/client-roles`,

        // POST single role
        assign: (userId: string, clientId: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles`,

        // DELETE single role
        remove: (userId: string, clientId: string, role: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles/${role}`,
      },
    },

    // -------- Metrics --------
    metrics: {
      overview: () => `${API_BASE}/admin/metrics/overview`,

      system: {
        countUsers: () => `${API_BASE}/admin/metrics/system/count-users`,
        countDisabledUsers: () => `${API_BASE}/admin/metrics/system/count-disabled-users`,
        activeToday: () => `${API_BASE}/admin/metrics/system/active-today`,
        activeThisWeek: () => `${API_BASE}/admin/metrics/system/active-this-week`,
        loginTrend: () => `${API_BASE}/admin/metrics/system/login-trend`,
        loginTrendRange: () => `${API_BASE}/admin/metrics/system/login-trend-range`,
      },

      security: {
        failedLogins: () => `${API_BASE}/admin/metrics/security/failed-logins`,
        failedLoginsRange: () => `${API_BASE}/admin/metrics/security/failed-logins-range`,
        suspiciousLogins: () => `${API_BASE}/admin/metrics/security/suspicious-logins`,
      },

      clients: {
        count: () => `${API_BASE}/admin/metrics/clients/count`,
        mostAccessed: () => `${API_BASE}/admin/metrics/clients/most-accessed`,
        loginCount: () => `${API_BASE}/admin/metrics/clients/login-count`,
        activeToday: () => `${API_BASE}/admin/metrics/clients/active-today`,
      },

      users: {
        newRange: () => `${API_BASE}/admin/metrics/users/new-range`,
        newTrend: () => `${API_BASE}/admin/metrics/users/new-trend`,
        neverLoggedIn: () => `${API_BASE}/admin/metrics/users/never-logged-in`,
        lastLogin: (userId: string) => `${API_BASE}/admin/metrics/users/last-login/${userId}`,
        clientUsage: (userId: string) => `${API_BASE}/admin/metrics/users/client-usage/${userId}`,
      },
    },

    // -------- Audit --------
    audit: {
      base: `${API_BASE}/admin/audit-logs`,
      list: () => `${API_BASE}/admin/audit-logs`,
      actions: () => `${API_BASE}/admin/audit-logs/actions`,
      byId: (id: string) => `${API_BASE}/admin/audit-logs/${id}`,

      metrics: {
        overview: () => `${API_BASE}/admin/audit-logs/metrics/overview`,
        failedLoginsByDay: () => `${API_BASE}/admin/audit-logs/metrics/failed-logins-by-day`,
        topFailureIps: () => `${API_BASE}/admin/audit-logs/metrics/top-failure-ips`,
      },

      export: () => `${API_BASE}/admin/audit-logs/export`,
    },

    // -------- Notifications --------
    notifications: {
      base: `${API_BASE}/admin/notifications`,
      notify: () => `${API_BASE}/admin/notifications`,
      list: () => `${API_BASE}/admin/notifications`,
      byId: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      markAsRead: (id: string) => `${API_BASE}/admin/notifications/${id}/read`,
      delete: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      count: () => `${API_BASE}/admin/notifications/count`,
      countUnread: () => `${API_BASE}/admin/notifications/count/unread`,
      deleteOld: () => `${API_BASE}/admin/notifications/cleanup`,
      deliveries: (id: string) => `${API_BASE}/admin/notifications/${id}/deliveries`,
      retryDelivery: (deliveryId: string) =>
        `${API_BASE}/admin/notifications/deliveries/${deliveryId}/retry`,
    },
  },

  // --------------------------------------------------
  // Health (not versioned)
  // --------------------------------------------------
  health: () => `/health`,
} as const;
