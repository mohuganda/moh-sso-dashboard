import { test, expect } from "@playwright/test";
import { URL } from "node:url";
import { readFileSync } from "node:fs";
import { districtKey } from "../../apps/surveillance/src/pages/surveillance-geography.ts";

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
const citySubcounties = JSON.parse(
  readFileSync(
    new URL("../../../backend/assets/geojson/uganda_subcounties.geojson", import.meta.url),
    "utf8",
  ),
).features.filter((feature) => feature.properties.District === "HOIMA CITY");
const collection = (features) => ({ type: "FeatureCollection", features });

async function mapFixture(
  page,
  noBoundaries = false,
  includeCity = false,
  statuses = ["MAROON", "RED"],
) {
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
        noBoundaries
          ? []
          : [
              polygon("Alpha", "Twin", 32),
              polygon("Other district", "Twin", 33),
              ...(includeCity ? citySubcounties : []),
            ],
      );
    else if (path.includes("/geojson/")) data = collection([]);
    else if (path.includes("/weekly-statuses/"))
      data = statuses.map((status, i) => ({
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
    for (const [tone, colour] of [
      ["immediate-action", "rgb(125, 23, 51)"],
      ["take-action", "rgb(230, 139, 125)"],
      ["alert", "rgb(228, 214, 78)"],
      ["watch", "rgb(167, 226, 182)"],
    ]) {
      await expect(page.locator(`.surveillance-tile--${tone}`)).toHaveCSS(
        "background-color",
        colour,
      );
    }
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
    await expect(paths.first()).toHaveCSS("fill", "rgb(125, 23, 51)");
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
    await expect(paths.first()).toHaveCSS("fill", "rgb(199, 220, 237)");
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

test("UBOS city divisions render without assigning district flags to sub-counties", async ({
  page,
}) => {
  const errors = await mapFixture(page, false, true);
  await page.goto(base);
  await page.getByLabel("Region", { exact: true }).selectOption("bunyoro");
  await expect(
    page
      .getByLabel("District", { exact: true })
      .getByRole("option", { name: "Hoima City", exact: true }),
  ).toHaveCount(1);
  await page.getByLabel("District", { exact: true }).selectOption("city");
  await expect(
    page.getByText("Some geographic sub-counties are not yet linked", { exact: false }),
  ).toBeVisible();
  const paths = page.locator(".surveillance-map__leaflet path.leaflet-interactive");
  await expect(paths).toHaveCount(2);
  for (const path of await paths.all()) await expect(path).toHaveCSS("fill", "rgb(199, 220, 237)");
  await expect(page.getByRole("link", { name: "UBOS NPHC 2024 data" })).toBeVisible();
  await page.screenshot({ path: "/tmp/moh-ubos-city-divisions.png", fullPage: true });
  expect(citySubcounties.map((feature) => feature.properties.Subcounty).sort()).toEqual([
    "HOIMA EAST DIVISION",
    "HOIMA WEST DIVISION",
  ]);
  expect(errors).toEqual([]);
});

for (const [status, tone, colour] of [
  ["MAROON", "immediate-action", "rgb(125, 23, 51)"],
  ["RED", "take-action", "rgb(230, 139, 125)"],
  ["YELLOW", "alert", "rgb(228, 214, 78)"],
  ["GREEN", "watch", "rgb(167, 226, 182)"],
]) {
  test(`${status} tile, map fill and legend share an exact colour`, async ({ page }) => {
    await mapFixture(page, false, false, [status]);
    await page.goto(base);
    await expect(page.locator(`.surveillance-tile--${tone}`)).toHaveCSS("background-color", colour);
    await expect(
      page.locator(`.surveillance-map__legend-swatch--${status.toLowerCase()}`),
    ).toHaveCSS("background-color", colour);
    await expect(
      page.locator(".surveillance-map__leaflet path.leaflet-interactive").first(),
    ).toHaveCSS("fill", colour);
  });
}

test("bundled sub-county boundaries cover every existing district name", () => {
  const readLayer = (name) =>
    JSON.parse(
      readFileSync(
        new URL(`../../../backend/assets/geojson/${name}.geojson`, import.meta.url),
        "utf8",
      ),
    ).features;
  const names = (features) =>
    [...new Set(features.map((feature) => districtKey(feature.properties.District)))].sort();
  expect(names(readLayer("uganda_subcounties"))).toEqual(names(readLayer("uganda_districts")));
});
