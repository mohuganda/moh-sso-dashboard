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

- Search filters the already-authorized list by application name, description or
  identifier. Filtering never adds applications or changes launch permissions.
- Results retain the configured ordering, with a live result count and a clear
  no-match state. Clearing search returns focus to the search field.
- Opening focuses search when present. Escape, the close control and outside
  clicks dismiss the launcher; closing restores focus to its trigger.
- The menu fits the viewport on mobile and scrolls internally when needed.
  Tile animation respects the reduced-motion preference.

## Verification

`tests/browser/shell.spec.mjs` covers both shell layouts, mobile navigation,
viewport transitions, keyboard links, launcher filtering and clearing, account
panel access, overlay hand-off, and the narrow-screen public header. The existing
root-page and portal browser suites continue to validate application mounting
and content alignment beneath the shell.
