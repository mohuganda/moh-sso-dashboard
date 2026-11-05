import { test, expect } from "@playwright/test";
import { URL } from "node:url";
import { readFileSync } from "node:fs";

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

function polygon(district, subcounty, x) {
  return {
    type: "Feature",
    properties: { District: district, Subcounty: subcounty },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [x, 1],
          [x + 0.1, 1],
          [x + 0.1, 1.1],
          [x, 1.1],
          [x, 1],
        ],
      ],
    },
  };
}
const hoimaCityBoundary = JSON.parse(
  readFileSync(
    new URL("../../../backend/assets/geojson/uganda_districts.geojson", import.meta.url),
    "utf8",
  ),
).features.find((feature) => feature.properties.District === "HOIMA CITY");
const collection = (features) => ({ type: "FeatureCollection", features });

async function mapFixture(page, noBoundaries = false) {
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.route("**/ssobackend/**", async (route) => {
    const path = new URL(route.request().url()).pathname;
    let data = [];
    if (path.endsWith("/auth/me")) data = { user };
    else if (path.endsWith("/notifications/count/unread")) data = { count: 0 };
    else if (path.endsWith("/surveillance/weeks"))
      data = [{ id: weekId, epi_year: 2026, epi_week: 39 }];
    else if (path.endsWith("/surveillance/diseases")) data = diseases;
    else if (path.endsWith("/surveillance/regions")) data = [{ id: "bunyoro", name: "Bunyoro" }];
    else if (path.endsWith("/regions/bunyoro/districts"))
      data = [
        { id: "city", name: "Hoima City" },
        { id: "alpha", name: "Alpha" },
      ];
    else if (path.endsWith("/districts/city/subcounties"))
      data = [{ id: "council", name: "Hoima City Council" }];
    else if (path.endsWith("/districts/alpha/subcounties")) data = [{ id: "twin", name: "Twin" }];
    else if (path.endsWith("/geo_districts"))
      data = collection(noBoundaries ? [] : [hoimaCityBoundary, polygon("Alpha", "", 32)]);
    else if (path.endsWith("/geo_subcounties"))
      data = collection(
        noBoundaries ? [] : [polygon("Alpha", "Twin", 32), polygon("Other district", "Twin", 33)],
      );
    else if (path.includes("/geojson/")) data = collection([]);
    else if (path.includes("/weekly-statuses/"))
      data = ["MAROON", "RED"].map((status, i) => ({
        id: String(i),
        epi_week_id: weekId,
        region_id: "bunyoro",
        region_name: "Bunyoro",
        district_id: "city",
        district_name: "Hoima City",
        disease_id: diseases[i].id,
        status,
      }));
    await route.fulfill({ json: { success: true, data } });
  });
  return errors;
}

for (const width of [390, 1440]) {
  test(`map retains district fallback and correct boundaries at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 1000 });
    const errors = await mapFixture(page);
    await page.goto(base);
    await page.getByLabel("Region", { exact: true }).selectOption("bunyoro");
    await expect(
      page
        .getByLabel("District", { exact: true })
        .getByRole("option", { name: "Hoima City", exact: true }),
    ).toHaveCount(1);
    await page.getByLabel("District", { exact: true }).selectOption("city");
    await expect(
      page.getByText("Sub-county boundaries are unavailable", { exact: false }),
    ).toBeVisible();
    const paths = page.locator(".surveillance-map__leaflet path.leaflet-interactive");
    await expect(paths).toHaveCount(1);
    await expect(paths.first()).toHaveAttribute("fill", "#7d1733");
    await paths.first().hover();
    await expect(page.locator(".leaflet-tooltip")).toContainText("HOIMA CITY");
    await expect(page.locator(".leaflet-tooltip")).toContainText("MAROON");
    await expect(
      page.getByText("No facility reports for the selected week and location.", { exact: true }),
    ).toBeVisible();
    await page.screenshot({ path: `/tmp/moh-map-fallback-${width}.png`, fullPage: true });
    await page.getByLabel("District", { exact: true }).selectOption("alpha");
    await expect(
      page.getByText("Sub-county boundaries are unavailable", { exact: false }),
    ).toHaveCount(0);
    await expect(paths).toHaveCount(1);
    await expect(page.locator(".leaflet-tooltip").filter({ hasText: "HOIMA CITY" })).toHaveCount(0);
    await paths.first().hover();
    await expect(page.locator(".leaflet-tooltip")).toContainText("Twin");
    await expect(paths.first()).toHaveAttribute("fill", "#c7dced");
    expect(errors).toEqual([]);
  });
}

test("missing boundaries show an explanation instead of a blank map", async ({ page }) => {
  await mapFixture(page, true);
  await page.goto(base);
  await expect(
    page.getByText("No map boundaries are available for the selected location."),
  ).toBeVisible();
  await expect(page.locator(".surveillance-map__canvas")).toBeHidden();
});
