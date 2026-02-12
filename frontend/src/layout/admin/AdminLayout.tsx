import { Notification, Logout, UserAvatarFilled } from "@carbon/react/icons";
import {
  HeaderGlobalAction,
  Content,
  Header,
  HeaderGlobalBar,
  HeaderName,
  SideNav,
  SideNavItems,
  SideNavLink,
} from "@carbon/react";
import { useSelector } from "react-redux";
import { useNavigate, useLocation, Outlet } from "react-router-dom";

import "./admin-layout.css";
import imagePath from "../../assets/logo.png";
import {
  HeaderPanelProvider,
  useHeaderPanel,
} from "../../components/header-panel/header-panel.context";
import { NotificationsPanel } from "../../components/notifications/notifications-panel.component";
import { ToastProvider } from "../../components/notifications/toast/ToastProvider";
import { API } from "../../lib/constants/api.constants";
import {
  useGetNotificationsQuery,
  useGetUnreadNotificationsCountQuery,
} from "../../store/api/notifications.api";
import { selectUser } from "../../store/auth/auth.selectors";

function HeaderActions() {
  const user = useSelector(selectUser);
  const { openPanel } = useHeaderPanel();

  const { data: notifications = [] } = useGetNotificationsQuery({
    unread: true,
    limit: 10,
    offset: 0,
  });

  const { data: unreadCount = 0 } = useGetUnreadNotificationsCountQuery();

  const handleLogout = () => {
    window.location.replace(API.auth.logout());
  };

  return (
    <HeaderGlobalBar>
      {/* 🔔 Notifications */}
      <HeaderGlobalAction
        aria-label="Notifications"
        tooltipAlignment="end"
        onClick={() => {
          openPanel({
            title: "Notifications",
            content: <NotificationsPanel notifications={notifications} />,
          });
        }}
      >
        <Notification size={20} />
        {unreadCount > 0 && <span className="notification-badge">{unreadCount}</span>}
      </HeaderGlobalAction>

      {/* 👤 User */}
      <HeaderGlobalAction
        aria-label={`Signed in as ${user?.username ?? "user"}`}
        tooltipAlignment="end"
      >
        <UserAvatarFilled size={20} />
      </HeaderGlobalAction>

      {/* 🚪 Logout */}
      <HeaderGlobalAction aria-label="Logout" tooltipAlignment="end" onClick={handleLogout}>
        <Logout size={20} />
      </HeaderGlobalAction>
    </HeaderGlobalBar>
  );
}

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <ToastProvider>
      <HeaderPanelProvider>
        {/* ================= Header ================= */}
        <Header aria-label="MOH Integrated Health Portal">
          <img src={imagePath} className={`moh-image-style`} />

          <HeaderName prefix="MOH" onClick={() => navigate("/admin")} style={{ cursor: "pointer" }}>
            Integrated Health Portal
          </HeaderName>

          <HeaderActions />
        </Header>

        {/* ================= Side Nav ================= */}
        <SideNav isFixedNav expanded aria-label="Admin navigation">
          <SideNavItems>
            <SideNavLink
              isActive={location.pathname === "/admin"}
              onClick={() => navigate("/admin")}
            >
              Home
            </SideNavLink>

            <SideNavLink
              isActive={location.pathname.startsWith("/admin/users")}
              onClick={() => navigate("/admin/users")}
            >
              Users
            </SideNavLink>

            <SideNavLink
              isActive={location.pathname.startsWith("/admin/clients")}
              onClick={() => navigate("/admin/clients")}
            >
              Clients
            </SideNavLink>

            <SideNavLink
              isActive={location.pathname.startsWith("/admin/audit-logs")}
              onClick={() => navigate("/admin/audit-logs")}
            >
              Audits
            </SideNavLink>
          </SideNavItems>
        </SideNav>

        {/* ================= Content ================= */}
        <Content style={{ marginLeft: 256, paddingTop: "3rem" }}>
          <Outlet />
        </Content>
      </HeaderPanelProvider>
    </ToastProvider>
  );
}
