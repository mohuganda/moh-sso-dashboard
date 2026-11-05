import { Add, UserFollow, Security, Need } from "@carbon/react/icons";
import { Tile, Tag, Stack, InlineLoading } from "@carbon/react";
import { lazy, Suspense, useMemo, type ReactNode } from "react";
import { useDispatch, useSelector } from "react-redux";
import "./home.scss";

import { EmptyState, useHeaderPanel } from "@moh-sso/ui";
import { ApplicationTile } from "@/app/home/components/app/ApplicationTile";
import { QuickAction } from "@/app/home/components/quick-action/quick-action.component";
import { NotificationsTable } from "@/app/notifications/components/NotificationsTable";
import { selectUser, useAuthorization } from "@moh-sso/auth";
import { setActiveClient } from "@moh-sso/state";
import { buildAccessibleLauncherEntries } from "@/app/access/accessClients";

const ClientFormPanel = lazy(() =>
  import("@moh-sso/clients").then((module) => ({ default: module.ClientFormPanel })),
);
const UserClientRolesPanel = lazy(() =>
  import("@moh-sso/clients").then((module) => ({ default: module.UserClientRolesPanel })),
);
const UserFormPanel = lazy(() =>
  import("@moh-sso/users").then((module) => ({ default: module.UserFormPanel })),
);
const ManageAnnouncementsPanel = lazy(() =>
  import("@moh-sso/announcements").then((module) => ({
    default: module.ManageAnnouncementsPanel,
  })),
);

function normalizePortalPath(href?: string): string | undefined {
  if (!href || /^https?:\/\//i.test(href)) {
    return href;
  }

  if (href === "/portal") {
    return "/apps";
  }

  if (href.startsWith("/portal/")) {
    return href.replace(/^\/portal/, "");
  }

  return href;
}

function LazyPanelFallback() {
  return <InlineLoading description="Loading panel..." />;
}

function lazyPanel(content: ReactNode) {
  return <Suspense fallback={<LazyPanelFallback />}>{content}</Suspense>;
}

export default function HomePage() {
  const { openPanel } = useHeaderPanel();
  const dispatch = useDispatch();
  const { accessibleSystems, can } = useAuthorization();

  /* -----------------------------
   * Identity
   * ----------------------------- */
  const user = useSelector(selectUser);

  /* -----------------------------
   * Applications
   * ----------------------------- */
  const visibleClients = useMemo(
    () =>
      buildAccessibleLauncherEntries({
        accessibleSystems,
        can: (permission) => can(permission as never),
      }),
    [accessibleSystems, can],
  );

  return (
    <div className="home-page">
      {/* ==================================================
       * HEADER
       * ================================================== */}
      <div>
        <h3 className="home-page__title">Admin Overview</h3>
        <p className="home-page__subtitle">System status, applications, and quick actions.</p>
      </div>
      {/* ==================================================
       * HERO / WELCOME
       * ================================================== */}
      <Tile className="home-hero">
        {!user ? (
          <InlineLoading description="Loading user…" />
        ) : (
          <Stack gap={3}>
            <h3 className="home-hero__title">
              Welcome back, {user.firstName || user.lastName || user.username}
            </h3>

            <div className="home-hero__identity">
              <Tag type="blue">{user?.realmRoles.join(", ")}</Tag>
              <span>{user.email}</span>
              <span className="muted">
                Last login: {user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleString() : "—"}
              </span>
            </div>
          </Stack>
        )}
      </Tile>
      {/* ==================================================
       * APPLICATIONS
       * ================================================== */}
      <Tile>
        <h4>
          Your applications{" "}
          <Tag size="sm" type="cool-gray">
            {visibleClients.length}
          </Tag>
        </h4>

        {visibleClients.length === 0 && (
          <EmptyState
            title="No applications assigned"
            description="You do not have access to any applications yet."
          />
        )}

        {visibleClients.length > 0 && (
          <div className="home-grid">
            {visibleClients.map((client) => (
              <ApplicationTile
                key={client.clientId}
                clientId={client.clientId}
                name={client.name}
                description={client.description}
                enabled={client?.enabled}
                rootUrl={normalizePortalPath(client.baseUrl)}
                launchMode={
                  (client.attributes?.["ui.launchMode"] as "internal" | "new_tab" | "same_tab") ||
                  "internal"
                }
                onLaunch={() =>
                  dispatch(
                    setActiveClient(client.attributes?.["ui.parentClientId"] || client.clientId),
                  )
                }
              />
            ))}
          </div>
        )}
      </Tile>

      {/* ==================================================
       * QUICK ACTIONS
       * ================================================== */}
      <Tile>
        {/* Header */}
        <div className="home-page__section-heading">
          <h4 className="home-page__section-title">Quick actions</h4>
          <p className="muted home-page__subtitle">Common administrative tasks</p>
        </div>

        <div className="home-grid">
          {/* Create user */}
          <QuickAction
            icon={<Add size={20} />}
            label="Create user"
            description="Add a new user to the system"
            onClick={() => {
              openPanel({
                title: "Create user",
                content: lazyPanel(<UserFormPanel mode="create" />),
                size: "md",
              });
            }}
            tone="warning"
          />

          {/* Create client */}
          <QuickAction
            icon={<UserFollow size={20} />}
            label="Create client"
            description="Register a new application client"
            onClick={() => {
              openPanel({
                title: "Create client",
                content: lazyPanel(<ClientFormPanel mode="create" />),
                size: "md",
              });
            }}
          />

          {/* Manage user roles */}
          <QuickAction
            icon={<Security size={20} />}
            label="Manage my roles"
            description="View and manage your application roles"
            tone="warning"
            onClick={() => {
              if (!user) return;

              openPanel({
                title: `Roles: ${user.username}`,
                size: "lg",
                content: lazyPanel(
                  <Stack gap={6}>
                    {/* User basic details */}
                    <UserFormPanel mode="edit" initialUser={user} />

                    {/* Client role assignment */}
                    <UserClientRolesPanel userId={user.id} />
                  </Stack>,
                ),
              });
            }}
          />

          {/* Add and announcement */}
          <QuickAction
            icon={<Need size={20} />}
            label="Manage  announcement"
            description="View and manage announcements"
            onClick={() => {
              openPanel({
                title: "Create Announcement",
                content: lazyPanel(<ManageAnnouncementsPanel mode="create" />),
                size: "lg",
              });
            }}
          />
        </div>
      </Tile>

      {/* ==================================================
       * NOTIFICATIONS (ACTIONABLE)
       * ================================================== */}
      <Tile className="home-notifications">
        <NotificationsTable />
      </Tile>
    </div>
  );
}
