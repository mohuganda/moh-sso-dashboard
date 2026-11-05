import { test, expect } from "@playwright/test";
import { URL } from "node:url";

const base = "/portal/apps/dwh/surveillance";
const weekId = "00000000-0000-0000-0000-000000000053";
const diseases = [
  { id: "measles", name: "Measles" },
  { id: "malaria", name: "Malaria" },
  { id: "paralysis", name: "Acute Flaccid Paralysis" },
];
const user = {
  id: "00000000-0000-0000-0000-000000000001",
  username: "navigation-reviewer",
  enabled: true,
  emailVerified: true,
  requirePwdChange: false,
  isAdmin: true,
  isUser: true,
  realmRoles: ["admin"],
  clientRoles: {},
  permissions: [
    "portal:access",
    "systems:read",
    "systems:launch",
    "surveillance:read",
    "documents:read",
    "document_templates:read",
  ],
  systems: ["data-statistics", "utilities"],
  accessibleSystems: [],
};

async function fixture(page) {
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.route("http://127.0.0.1:5417/ssobackend/**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    const data = path.endsWith("/auth/me")
      ? { user }
      : path.endsWith("/notifications/count/unread")
        ? { count: 0 }
        : path.endsWith("/surveillance/weeks")
          ? [{ id: weekId, epi_year: new Date().getFullYear(), epi_week: 1 }]
          : path.endsWith("/surveillance/diseases")
            ? diseases
            : path.includes("/weekly-statuses/")
              ? diseases.map((disease) => ({
                  id: disease.id,
                  disease_id: disease.id,
                  disease_name: disease.name,
                  epi_week_id: weekId,
                  status: "RED",
                }))
              : path.includes("geojson")
                ? { type: "FeatureCollection", features: [] }
                : [];
    await route.fulfill({ json: { success: true, data } });
  });
  return errors;
}

for (const width of [390, 1440]) {
  test(`disease tile and breadcrumb use one surveillance path at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 1000 });
    const errors = await fixture(page);
    await page.goto(base);
    await expect(
      page.getByRole("heading", { name: "Surveillance Dashboard", exact: true }),
    ).toBeVisible();
    for (const [label, slug] of [
      ["Measles", "measles"],
      ["Malaria", "malaria"],
      ["Acute Flaccid Paralysis", "acute-flaccid-paralysis"],
    ]) {
      await page
        .locator(".surveillance-tile")
        .getByRole("button", { name: label, exact: true })
        .click();
      await expect(page).toHaveURL(`${base}/${slug}?weekId=${weekId}`);
      await expect(
        page
          .locator(".disease-details-page__header")
          .getByRole("heading", { name: label, exact: true }),
      ).toBeVisible();
      const breadcrumb = page.getByRole("navigation", { name: "Breadcrumb", exact: true });
      await expect(breadcrumb).toHaveCount(1);
      await expect(breadcrumb.locator(".moh-breadcrumbs__current")).toHaveText(label);
      await expect(
        breadcrumb.getByRole("link", { name: "Surveillance", exact: true }),
      ).toHaveAttribute("href", "/apps/dwh/surveillance");
      await breadcrumb.getByRole("link", { name: "Surveillance", exact: true }).click();
      await expect(page).toHaveURL(new RegExp(`${base}/?$`));
      await expect(
        page.getByRole("heading", { name: "Surveillance Dashboard", exact: true }),
      ).toBeVisible();
    }
    expect(errors).toEqual([]);
  });
}

test("malformed accumulated paths recover to the app root before a tile click", async ({
  page,
}) => {
  await fixture(page);
  await page.goto(`${base}/apps/utilities/self-service/eservice/document-upload/malaria/measles`);
  await expect(page).toHaveURL(new RegExp(`${base}/?$`));
  await page
    .locator(".surveillance-tile")
    .getByRole("button", { name: "Measles", exact: true })
    .click();
  await expect(page).toHaveURL(`${base}/measles?weekId=${weekId}`);
});

test("relevant document links leave the surveillance router at the portal root", async ({
  page,
}) => {
  await fixture(page);
  await page.goto(`${base}/measles?weekId=${weekId}`);
  const link = page.getByRole("link", { name: "Measles Surveillance Guidelines", exact: true });
  await expect(link).toHaveAttribute(
    "href",
    "/portal/apps/utilities/self-service/eservice/document-upload",
  );
  await link.click();
  await expect(page).toHaveURL("/portal/apps/utilities/self-service/eservice/document-upload");
});

test("direct disease detail loads use the single shell breadcrumb", async ({ page }) => {
  await fixture(page);
  await page.goto(`${base}/measles?weekId=${weekId}`);
  const breadcrumb = page.getByRole("navigation", { name: "Breadcrumb", exact: true });
  await expect(breadcrumb).toHaveCount(1);
  await expect(breadcrumb.locator(".moh-breadcrumbs__current")).toHaveText("Measles");
  await expect(page.locator(".disease-details-page nav")).toHaveCount(0);
});
