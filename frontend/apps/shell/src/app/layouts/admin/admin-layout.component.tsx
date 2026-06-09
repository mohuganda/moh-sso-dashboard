import { useMemo } from "react";
import {
  Header,
  HeaderGlobalAction,
  HeaderGlobalBar,
  HeaderName,
  SideNav,
  SideNavItems,
  SideNavLink,
  Content,
} from "@carbon/react";
import {
  Notification,
  Logout,
  UserAvatarFilled,
  Dashboard,
  UserMultiple,
  Api,
  Bullhorn,
  Email,
  Activity,
} from "@carbon/react/icons";
import { useSelector } from "react-redux";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import {
  HeaderPanelProvider,
  NotificationsPanel,
  ToastProvider,
  useHeaderPanel,
} from "@moh-sso/ui";
import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";

import MohLogo from "@/assets/logo.png";
import { API } from "@moh-sso/config";
import { useGetNotificationsQuery, useGetUnreadNotificationsCountQuery } from "@moh-sso/api";
import { selectUser } from "@moh-sso/auth";

import "./admin-layout.scss";

type AdminNavItem = {
  id: string;
  label: string;
  path: string;
  icon: React.ElementType;
  exact?: boolean;
};

const ADMIN_NAV_ITEMS: AdminNavItem[] = [
  {
    id: "home",
    label: "Home",
    path: "/admin",
    icon: Dashboard,
    exact: true,
  },
  {
    id: "users",
    label: "Users",
    path: "/admin/users",
    icon: UserMultiple,
  },
  {
    id: "clients",
    label: "Clients",
    path: "/admin/clients",
    icon: Api,
  },
  {
    id: "announcements",
    label: "Announcements",
    path: "/admin/announcements",
    icon: Bullhorn,
  },
  {
    id: "emails",
    label: "Emails",
    path: "/admin/emails",
    icon: Email,
  },
  {
    id: "audit-logs",
    label: "Audits",
    path: "/admin/audit-logs",
    icon: Activity,
  },
];

function isActiveRoute(pathname: string, item: AdminNavItem): boolean {
  if (item.exact) {
    return pathname === item.path;
  }

  return pathname === item.path || pathname.startsWith(`${item.path}/`);
}

function HeaderActions() {
  const user = useSelector(selectUser);
  const { openPanel } = useHeaderPanel();

  const {
    data: notifications = [],
    isLoading: isLoadingNotifications,
    refetch,
  } = useGetNotificationsQuery({
    unread: true,
    limit: 10,
    offset: 0,
  });

  const { data: unreadCount = 0 } = useGetUnreadNotificationsCountQuery();

  const safeUnreadCount = Number(unreadCount) || 0;

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };

  const handleOpenNotifications = () => {
    openPanel({
      title: "Notifications",
      className: "notifications-header-panel",
      content: (
        <NotificationsPanel
          notifications={notifications}
          isLoading={isLoadingNotifications}
          onMarkRead={() => {
            refetch();
          }}
        />
      ),
    });
  };

  return (
    <HeaderGlobalBar>
      <HeaderGlobalAction
        aria-label={
          safeUnreadCount > 0 ? `Notifications, ${safeUnreadCount} unread` : "Notifications"
        }
        tooltipAlignment="end"
        onClick={handleOpenNotifications}
      >
        <Notification size={20} />

        {safeUnreadCount > 0 && (
          <span className="admin-notification-badge">
            {safeUnreadCount > 99 ? "99+" : safeUnreadCount}
          </span>
        )}
      </HeaderGlobalAction>

      <HeaderGlobalAction
        aria-label={`Signed in as ${user?.username ?? "user"}`}
        tooltipAlignment="end"
      >
        <UserAvatarFilled size={20} />
      </HeaderGlobalAction>

      <HeaderGlobalAction aria-label="Logout" tooltipAlignment="end" onClick={handleLogout}>
        <Logout size={20} />
      </HeaderGlobalAction>
    </HeaderGlobalBar>
  );
}

function AdminSideNav() {
  const navigate = useNavigate();
  const location = useLocation();

  const navItems = useMemo(() => ADMIN_NAV_ITEMS, []);

  return (
    <SideNav isFixedNav expanded aria-label="Admin navigation" className="admin-layout__sidenav">
      <SideNavItems>
        {navItems.map((item) => {
          const Icon = item.icon;
          const active = isActiveRoute(location.pathname, item);

          return (
            <SideNavLink
              key={item.id}
              isActive={active}
              aria-current={active ? "page" : undefined}
              renderIcon={Icon}
              onClick={() => {
                if (location.pathname !== item.path) {
                  navigate(item.path);
                }
              }}
            >
              {item.label}
            </SideNavLink>
          );
        })}
      </SideNavItems>
    </SideNav>
  );
}

export default function AdminLayout() {
  const navigate = useNavigate();

  return (
    <ToastProvider>
      <HeaderPanelProvider>
        <div className="admin-layout">
          <Header aria-label="MOH Integrated Health Portal" className="admin-layout__header">
            <button
              type="button"
              className="admin-layout__brand"
              onClick={() => navigate("/admin")}
              aria-label="Go to admin home"
            >
              <img src={MohLogo} className="admin-layout__brand-logo" alt="" aria-hidden="true" />

              <HeaderName prefix="MOH">Integrated Health Portal</HeaderName>
            </button>

            <HeaderActions />
          </Header>

          <AdminSideNav />

          <div className="admin-layout__content-shell">
            <RouteBreadcrumbBar />

            <Content id="main-content" className="admin-layout__content">
              <Outlet />
            </Content>
          </div>
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
