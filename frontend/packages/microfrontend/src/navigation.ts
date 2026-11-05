// Embedded routers notify the shell after their own location has committed.
export const MICROFRONTEND_NAVIGATION_EVENT = "moh-sso:navigation";

export function notifyMicrofrontendNavigation() {
  window.dispatchEvent(new Event(MICROFRONTEND_NAVIGATION_EVENT));
}
