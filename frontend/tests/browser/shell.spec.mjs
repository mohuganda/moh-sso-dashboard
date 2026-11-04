import { test, expect } from "@playwright/test";
import { URL } from "node:url";

const navigation = JSON.stringify([
  {
    id: "validation",
    label: "Data Validation",
    path: "/apps/dwh/data-validation",
    permission: "data_quality:read",
  },
  {
    id: "visualizer",
    label: "Data Visualizer",
    path: "/apps/dwh/data-visualizer",
    permission: "documents:read",
  },
]);
const user = {
  id: "00000000-0000-0000-0000-000000000001",
  username: "shell-reviewer",
  firstName: "Portal",
  lastName: "Reviewer",
  email: "reviewer@example.invalid",
  isAdmin: true,
  isUser: true,
  enabled: true,
  realmRoles: ["admin"],
  clientRoles: {},
  permissions: [
    "portal:access",
    "systems:launch",
    "users:read",
    "clients:read",
    "data_quality:read",
    "documents:read",
  ],
  systems: ["data-statistics", "settings"],
  accessibleSystems: [
    {
      clientId: "data-statistics",
      displayName: "Data & Statistics",
      launchUrl: "/apps/dwh/data-validation",
      navigation,
      systemType: "platform",
      displayInLauncher: true,
      displayInSideNav: true,
      roles: [],
    },
    {
      clientId: "settings",
      displayName: "Settings",
      launchUrl: "/apps/settings/profile",
      systemType: "platform",
      displayInLauncher: true,
      displayInSideNav: false,
      roles: [],
    },
  ],
  emailVerified: true,
  requirePwdChange: false,
};

async function fixture(page, account = user) {
  const failures = [];
  page.on("pageerror", (error) => failures.push(error.message));
  await page.route("http://127.0.0.1:5417/ssobackend/**", (route) => {
    const path = new URL(route.request().url()).pathname;
    const data = path.endsWith("/auth/me")
      ? { user: account }
      : path.includes("unread")
        ? { count: 0 }
        : [];
    return route.fulfill({ json: { success: true, data } });
  });
  return failures;
}

for (const [route, navLabel] of [
  ["/admin/users", "Admin navigation"],
  ["/apps/dwh/data-validation", "Application navigation"],
]) {
  test(`mobile navigation closes on Escape and route changes: ${navLabel}`, async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    const failures = await fixture(page);
    await page.goto(`/portal${route}`);
    const toggle = page.getByRole("button", { name: "Open navigation", exact: true });
    await expect(toggle).toBeVisible();
    await expect(page.getByRole("navigation", { name: navLabel })).toHaveCount(0);
    await toggle.click();
    const nav = page.getByRole("navigation", { name: navLabel });
    await expect(nav).toBeVisible();
    await expect(page.locator("body")).toHaveCSS("overflow", "hidden");
    await page.keyboard.press("Escape");
    await expect(toggle).toHaveAttribute("aria-expanded", "false");
    await expect(toggle).toBeFocused();
    await expect(page.locator("body")).not.toHaveCSS("overflow", "hidden");
    await toggle.click();
    const destination = navLabel === "Admin navigation" ? "Clients" : "Data Visualizer";
    const link = nav.getByRole("link", { name: destination, exact: true });
    await expect(link).toHaveAttribute(
      "href",
      navLabel === "Admin navigation"
        ? "/portal/admin/clients"
        : "/portal/apps/dwh/data-visualizer",
    );
    await link.focus();
    await page.keyboard.press("Enter");
    await expect(toggle).toHaveAttribute("aria-expanded", "false");
    await expect(page.locator("body")).not.toHaveCSS("overflow", "hidden");
    expect(failures).toEqual([]);
  });

  test(`desktop navigation preference survives mobile use: ${navLabel}`, async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await fixture(page);
    await page.goto(`/portal${route}`);
    await expect(page.getByRole("navigation", { name: navLabel })).toBeVisible();
    await page.setViewportSize({ width: 390, height: 844 });
    const toggle = page.getByRole("button", { name: "Open navigation", exact: true });
    await expect(toggle).toBeVisible();
    await toggle.click();
    await expect(page.locator("body")).toHaveCSS("overflow", "hidden");
    await page.setViewportSize({ width: 1440, height: 900 });
    await expect(page.locator("body")).not.toHaveCSS("overflow", "hidden");
    await expect(page.getByRole("navigation", { name: navLabel })).toBeVisible();
    await page.getByRole("button", { name: "Collapse navigation", exact: true }).click();
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(toggle).toBeVisible();
    await page.setViewportSize({ width: 1440, height: 900 });
    await expect(page.getByRole("navigation", { name: navLabel })).toHaveCount(0);
    await expect(
      page.getByRole("button", { name: "Expand navigation", exact: true }),
    ).toBeVisible();
  });

  test(`launcher search and account panel work: ${navLabel}`, async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    const failures = await fixture(page);
    await page.goto(`/portal${route}`);
    const launcher = page.getByRole("button", { name: "Open application menu", exact: true });
    await launcher.click();
    const menu = page.getByRole("dialog", { name: "Application menu", exact: true });
    const search = menu.getByRole("searchbox");
    await expect(search).toBeFocused();
    await search.fill("statistics");
    await expect(menu.getByRole("listitem")).toHaveCount(1);
    await expect(menu.getByRole("status")).toContainText("1 application");
    await search.fill("no-matching-service");
    await expect(menu.getByText("No matching applications")).toBeVisible();
    await menu.getByRole("button", { name: "Clear search", exact: true }).click();
    await expect(search).toHaveValue("");
    await expect(search).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await expect(launcher).toBeFocused();
    const profile = page.getByRole("button", { name: "Signed in as shell-reviewer" });
    await profile.click();
    const account = page.getByRole("dialog", { name: "Your account", exact: true });
    await expect(
      account.getByRole("heading", { name: "Portal Reviewer", exact: true }),
    ).toBeVisible();
    await expect(account.getByRole("button", { name: "Profile and settings" })).toBeVisible();
    await expect(account.getByRole("button", { name: "Sign out", exact: true })).toBeVisible();
    await expect(account.getByRole("button", { name: "Maximize panel" })).toHaveCount(0);
    const bounds = await account.boundingBox();
    expect(bounds.x).toBeGreaterThanOrEqual(0);
    expect(bounds.width).toBeLessThanOrEqual(390);
    await page.keyboard.press("Escape");
    await expect(profile).toBeFocused();
    await expect(account).toHaveCount(0);
    for (const panel of await page.locator(".app-side-panel").all()) {
      await expect(panel).toHaveAttribute("inert", "");
    }
    expect(failures).toEqual([]);
  });
}

test("mobile navigation hands focus to the launcher", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await fixture(page);
  await page.goto("/portal/admin/users");
  await page.getByRole("button", { name: "Open navigation", exact: true }).click();
  await page.getByRole("button", { name: "Open application menu", exact: true }).click();
  await expect(page.getByRole("navigation", { name: "Admin navigation" })).toHaveCount(0);
  await expect(
    page.getByRole("dialog", { name: "Application menu" }).getByRole("searchbox"),
  ).toBeFocused();
});

test("public shell has a visible sign-in action on a narrow screen", async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 740 });
  await fixture(page, null);
  await page.goto("/portal/");
  const signIn = page.getByRole("button", { name: "Sign in", exact: true });
  await expect(signIn).toBeVisible();
  await expect(signIn).toContainText("Sign in");
  const box = await signIn.boundingBox();
  expect(box.x + box.width).toBeLessThanOrEqual(320);
  await expect(page.getByRole("heading", { name: "News & Updates" })).toBeVisible();
});
