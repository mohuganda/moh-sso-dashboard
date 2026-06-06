import { Add, UserFollow, Security, Notification, Need } from "@carbon/react/icons";
import { Tile, Button, Tag, Stack, InlineLoading } from "@carbon/react";
import { useMemo } from "react";
import { useSelector } from "react-redux";
import "./home.css";

import { EmptyState , ErrorState , useHeaderPanel , getSeverityTagType } from "@moh-sso/ui";
import { ApplicationTile } from "@/app/home/components/app/ApplicationTile";
import { QuickAction } from "@/app/home/components/quick-action/quick-action.component";
import { SignalTile } from "@/app/home/components/signal-tile/signal-tile.component";
import { ClientFormPanel, UserClientRolesPanel } from "@moh-sso/clients";
import { UserFormPanel } from "@moh-sso/users";
import { useListClientsQuery , useAuditOverviewQuery ,
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
} from "@moh-sso/api";
import { selectUser } from "@moh-sso/auth";
import { ManageAnnouncementsPanel } from "@moh-sso/announcements";

/* -----------------------------
 * Utils
 * ----------------------------- */
const toRFC3339 = (d: Date) => d.toISOString();

export default function HomePage() {
  const { openPanel } = useHeaderPanel();
  /* -----------------------------
   * Date range (last 7 days)
   * ----------------------------- */
  const { from, to } = useMemo(() => {
    const now = new Date();
    const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    return {
      from: toRFC3339(start),
      to: toRFC3339(now),
    };
  }, []);

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

  /* -----------------------------
   * Metrics
   * ----------------------------- */
  const {
    data: metrics,
    isLoading: metricsLoading,
    isError: metricsError,
    refetch: refetchMetrics,
  } = useAuditOverviewQuery({ from, to }, { skip: !from || !to });

  /* -----------------------------
   * Security Health Score
   * ----------------------------- */
  // const securityScore = useMemo(() => {
  //   if (!metrics) return 100;

  //   let score = 100;
  //   score -= Math.min(metrics.failed_logins * 2, 40);
  //   score -= Math.min((metrics.suspicious_logins ?? 0) * 5, 40);

  //   return Math.max(score, 0);
  // }, [metrics]);

  /* -----------------------------
   * Notifications
   * ----------------------------- */
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
       * SYSTEM SIGNALS
       * ================================================== */}
      <Tile>
        <h4>System signals</h4>

        {metricsLoading && <InlineLoading description="Loading metrics…" />}

        {metricsError && (
          <ErrorState
            title="Failed to load metrics"
            description="System metrics are currently unavailable."
            primaryAction={{ label: "Retry", onClick: refetchMetrics }}
          />
        )}

        {metrics && (
          <div className="home-grid">
            {/* <SignalTile
              label="Security health score"
              value={`${securityScore}%`}
              severity={securityScore > 80 ? "success" : securityScore > 50 ? "warning" : "danger"}
            /> */}
            <SignalTile
              label="Failed logins (24h)"
              value={metrics.failed_logins}
              severity="warning"
            />
            <SignalTile label="Suspicious logins" value={0} severity="danger" />
          </div>
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
       * CLIENT USAGE METRICS
       * ================================================== */}
      <Tile>
        <h4>Client usage (last 7 days)</h4>

        {/* {!metrics?.top_clients?.length && (
          <EmptyState
            title="No usage data"
            description="No client activity recorded for this period."
          />
        )} */}

        {/* {metrics?.top_clients?.length > 0 && (
          <Stack gap={3}>
            {metrics.top_clients.map((c) => (
              <div
                key={c.client_id}
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <strong>{c.name}</strong>
                <Tag type="cool-gray">{c.logins} logins</Tag>
              </div>
            ))}
          </Stack>
        )} */}
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
                content: <ManageAnnouncementsPanel />,
                size: "lg",
              });
            }}
          />

          {/* Security alerts */}
          <QuickAction
            icon={<Security size={20} />}
            label="Security alerts"
            description="Review suspicious activity"
            href="/admin/security/alerts"
            tone="warning"
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
