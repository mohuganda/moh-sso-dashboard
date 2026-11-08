import { Add, Application, Security, Need, Time, UserAvatar } from "@carbon/react/icons";
import { Tile, Tag, Search, InlineLoading } from "@carbon/react";
import { lazy, Suspense, useMemo, useState, type ReactNode } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import "./home.scss";

import { EmptyState, PageShell, useHeaderPanel } from "@moh-sso/ui";
import { ApplicationTile } from "@/app/home/components/app/ApplicationTile";
import { QuickAction } from "@/app/home/components/quick-action/quick-action.component";
import { NotificationsTable } from "@/app/notifications/components/NotificationsTable";
import { PERMISSIONS, selectUser, useAuthorization } from "@moh-sso/auth";
import { setActiveClient } from "@moh-sso/state";
import { buildAccessibleLauncherEntries } from "@/app/access/accessClients";

const ClientFormPanel = lazy(() =>
  import("@moh-sso/clients").then((module) => ({ default: module.ClientFormPanel })),
);
const UserFormPanel = lazy(() =>
  import("@moh-sso/users").then((module) => ({ default: module.UserFormPanel })),
);
const ManageAnnouncementsPanel = lazy(() =>
  import("@moh-sso/announcements").then((module) => ({ default: module.ManageAnnouncementsPanel })),
);

function normalizePortalPath(href?: string): string | undefined {
  if (!href || /^https?:\/\//i.test(href)) return href;
  if (href === "/portal") return "/apps";
  return href.startsWith("/portal/") ? href.replace(/^\/portal/, "") : href;
}

function lazyPanel(content: ReactNode) {
  return <Suspense fallback={<InlineLoading description="Loading panel…" />}>{content}</Suspense>;
}

function formatLastLogin(value?: string | null) {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(
    date,
  );
}

export default function HomePage() {
  const { openPanel } = useHeaderPanel();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { accessibleSystems, can } = useAuthorization();
  const user = useSelector(selectUser);
  const [search, setSearch] = useState("");
  const visibleClients = useMemo(
    () =>
      buildAccessibleLauncherEntries({
        accessibleSystems,
        can: (permission) => can(permission as never),
      }),
    [accessibleSystems, can],
  );
  const filteredClients = useMemo(() => {
    const query = search.trim().toLocaleLowerCase();
    return visibleClients.filter((client) =>
      `${client.name} ${client.description ?? ""}`.toLocaleLowerCase().includes(query),
    );
  }, [visibleClients, search]);
  const lastLogin = formatLastLogin(user?.lastLoginAt);
  const actions = [
    {
      id: "user",
      allowed: can(PERMISSIONS.usersWrite),
      icon: <Add size={20} />,
      label: "Create user",
      description: "Add a person and set up their account.",
      onClick: () =>
        openPanel({
          title: "Create user",
          content: lazyPanel(<UserFormPanel mode="create" />),
          size: "md",
        }),
    },
    {
      id: "client",
      allowed: can(PERMISSIONS.clientsWrite),
      icon: <Application size={20} />,
      label: "Create client",
      description: "Connect an application to Ministry sign-in.",
      onClick: () =>
        openPanel({
          title: "Create client",
          content: lazyPanel(<ClientFormPanel mode="create" />),
          size: "md",
        }),
    },
    {
      id: "access",
      allowed: can(PERMISSIONS.rbacRead),
      icon: <Security size={20} />,
      label: "Manage access",
      description: "Review application roles and permissions.",
      onClick: () => navigate("/admin/rbac"),
    },
    {
      id: "announcement",
      allowed: can(PERMISSIONS.announcementsWrite),
      icon: <Need size={20} />,
      label: "New announcement",
      description: "Draft an update for portal users.",
      onClick: () =>
        openPanel({
          title: "Create announcement",
          content: lazyPanel(<ManageAnnouncementsPanel mode="create" />),
          size: "lg",
        }),
    },
  ].filter((action) => action.allowed);

  return (
    <div className="home-page">
      <PageShell
        title="Admin Overview"
        subtitle="Your applications, administrative tasks, and latest activity."
      >
        <Tile className="home-hero">
          {!user ? (
            <InlineLoading description="Loading your workspace…" />
          ) : (
            <>
              <div className="home-hero__welcome">
                <span className="home-hero__avatar" aria-hidden="true">
                  <UserAvatar size={28} />
                </span>
                <div>
                  <p className="home-hero__eyebrow">Your workspace</p>
                  <h4 className="home-hero__title">
                    Welcome back, {user.firstName || user.username}
                  </h4>
                  {user.email && <p className="home-hero__email">{user.email}</p>}
                </div>
              </div>
              <div className="home-hero__context">
                <Tag type="blue" size="sm">
                  {user.isAdmin ? "Administrator" : "Signed in"}
                </Tag>
                {lastLogin && (
                  <p className="home-hero__last-login">
                    <Time size={16} aria-hidden="true" />
                    <span>Last sign-in: {lastLogin}</span>
                  </p>
                )}
              </div>
            </>
          )}
        </Tile>

        <section className="home-applications" aria-labelledby="home-applications-title">
          <header className="home-section-heading">
            <div>
              <div className="home-section-heading__title">
                <h4 id="home-applications-title">Your applications</h4>
                <Tag type="cool-gray" size="sm">
                  {visibleClients.length}
                </Tag>
              </div>
              <p>Open the tools available to your account.</p>
            </div>
            {visibleClients.length > 0 && (
              <Search
                id="home-application-search"
                size="md"
                labelText="Search your applications"
                placeholder="Find an application"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                onClear={() => setSearch("")}
              />
            )}
          </header>
          {visibleClients.length === 0 ? (
            <EmptyState
              title="No applications assigned"
              description="Applications will appear here when access is assigned to your account."
            />
          ) : (
            <>
              <p className="home-applications__results" role="status">
                {search.trim()
                  ? `${filteredClients.length} of ${visibleClients.length} applications`
                  : `${visibleClients.length} applications available`}
              </p>
              {filteredClients.length === 0 ? (
                <EmptyState
                  title="No matching applications"
                  description="Try a different application name or clear your search."
                  primaryAction={{ label: "Clear search", onClick: () => setSearch("") }}
                />
              ) : (
                <div className="home-applications__grid">
                  {filteredClients.map((client) => (
                    <ApplicationTile
                      key={client.clientId}
                      clientId={client.clientId}
                      name={client.name}
                      description={client.description}
                      enabled={client.enabled}
                      rootUrl={normalizePortalPath(client.baseUrl)}
                      launchMode={
                        (client.attributes?.["ui.launchMode"] as
                          | "internal"
                          | "new_tab"
                          | "same_tab") || "internal"
                      }
                      onLaunch={() =>
                        dispatch(
                          setActiveClient(
                            client.attributes?.["ui.parentClientId"] || client.clientId,
                          ),
                        )
                      }
                    />
                  ))}
                </div>
              )}
            </>
          )}
        </section>

        {actions.length > 0 && (
          <section aria-labelledby="home-actions-title">
            <header className="home-section-heading">
              <div>
                <h4 id="home-actions-title">Quick actions</h4>
                <p>Common tasks for your administrative access.</p>
              </div>
            </header>
            <div className="home-actions">
              {actions.map(({ id, ...action }) => (
                <QuickAction key={id} {...action} />
              ))}
            </div>
          </section>
        )}

        {can(PERMISSIONS.notificationsRead) && (
          <Tile className="home-notifications">
            <NotificationsTable />
          </Tile>
        )}
      </PageShell>
    </div>
  );
}
