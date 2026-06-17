import { Bee } from "@carbon/react/icons";
import type React from "react";
import { Link } from "react-router-dom";

import "./ComingSoon.scss";

interface ComingSoonProps {
  title?: string;
  description?: string;
}

export const ComingSoon: React.FC<ComingSoonProps> = ({
  title = "Coming Soon",
  description = "This module is currently under development. Please check back later.",
}) => {
  return (
    <div className="moh-coming-soon">
      <Bee size={64} className="moh-coming-soon__icon" />
      <h2 className="moh-coming-soon__title">{title}</h2>
      <p className="moh-coming-soon__description">{description}</p>
      <Link to="/apps" className="moh-coming-soon__link">
        Return to Dashboard
      </Link>
    </div>
  );
};
