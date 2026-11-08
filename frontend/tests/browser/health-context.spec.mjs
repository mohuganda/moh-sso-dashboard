/* global window */
import { URL } from "node:url";
import { test, expect } from "@playwright/test";

for (const width of [1440, 390]) {
  test(`health context survives the updated shell and scopes refreshed requests at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 900 });
    const national = "00000000-0000-0000-0000-000000000001";
    const district = "00000000-0000-0000-0000-000000000002";
    const foreign = "00000000-0000-0000-0000-000000000099";
    let selected = national;
    const headers = [];
    const contexts = () => [
      {
        id: national,
        name: "Uganda",
        contextType: "NATIONAL",
        isDefault: true,
        isActive: selected === national,
      },
      {
        id: district,
        name: "Test district",
        contextType: "DISTRICT",
        isActive: selected === district,
      },
    ];
    await page.addInitScript(
      (foreign) => window.localStorage.setItem("moh.activeHealthContextId", foreign),
      foreign,
    );
    await page.route("http://127.0.0.1:5417/ssobackend/**", async (route) => {
      const path = new URL(route.request().url()).pathname;
      const header = route.request().headers()["x-health-context-id"];
      if (header) headers.push(header);
      let data = [];
      if (path.endsWith("/auth/me"))
        data = {
          user: {
            id: "fixture-admin",
            username: "admin",
            isAdmin: true,
            enabled: true,
            realmRoles: ["admin"],
            clientRoles: {},
            permissions: ["portal:access", "clients:read"],
            systems: [],
            accessibleSystems: [],
            healthContexts: contexts(),
            activeHealthContext: contexts().find((context) => context.id === selected),
          },
        };
      else if (path.endsWith("/me/health-contexts/select")) {
        const id = route.request().postDataJSON().contextId;
        if (id === national)
          return route.fulfill({
            status: 503,
            json: { success: false, error: { message: "Unavailable" } },
          });
        selected = id;
        data = { contextId: selected };
      } else if (path.endsWith("/me/health-contexts")) data = { healthContexts: contexts() };
      else if (path.includes("unread")) data = { count: 0 };
      return route.fulfill({ json: { success: true, data } });
    });
    await page.goto("/portal/admin/clients");
    const selector = page.getByLabel("Health context", { exact: true });
    await expect(selector).toHaveValue(national);
    const refreshed = page.waitForRequest(
      (request) =>
        new URL(request.url()).pathname.endsWith("/clients") &&
        request.headers()["x-health-context-id"] === district,
    );
    await selector.selectOption(district);
    await refreshed;
    await expect(selector).toHaveValue(district);
    await selector.selectOption(national);
    await expect(page.getByText("Context change failed", { exact: true })).toBeVisible();
    await expect(selector).toHaveValue(district);
    expect(headers).not.toContain(foreign);
    expect(
      await page.evaluate(() => window.localStorage.getItem("moh.activeHealthContextId")),
    ).toBe(district);
    const bounds = await selector.boundingBox();
    expect(bounds.x + bounds.width).toBeLessThanOrEqual(width);
  });
}
