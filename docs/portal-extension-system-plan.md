# Portal extension system implementation plan

Branch: `feat/portal-extension-system`, created directly in the primary checkout
from upstream `main` at `615052d8`. This branch contains the implementation plan;
runtime implementation starts with task EXT-01.

## Objective

Allow frontend modules to contribute independently loaded UI features to named
slots, with configuration controlling placement, order, settings and visibility.
Use single-spa parcels for mounting and follow the OpenMRS extension/slot pattern.
The portal owns its contracts and permission model; installing OpenMRS framework
packages or claiming OpenMRS module compatibility is outside the initial scope.

Existing route-based applications continue to own their screens. Shared React
controls continue to use regular component imports. The first delivery provides
configuration-driven composition; an administrator editing UI and persisted
configuration are a later delivery.

## Current implementation and constraints

- `frontend/packages/microfrontend/src/lifecycle.ts` defines whole-application
  lifecycles; `props.ts` supplies auth, API base URL, event bus and mount DOM node.
- `frontend/packages/microfrontend/src/react-lifecycle.tsx` wraps React roots and
  shared providers. It currently caches a lifecycle/root for an application;
  parcel instances need independent roots when rendered in multiple slots.
- The shell's `orchestrator.ts` handles remote single-spa applications;
  `SingleSpaApp.tsx` handles hybrid/local mounting and runtime props.
- `frontend/apps/shell/src/app/microfrontends/registry.ts` defines app routes and
  existing permissions/system requirements. Extension registrations must be
  separate from URL activation and use those same authorization semantics.
- `frontend/packages/auth/src/auth/permissions.ts` defines permission names such
  as `documents:read` and `report_scheduler:read`. Avoid introducing a parallel
  permission vocabulary.
- Production modules are loaded through import maps. A runtime registry must be
  a shared singleton across module bundles, with local/remote loading parity.
- The existing public news feed contains announcements. Its anonymous access and
  layout must remain correct when introducing authenticated dashboard widgets.

## Tasks and acceptance criteria

### EXT-01 — Define the contracts and pilot host

Depends on: none.

- [ ] Define types for slots, extension definitions, stable instance IDs, parcel
      exports, props, configuration, permission/system requirements and failures.
- [ ] Specify required mount/unmount lifecycles and optional bootstrap/update,
      including behavior when an extension does not support updates.
- [ ] Define data-only configuration: add, remove, order and configure. Establish
      override precedence, deterministic tie-breaking and invalid-config behavior.
- [ ] Define a registry API version and compatibility checks for extension modules.
- [ ] Select an existing authenticated dashboard/workspace host and its empty
      state; document its route and permission policy before changing the screen.
- [ ] Define the supported v1 behavior in hybrid/local and orchestrated/remote modes.

Acceptance: a typed example declares a slot, registers an extension and resolves
its final configuration; permission and configuration semantics are documented.

### EXT-02 — Implement the shared registry and configuration resolver

Depends on: EXT-01.

- [ ] Add extension definitions, slot ownership and registration APIs under
      `@moh-sso/microfrontend`, with a singleton shared by shell and remote modules.
- [ ] Register public module metadata without importing each extension's UI.
- [ ] Resolve defaults and configuration overrides deterministically; validate
      unknown extension IDs, duplicate IDs, incompatible contracts and settings.
- [ ] Support multiple instances of one extension with separate configuration.
- [ ] Filter assignments using existing permissions, system access and feature
      flags before loading unauthorized feature code. Refresh after auth changes.
- [ ] Cache module loading while preserving separate mounted instance state.

Acceptance: focused tests cover ordering, overrides, duplicate registration,
instance isolation, access changes and local/remote registry identity.

### EXT-03 — Add parcel lifecycle and runtime services

Depends on: EXT-01; integrate with EXT-02.

- [ ] Add a React parcel factory that creates a separate root/provider lifecycle
      per instance, without changing existing whole-application lifecycle behavior.
- [ ] Supply documented auth, API, shared state/services, event bus, navigation,
      slot state and instance configuration explicitly to each parcel root.
