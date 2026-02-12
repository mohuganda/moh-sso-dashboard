import { Content } from "@carbon/react";
import { Outlet } from "react-router-dom";

import { PublicFooter } from "../../components/footer/PublicFooter";
import PublicHeader from "../../components/header/PublicHeader.component";

export default function PublicLayout() {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
      }}
    >
      {/* 🔹 Global Header */}
      <PublicHeader />

      {/* 🔹 SideNav + Content row */}
      <div
        style={{
          display: "flex",
          flex: 1,
        }}
      >
        {/* 🔹 Main content */}
        <Content
          id="main-content"
          style={{
            marginTop: "3rem", // Carbon header offset
            flex: 1,
            background: "#f9fafb",
          }}
        >
          <Outlet />
        </Content>
      </div>

      {/* 🔹 Footer */}
      <PublicFooter />
    </div>
  );
}
