export type PortalPreferences = {
  compactTables: boolean;
  itemsPerPage: number;
  sideNavCollapsed: boolean;
  reducedMotion: boolean;
  largerText: boolean;
  highContrast: boolean;
  timezone: string;
  dateFormat: "system" | "short" | "medium";
  favoriteSystemIds: string[];
  defaultLandingPath: string;
};

const STORAGE_KEY = "moh.portal.userPreferences";

export const defaultPortalPreferences: PortalPreferences = {
  compactTables: false,
  itemsPerPage: 20,
  sideNavCollapsed: false,
  reducedMotion: false,
  largerText: false,
  highContrast: false,
  timezone: "Africa/Kampala",
  dateFormat: "system",
  favoriteSystemIds: [],
  defaultLandingPath: "/portal/apps/news",
};

export function readPortalPreferences(): PortalPreferences {
  if (typeof window === "undefined") {
    return defaultPortalPreferences;
  }

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);

    if (!stored) {
      return defaultPortalPreferences;
    }

    return {
      ...defaultPortalPreferences,
      ...(JSON.parse(stored) as Partial<PortalPreferences>),
    };
  } catch {
    return defaultPortalPreferences;
  }
}

export function savePortalPreferences(preferences: PortalPreferences) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(preferences));
}

export function isSafeInternalPath(path: string): boolean {
  return path.startsWith("/apps/") && !path.startsWith("//");
}
