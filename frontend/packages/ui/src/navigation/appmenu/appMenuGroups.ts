import type { Client } from "@moh-sso/types";

const GROUPS = [
  { key: "data-reporting", label: "Data & Reporting" },
  { key: "surveillance", label: "Surveillance" },
  { key: "documents-support", label: "Documents & Support" },
  { key: "services-utilities", label: "Services & Utilities" },
  { key: "administration", label: "Administration" },
] as const;

type GroupKey = (typeof GROUPS)[number]["key"];
type Category = { key: string; label: string };
export type AppMenuGroup = Category & { clients: Client[] };

const CATEGORY_ALIASES: Record<string, GroupKey> = {
  "data-reporting": "data-reporting",
  "data-statistics": "data-reporting",
  data: "data-reporting",
  reporting: "data-reporting",
  surveillance: "surveillance",
  "documents-support": "documents-support",
  documents: "documents-support",
  support: "documents-support",
  "services-utilities": "services-utilities",
  services: "services-utilities",
  utilities: "services-utilities",
  administration: "administration",
  admin: "administration",
};

// Older system records use platform/external to describe the system type rather
// than its purpose. Resolve their built-in routes without changing stored data.
const BUILT_IN_ROUTES: [RegExp, GroupKey][] = [
  [/^\/apps\/dwh\/(documents|issue-tracker)(\/|$)/, "documents-support"],
  [/^\/apps\/(dwh\/surveillance|outbreak-management)(\/|$)/, "surveillance"],
  [/^\/apps\/dwh(\/|$)/, "data-reporting"],
  [/^\/apps\/(eservices|utilities|settings)(\/|$)/, "services-utilities"],
  [/^\/admin(\/|$)/, "administration"],
];

const OTHER: Category = { key: "other", label: "Other applications" };

function normalizeCategory(category: string): string {
  return category.toLocaleLowerCase().replace(/[\s_&-]+/g, "-");
}

export function getAppMenuCategory(client: Client): Category {
  const category = client.attributes?.["ui.category"]?.trim() ?? "";
  const key = normalizeCategory(category);
  const canonicalKey = CATEGORY_ALIASES[key];
  const known = GROUPS.find((group) => group.key === canonicalKey);
  if (known) return known;

  if (key && !["platform", "external", "other", "other-applications"].includes(key)) {
    const label = category.replace(/[_-]+/g, " ").replace(/\s+/g, " ");
    return { key: `custom:${key}`, label: label.charAt(0).toLocaleUpperCase() + label.slice(1) };
  }

  // An explicit Other category opts out of inference.
  if (key === "other" || key === "other-applications") return OTHER;

  const path = (client.baseUrl?.trim() ?? "").split(/[?#]/)[0].replace(/^\/portal(?=\/|$)/, "");
  const routeGroup = BUILT_IN_ROUTES.find(([pattern]) => pattern.test(path));
  return GROUPS.find((group) => group.key === routeGroup?.[1]) ?? OTHER;
}

export function groupAppMenuClients(clients: Client[]): AppMenuGroup[] {
  const groups = new Map<string, AppMenuGroup>();
  for (const client of clients) {
    const category = getAppMenuCategory(client);
    const group = groups.get(category.key) ?? { ...category, clients: [] };
    group.clients.push(client);
    groups.set(category.key, group);
  }

  const order = (key: string) => {
    if (key === OTHER.key) return GROUPS.length + 1;
    const index = GROUPS.findIndex((group) => group.key === key);
    return index < 0 ? GROUPS.length : index;
  };
  return [...groups.values()].sort(
    (a, b) => order(a.key) - order(b.key) || a.label.localeCompare(b.label),
  );
}
