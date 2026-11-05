# Shell UX and interaction conventions

This refinement preserves the existing Carbon design and the portal's routes,
permissions, typography and branding. It changes shell navigation and shared
shell components; individual application screens remain separate work.

## Navigation

- On compact screens (1056px and below), the header menu button opens navigation.
  A route change, backdrop click or Escape closes it and restores page scrolling.
- Desktop collapse/expand preferences are stored separately for administration
  and the user workspace. Opening a mobile drawer does not overwrite them.
- Resizing between desktop and compact layouts reconciles drawer visibility,
  focus trapping and body scroll locking.
- Navigation links have real, base-path-aware destinations. Plain clicks use the
  existing client-side navigation; modified clicks retain native link behavior.
- Hidden navigation and closed panels use `inert` as well as their existing
  visibility semantics, excluding off-screen controls from keyboard navigation.

## Header and account

- Public visitors receive a visible Sign in action.
- Both workspaces expose the application launcher and an account panel. The
  account panel identifies the signed-in user and provides workspace switching,
  profile/settings access when permitted, and sign-out.
- On compact screens, sign-out lives in the account panel instead of occupying
  another header action. Components without an account action retain their
  existing standalone logout control.
- Brand text truncates inside its own flex item instead of overlapping the MOH
  prefix or action buttons. Header home links are keyboard-accessible links.

## Application launcher

- The launcher has expanded, labelled sections: Data & Reporting, Surveillance,
  Documents & Support, Services & Utilities, and Administration. Empty sections
  are hidden. A tile appears in exactly one section; grouping never grants access.
- `ui.category` selects a section. Existing `services`, `utilities`, `support`,
  `documents`, `data-statistics` and `admin` values map to the corresponding labels.
  Other meaningful categories become additional sections, sorted alphabetically
  after the standard sections. Uncategorized apps appear under Other applications.
- Legacy `platform`/`external` or blank categories use the known built-in route's
  purpose, so deployed records work without reseeding. Unrecognized routes fall
  back to Other applications. An explicit `other` category disables inference.
- A navigation item's optional `category` overrides its system category, for
  example `"category": "documents-support"`. The Systems editor's navigation JSON
  can configure this independently for each launcher entry.
- Search filters the already-authorized list by application name, description,
  identifier or group label. Filtering never adds applications or changes launch permissions.
- Results retain the configured ordering within each section, with a live total result count and a clear
  no-match state. Clearing search returns focus to the search field.
- Opening focuses search when present. Escape, the close control and outside
  clicks dismiss the launcher; closing restores focus to its trigger.
- The menu fits the viewport on mobile and scrolls internally when needed.
  Tile animation respects the reduced-motion preference.

## Verification

`tests/browser/shell.spec.mjs` covers both shell layouts, mobile navigation,
viewport transitions, keyboard links, launcher filtering and clearing, account
panel access, overlay hand-off, and the narrow-screen public header. Grouped launcher
checks cover section order, custom categories and item overrides, duplicate
destinations, permission filtering, category search, and internal/external launching
on desktop and mobile. The existing
root-page and portal browser suites continue to validate application mounting
and content alignment beneath the shell.

## Notifications

The overview uses a full-width Carbon table consistent with Announcements. It
provides text search, Status and Severity filters, a reset action, and pagination
with 5/10/20 rows per page. Notification titles and the row's View details action
open the detail drawer. Unread rows also offer Mark as read. The query loads the
latest 100 records; filters and table counts apply to that recent batch, and the
heading identifies the limit when reached. The unread badge is the server total.

The header drawer retains its compact feed and subscribes inside its content, so
read/delete mutations refresh the visible list and badge without closing it.
Opening the drawer refreshes cached activity. Each drawer view identifies its
latest entries (up to 20).

A single drawer scrollbar handles feed and detail content. Content fills the normal
36rem drawer and stays centered at a readable width when maximized. Notification
cards use a real detail button alongside separate read/delete actions; nested
interactive controls are avoided. Deleting requires inline confirmation, with
focus moving to the safe Keep action and returning on cancellation. Removed items
return keyboard focus to the remaining feed controls.

Details provide Back navigation, live read status and a mark-read action. Returning
preserves the selected Unread/Recent view and restores focus. Event identifiers,
target roles and JSON metadata are available in collapsed Technical details.
Delivery history separates channel names, attempt counts, statuses and timestamps;
retry controls retain their existing permission checks. Load/action failures show
explicit errors and retry options, and failed mutations do not imply success.

`tests/browser/notifications.spec.mjs` covers normal/maximized/mobile layout,
live cache updates, keyboard navigation, delete confirmation, empty/error recovery,
overview table filters, pagination, row actions and delivery retry permissions using disposable API data.