- [ ] Forward application-bound `mountParcel` through orchestrated lifecycle props.
- [ ] Define the hybrid/shell-owned mounting adapter; any root parcels it creates
      must be explicitly unmounted when their host is removed.
- [ ] Handle asynchronous load/mount/unmount races, prop updates and React Strict
      Mode. Clean up subscriptions, listeners and DOM roots on every removal path.

Acceptance: mounting one extension twice creates independent instances; route
changes, logout, configuration removal and a late module load leave no orphan
roots. Provider access and updates work in both supported runtime modes.

### EXT-04 — Build the ExtensionSlot renderer

Depends on: EXT-02, EXT-03.

- [ ] Implement `ExtensionSlot` with typed state and per-instance containers.
- [ ] Add loading, empty, error and retry behavior per extension, including module
      load and lifecycle failures as well as React render failures.
- [ ] Keep grid sizing, spacing and responsive layout in the host slot; feature
      components adapt to the available space.
- [ ] Make order changes preserve instance identity where possible; ensure useful
      accessible labels and focus handling when items are removed.
- [ ] Add development diagnostics for slot, extension ID and lifecycle stage,
      excluding tokens and sensitive props.

Acceptance: one broken extension leaves its siblings usable; reordering, repeated
navigation and multiple slot instances preserve layout and lifecycle correctness.

### EXT-05 — Deliver the first dashboard extension

Depends on: EXT-04.

- [ ] Export a read-only scheduled-reports summary parcel from Report Scheduler,
      using an existing authorized API and `report_scheduler:read` access policy.
- [ ] Add the authenticated dashboard slot chosen in EXT-01 and register that
      extension as a default contribution.
- [ ] Confirm its entry point renders a widget without application-level routing
      or a second shell layout.
- [ ] Demonstrate configuration-driven hide, reorder and per-instance settings.
- [ ] Keep the existing scheduled-reports screen and public news feed working.

Acceptance: the widget loads locally and from the production import map; allowed,
denied, loading, empty and API-error states behave correctly. Disabling the
extension restores the host's intended empty/fallback state.

### EXT-06 — Verify production packaging and browser behavior

Depends on: EXT-05.

- [ ] Include parcel exports in published modules and generated declarations;
      verify import maps and external/shared dependencies retain one registry.
- [ ] Extend staged-module checks where needed to verify parcel export contracts.
- [ ] Add browser coverage for multiple instances, auth/config changes, failed
      loading, route cleanup and layout at narrow and wide viewport sizes.
- [ ] Run the existing frontend lint, typecheck, architecture checks, runtime build
      and browser smoke tests; verify existing route-based apps still mount.

Acceptance: CI passes with the production remote runtime as well as the local
pilot, and no duplicate provider/runtime or registry instance is introduced.

### EXT-07 — Document authoring and enable the pilot

Depends on: EXT-06.

- [ ] Add an authoring guide with a working extension, slot and configuration example.
- [ ] Document supported props, permissions, version compatibility, CSS ownership,
      lifecycle cleanup, error handling and the trusted-module loading policy.
- [ ] Start behind a disabled-by-default feature flag and verify the enabled pilot
      before changing the deployment's default configuration.
- [ ] Document how to disable the pilot and restore the host's previous rendering.

Acceptance: a developer can add a second extension using the guide, and an operator
can enable or disable the pilot without editing application source.

## Follow-up delivery

After the pilot is verified, separately scope additional document/announcement
widgets, administrator configuration editing, authorized server persistence,
audit history and organization-specific configuration. Adding those requires
explicit API/storage and permission design; v1 configuration comes from reviewed
runtime configuration. Connectivity rules and richer display expressions can be
added once there is a demonstrated need; configuration must never evaluate
arbitrary JavaScript.

## Suggested review sequence

1. Contracts and registry: EXT-01–02.
2. Parcel runtime and slot renderer: EXT-03–04.
3. Working dashboard pilot, verification and documentation: EXT-05–07.

Each review should leave existing application routing functional. Feature
implementation begins after the contract and pilot-host decisions in EXT-01.

## References

- [single-spa parcels](https://single-spa.js.org/docs/parcels-overview/)
- [OpenMRS extension system](https://o3-docs.openmrs.org/en-US/docs/extension-system/)
