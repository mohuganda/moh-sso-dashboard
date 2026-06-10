import { Login } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";
import { useRef } from "react";
import { useNavigate } from "react-router-dom";

import { API } from "@moh-sso/config";
import "./public-header.css";
import imagePath from "../../../assets/logo.png";

const PublicHeader: React.FC = () => {
  const navigate = useNavigate();
  const loginStartedRef = useRef(false);

  const handleLogin = () => {
    // Prevent double redirects/clicks from overwriting oauth_state/pkce cookies.
    if (loginStartedRef.current) return;

    loginStartedRef.current = true;
    window.location.assign(API.auth.login());
  };

  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      <img src={imagePath} className="moh-image-style" alt="Ministry of Health logo" />

      <HeaderName prefix="MOH" onClick={() => navigate("/")} style={{ cursor: "pointer" }}>
        Integrated Health Portal
      </HeaderName>

      <HeaderGlobalBar>
        <HeaderGlobalAction aria-label="Login" tooltipAlignment="end" onClick={handleLogin}>
          <Login size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
