import { Login } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";
import { useNavigate } from "react-router-dom";

import { API, BRANDING } from "@moh-sso/config";
import "./public-header.css";

import imagePath from "../../../assets/logo.png";

const PublicHeader: React.FC = () => {
  const navigate = useNavigate();

  const handleLogin = () => {
    window.location.replace(API.auth.login());
  };
  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      <img src={imagePath} className={`moh-image-style`} />
      <HeaderName prefix="MOH" onClick={() => navigate("/")} style={{ cursor: "pointer" }}>
        {BRANDING.title}
      </HeaderName>

      <HeaderGlobalBar>
        <HeaderGlobalAction
          aria-label="login"
          tooltipAlignment="end"
          onClick={() => {
            handleLogin();
          }}
        >
          <Login size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
