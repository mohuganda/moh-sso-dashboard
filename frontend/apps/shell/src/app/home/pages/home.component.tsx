import { Add, UserFollow, Security, Notification, Need } from "@carbon/react/icons";
import { Tile, Button, Tag, Stack, InlineLoading } from "@carbon/react";
import { lazy, Suspense, useMemo, type ReactNode } from "react";
import { useSelector } from "react-redux";
import "./home.scss";

import { EmptyState, useHeaderPanel, getSeverityTagType } from "@moh-sso/ui";
import { ApplicationTile } from "@/app/home/components/app/ApplicationTile";
import { QuickAction } from "@/app/home/components/quick-action/quick-action.component";
import {
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
} from "../../api";
import { selectUser, useAuthorization } from "@moh-sso/auth";
import { buildAccessibleClients } from "@/app/access/accessClients";

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
  const { accessibleSystems } = useAuthorization();

  /* -----------------------------
   * Identity
   * ----------------------------- */
  const user = useSelector(selectUser);

  /* -----------------------------
   * Applications
   * ----------------------------- */
  const visibleClients = useMemo(
    () =>
      buildAccessibleClients({
        accessibleSystems,
      }),
    [accessibleSystems],
  );

  const { data: notifications = [], isLoading: notificationsLoading } = useGetNotificationsQuery({
    unread: true,
    limit: 5,
    offset: 0,
  });

  const [markNotificationAsRead] = useMarkNotificationAsReadMutation();

  return (
    <div className="home-page">
      {/* ==================================================
       * HEADER
       * ================================================== */}
      <div>
        <h3 className="home-page__title">Admin Overview</h3>
        <p className="home-page__subtitle">
          System status, applications, and quick actions.
        </p>
      </div>
      {/* ==================================================
       * HERO / WELCOME
       * ================================================== */}
      <Tile className="home-hero">
        {!user ? (
          <InlineLoading description="Loading user…" />
        ) : (
          <Stack gap={3}>
            <h3 className="home-hero__title">Welcome back, {user.lastName}</h3>

            <Stack orientation="horizontal" gap={3}>
              <Tag type="blue">{user?.realmRoles.join(", ")}</Tag>
              <span>{user.email}</span>
              <span className="muted">
                Last login: {user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleString() : "—"}
              </span>
            </Stack>
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
                launchMode={(client.attributes?.["ui.launchMode"] as "internal" | "new_tab" | "same_tab") || "internal"}
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
          <p className="muted home-page__subtitle">
            Common administrative tasks
          </p>
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
      <Tile>
        {/* Header */}
        <div className="notification-section__header">
          <h4 className="home-page__section-title">Notifications</h4>

          {notifications.length > 0 && (
            <Tag size="sm" type="gray">
              {notifications.filter((n) => !n.read).length} unread
            </Tag>
          )}
        </div>

        {/* Loading */}
        {notificationsLoading && <InlineLoading description="Loading notifications…" />}

        {/* Empty state */}
        {!notificationsLoading && notifications.length === 0 && (
          <EmptyState title="No notifications" description="You're all caught up." />
        )}

        {/* List */}
        {!notificationsLoading && notifications.length > 0 && (
          <Stack gap={3}>
            {notifications.map((n) => {
              return (
                <div
                  key={n.id}
                  role="listitem"
                  className={`notification-item ${!n.read ? "notification-unread" : ""}`}
                >
                  <Notification size={16} className="notification-item__icon" />

                  <div className="notification-content">
                    <div className="notification-header">
                      <strong>{n.title}</strong>
                      <Tag size="sm" type={getSeverityTagType(n.severity)}>
                        {n.severity}
                      </Tag>
                    </div>

                    <p className="notification-message">{n.message}</p>

                    <span className="notification-meta">
                      {new Date(n.created_at).toLocaleString()}
                    </span>
                  </div>

                  {!n.read && (
                    <Button size="sm" kind="ghost" onClick={() => markNotificationAsRead(n.id)}>
                      Mark as read
                    </Button>
                  )}
                </div>
              );
            })}
          </Stack>
        )}
      </Tile>
    </div>
  );
}
