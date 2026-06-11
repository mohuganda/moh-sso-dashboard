import { useRef } from "react";
import { useNavigate } from "react-router-dom";

import { API } from "@moh-sso/config";
import { PublicHeader } from "@moh-sso/ui";

export function ConnectedPublicHeader() {
  const navigate = useNavigate();
  const loginStartedRef = useRef(false);

  const handleLogin = () => {
    if (loginStartedRef.current) {
      return;
    }

    loginStartedRef.current = true;
    window.location.assign(API.auth.login());
  };

  return <PublicHeader onNavigateHome={() => navigate("/")} onLogin={handleLogin} />;
}
