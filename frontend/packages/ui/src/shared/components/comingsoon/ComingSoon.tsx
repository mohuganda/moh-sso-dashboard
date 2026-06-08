import { Bee } from "@carbon/react/icons";
import type React from "react";
import { Link } from "react-router-dom";

interface ComingSoonProps {
  title?: string;
  description?: string;
}

export const ComingSoon: React.FC<ComingSoonProps> = ({
  title = "Coming Soon",
  description = "This module is currently under development. Please check back later.",
}) => {
  return (
    <div style={{
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      justifyContent: "center",
      height: "100%",
      minHeight: "400px",
      textAlign: "center",
      padding: "2rem",
      backgroundColor: "var(--cds-background)",
    }}>
      <Bee size={64} style={{ marginBottom: "1.5rem", color: "var(--cds-interactive-01)" }} />
      <h2 style={{ marginBottom: "1rem" }}>{title}</h2>
      <p style={{ marginBottom: "2rem", maxWidth: "400px", color: "var(--cds-text-secondary)" }}>
        {description}
      </p>
      <Link 
        to="/apps" 
        style={{ 
          textDecoration: "none", 
          color: "var(--cds-link-01)",
          fontWeight: 600 
        }}
      >
        Return to Dashboard
      </Link>
    </div>
  );
};
