import { Add, UserFollow, Security, Notification, Need } from "@carbon/react/icons";
import { Tile, Button, Tag, Stack, InlineLoading } from "@carbon/react";
import { useSelector } from "react-redux";
import "./home.scss";

import { EmptyState, ErrorState, useHeaderPanel, getSeverityTagType } from "@moh-sso/ui";
import { ApplicationTile } from "@/app/home/components/app/ApplicationTile";
import { QuickAction } from "@/app/home/components/quick-action/quick-action.component";
import { ClientFormPanel, UserClientRolesPanel } from "@moh-sso/clients";
import { UserFormPanel } from "@moh-sso/users";
import {
  useListClientsQuery,
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
} from "@moh-sso/api";
import { selectUser } from "@moh-sso/auth";
import { ManageAnnouncementsPanel } from "@moh-sso/announcements";

export default function HomePage() {
  const { openPanel } = useHeaderPanel();

  /* -----------------------------
   * Identity
   * ----------------------------- */
  const user = useSelector(selectUser);

  /* -----------------------------
   * Applications
   * ----------------------------- */
  const {
    data: clients = [],
    isLoading: appsLoading,
    isError: appsError,
    refetch: refetchApps,
  } = useListClientsQuery();

  const { data: notifications = [], isLoading: notificationsLoading } = useGetNotificationsQuery({
    unread: true,
    limit: 5,
    offset: 0,
  });

  const [markNotificationAsRead] = useMarkNotificationAsReadMutation();

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* ==================================================
       * HEADER
       * ================================================== */}
      <div>
        <h3 style={{ margin: 0 }}>Admin Overview</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
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
            <h3 style={{ margin: 0 }}>Welcome back, {user.lastName}</h3>

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
            {clients.length}
          </Tag>
        </h4>

        {appsLoading && <InlineLoading description="Loading applications…" />}

        {appsError && (
          <ErrorState
            title="Failed to load applications"
            description="Unable to fetch assigned applications."
            primaryAction={{ label: "Retry", onClick: refetchApps }}
          />
        )}

        {!appsLoading && !appsError && clients.length === 0 && (
          <EmptyState
            title="No applications assigned"
            description="You do not have access to any applications yet."
          />
        )}

        {!appsLoading && !appsError && clients.length > 0 && (
          <div className="home-grid">
            {clients.map((client) => (
              <ApplicationTile
                key={client.clientId}
                clientId={client.clientId}
                name={client.name}
                description={client.description}
                enabled={client?.enabled}
                rootUrl={client.baseUrl}
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
        <div style={{ marginBottom: "0.75rem" }}>
          <h4 style={{ margin: 0 }}>Quick actions</h4>
          <p className="muted" style={{ marginTop: 4 }}>
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
                content: <UserFormPanel mode="create" />,
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
                content: <ClientFormPanel mode="create" />,
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
                content: (
                  <Stack gap={6}>
                    {/* User basic details */}
                    <UserFormPanel mode="edit" initialUser={user} />

                    {/* Client role assignment */}
                    <UserClientRolesPanel userId={user.id} />
                  </Stack>
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
                content: <ManageAnnouncementsPanel mode={"create"} />,
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
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "1rem",
          }}
        >
          <h4 style={{ margin: 0 }}>Notifications</h4>

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
                  style={{
                    display: "flex",
                    gap: "0.75rem",
                    padding: "0.75rem",
                    borderRadius: "4px",
                    background: !n.read ? "var(--cds-layer-accent)" : "transparent",
                    alignItems: "flex-start",
                  }}
                >
                  <Notification size={16} style={{ marginTop: 2 }} />

                  <div style={{ flex: 1 }}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        gap: "0.5rem",
                      }}
                    >
                      <strong>{n.title}</strong>
                      <Tag size="sm" type={getSeverityTagType(n.severity)}>
                        {n.severity}
                      </Tag>
                    </div>

                    <p
                      style={{
                        margin: "0.25rem 0",
                        opacity: 0.85,
                      }}
                    >
                      {n.message}
                    </p>

                    <span
                      style={{
                        fontSize: "0.75rem",
                        opacity: 0.6,
                      }}
                    >
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
