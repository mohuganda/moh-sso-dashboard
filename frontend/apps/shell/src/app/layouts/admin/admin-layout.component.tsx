import { useMemo } from "react";
import type { ComponentType } from "react";
import {
  Content,
  Header,
  HeaderGlobalAction,
  HeaderGlobalBar,
  HeaderName,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import {
  Activity,
  Api,
  Bullhorn,
  Dashboard,
  Email,
  Logout,
  Notification,
  UserAvatarFilled,
  UserMultiple,
  UserRole,
} from "@carbon/react/icons";
import { useSelector } from "react-redux";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import {
  HeaderPanelProvider,
  NotificationsPanel,
  PublicFooter,
  ToastProvider,
  useHeaderPanel,
} from "@moh-sso/ui";

import { API } from "@moh-sso/config";
import {
  useDeleteNotificationMutation,
  useGetNotificationsQuery,
  useGetUnreadNotificationsCountQuery,
  useMarkNotificationAsReadMutation,
} from "@moh-sso/api";
import { PERMISSIONS, selectUser, useAuthorization } from "@moh-sso/auth";
import type { Permission } from "@moh-sso/auth";

import { RouteBreadcrumbBar } from "@/app/navigation/RouteBreadcrumbBar";

import "./admin-layout.scss";

import imagePath from "../../../assets/logo.png";

type CarbonIconComponent = ComponentType<{
  size?: number | string;
  className?: string;
  "aria-label"?: string;
}>;

type AdminNavItem = {
  id: string;
  label: string;
  path: string;
  icon: CarbonIconComponent;
  exact?: boolean;
  requiredPermission?: Permission;
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
  {
    id: "rbac",
    label: "RBAC",
    path: "/admin/rbac",
    icon: UserRole,
    requiredPermission: PERMISSIONS.rbacRolesWrite,
  },
];

function isActiveRoute(pathname: string, item: AdminNavItem): boolean {
  if (item.exact) {
    return pathname === item.path || pathname === "portal/admin/home";
  }

  return pathname === item.path || pathname.startsWith(`${item.path}/`);
}

function HeaderActions() {
  const user = useSelector(selectUser);
  const { openPanel } = useHeaderPanel();

  const {
    data: notifications = [],
    isLoading: isLoadingNotifications,
    refetch: refetchUnreadNotifications,
  } = useGetNotificationsQuery({
    unread: true,
    limit: 10,
    offset: 0,
  });
  const {
    data: recentNotifications = [],
    isLoading: isLoadingRecentNotifications,
    refetch: refetchRecentNotifications,
  } = useGetNotificationsQuery({
    limit: 20,
    offset: 0,
  });

  const { data: unreadCount = 0 } = useGetUnreadNotificationsCountQuery();
  const [markNotificationAsRead] = useMarkNotificationAsReadMutation();
  const [deleteNotification] = useDeleteNotificationMutation();

  const safeUnreadCount = Number(unreadCount) || 0;

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };

  const handleOpenNotifications = () => {
    openPanel({
      title: "Notifications",
      content: (
        <div className="notifications-header-panel">
          <NotificationsPanel
            notifications={notifications}
            recentNotifications={recentNotifications}
            isLoading={isLoadingNotifications}
            isLoadingRecent={isLoadingRecentNotifications}
            onMarkRead={(id) => {
              void markNotificationAsRead(id).then(() => {
                void refetchUnreadNotifications();
                void refetchRecentNotifications();
              });
            }}
            onDelete={(id) => {
              void deleteNotification(id).then(() => {
                void refetchUnreadNotifications();
                void refetchRecentNotifications();
              });
            }}
          />
        </div>
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
  const { can } = useAuthorization();

  const navItems = useMemo(
    () =>
      ADMIN_NAV_ITEMS.filter((item) => !item.requiredPermission || can(item.requiredPermission)),
    [can],
  );

  return (
    <SideNav isFixedNav expanded aria-label="Admin navigation" className="admin-layout__sidenav">
      <SideNavItems>
        {navItems.map((item) => {
          const active = isActiveRoute(location.pathname, item);

          return (
            <SideNavLink
              key={item.id}
              isActive={active}
              aria-current={active ? "page" : undefined}
              renderIcon={item.icon}
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
            <div className="admin-layout__brand">
              <button
                type="button"
                className="admin-layout__brand-logo-button"
                onClick={() => navigate("/admin")}
                aria-label="Go to admin home"
              >
                <img
                  src={imagePath}
                  className="admin-layout__brand-logo"
                  alt=""
                  aria-hidden="true"
                />
              </button>

              <HeaderName
                prefix="MOH"
                onClick={() => navigate("/admin")}
                className="admin-layout__brand-name"
              >
                Integrated Health Portal
              </HeaderName>
            </div>

            <HeaderActions />
          </Header>

          <AdminSideNav />

          <div className="admin-layout__content-shell">
            <RouteBreadcrumbBar />

            <Content id="main-content" className="admin-layout__content">
              <Outlet />
            </Content>
          </div>
          <PublicFooter />
        </div>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
