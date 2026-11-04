import { test, expect } from "@playwright/test";
import { URL } from "node:url";

const rootPages = [
  ["/admin/users", "Users"],
  ["/admin/clients", "Clients"],
  ["/admin/audit-logs", "Audit Logs"],
  ["/admin/announcements", "Announcements"],
  ["/admin/emails", "Email Outbox"],
  ["/admin/rbac", "RBAC Management"],
  ["/apps/dwh/data-visualizer", "Data Visualizer"],
  ["/apps/dwh/data-validation", "Data Validation"],
  ["/apps/dwh/documents", "Document Management"],
  ["/apps/dwh/issue-tracker", "Issue Tracker"],
  ["/apps/dwh/dashboards", "Report Browser"],
  ["/apps/dwh/report-scheduler", "Report Scheduler"],
  ["/apps/dwh/surveillance", "Surveillance Dashboard"],
  ["/apps/eservices/ihris/meeting-manager", "Meeting Manager"],
  ["/apps/utilities/self-service/timesheet", "Timesheet"],
];

const user = {
  id: "00000000-0000-0000-0000-000000000001",
  username: "root-layout-admin",
  isAdmin: true,
  isUser: true,
  enabled: true,
  realmRoles: ["admin"],
  clientRoles: {},
  permissions: [
    "portal:access",
    "systems:read",
    "systems:launch",
    "users:read",
    "clients:read",
    "audit:read",
    "announcements:read",
    "email:read",
    "rbac:read",
    "data_quality:read",
    "documents:read",
    "document_templates:read",
    "issue_tracker:read",
    "report_browser:read",
    "report_scheduler:read",
    "surveillance:read",
  ],
  systems: ["data-statistics", "eservices", "utilities"],
  accessibleSystems: [],
  emailVerified: true,
  requirePwdChange: false,
};

for (const width of [1440, 390]) {
  for (const [path, title] of rootPages) {
    test(`${title} has one aligned root header at ${width}px`, async ({ page }) => {
      await page.setViewportSize({ width, height: 900 });
      const failures = [];
      page.on("pageerror", (error) => failures.push(error.message));
      await page.route("http://127.0.0.1:5417/ssobackend/**", async (route) => {
        const path = new URL(route.request().url()).pathname;
        const data = path.endsWith("/auth/me")
          ? { user }
          : path.endsWith("/admin/audit-logs/actions")
            ? { actions: [] }
            : path.endsWith("/admin/audit-logs")
              ? { items: [], has_more: false, next_cursor: {} }
              : path.endsWith("/admin/rbac/drift")
                ? {
                    source: "fixture",
                    systems: [],
                    realmRoles: [],
                    warnings: [],
                    summary: {
                      systemsInSync: 0,
                      systemsMissingInRbac: 0,
                      systemsStaleInRbac: 0,
                      rolesMissingInRbac: 0,
                      rolesStaleInRbac: 0,
                      realmRolesMissing: 0,
                    },
                  }
                : path.includes("unread")
                  ? { count: 0 }
                  : path.includes("geojson")
                    ? { type: "FeatureCollection", features: [] }
                    : [];
        await route.fulfill({ json: { success: true, data } });
      });
      // The portal header is under test; the separately hosted report app is not.
      await page.route("https://dashboards.health.go.ug/**", (route) =>
        route.fulfill({ contentType: "text/html", body: "<p>Report browser fixture</p>" }),
      );

      await page.goto(`/portal${path}`);
      const main = page.locator("#main-content");
      const header = main.locator(".moh-page-shell__header");
      await expect(header).toHaveCount(1);
      const heading = header.getByRole("heading", { name: title, exact: true });
      await expect(heading).toBeVisible();
      await expect(main.getByRole("heading", { name: title, exact: true })).toHaveCount(1);
      const containerBounds = await main.boundingBox();
      const headingBounds = await heading.boundingBox();
      const padding = await main.evaluate((element) =>
        parseFloat(element.ownerDocument.defaultView.getComputedStyle(element).paddingLeft),
      );
      expect(Math.abs(headingBounds.x - containerBounds.x - padding)).toBeLessThanOrEqual(2);
      expect(headingBounds.y - containerBounds.y).toBeGreaterThanOrEqual(0);
      expect(headingBounds.y - containerBounds.y).toBeLessThan(100);
      expect(headingBounds.x + headingBounds.width).toBeLessThanOrEqual(width);
      await expect(header.locator(".moh-page-shell__subtitle")).toBeVisible();
      expect(failures).toEqual([]);
    });
  }
}
