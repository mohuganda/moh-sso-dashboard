import { Login } from "@carbon/react/icons";
import {
  Header,
  HeaderName,
  HeaderGlobalBar,
  HeaderGlobalAction,
  SkipToContent,
} from "@carbon/react";
import type React from "react";

import "./public-header.css";
import imagePath from "../assets/logo.png";

type PublicHeaderProps = {
  onNavigateHome?: () => void;
  onLogin?: () => void;
};

const PublicHeader: React.FC<PublicHeaderProps> = ({ onNavigateHome, onLogin }) => {
  return (
    <Header aria-label="MOH Integrated Health Portal">
      <SkipToContent />

      <img src={imagePath} className="moh-image-style" alt="Ministry of Health logo" />

      <HeaderName prefix="MOH" onClick={onNavigateHome} className="moh-header-name">
        Integrated Health Portal
      </HeaderName>

      <HeaderGlobalBar>
        <HeaderGlobalAction aria-label="Login" tooltipAlignment="end" onClick={onLogin}>
          <Login size={20} />
        </HeaderGlobalAction>
      </HeaderGlobalBar>
    </Header>
  );
};

export default PublicHeader;
