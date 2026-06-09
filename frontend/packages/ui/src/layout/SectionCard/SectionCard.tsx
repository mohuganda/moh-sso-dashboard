import type { ReactNode } from "react";
import { Tile } from "@carbon/react";

import "./section-card.scss";

type SectionCardProps = {
  title?: string;
  description?: string;
  icon?: React.ElementType;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
};

export function SectionCard({
  title,
  description,
  icon: Icon,
  actions,
  children,
  className = "",
}: SectionCardProps) {
  return (
    <Tile className={`moh-section-card ${className}`}>
      {(title || description || actions) && (
        <div className="moh-section-card__header">
          <div className="moh-section-card__title-wrap">
            {Icon && (
              <span className="moh-section-card__icon" aria-hidden="true">
                <Icon size={18} />
              </span>
            )}

            <div>
              {title && <h4 className="moh-section-card__title">{title}</h4>}
              {description && <p className="moh-section-card__description">{description}</p>}
            </div>
          </div>

          {actions && <div className="moh-section-card__actions">{actions}</div>}
        </div>
      )}

      <div className="moh-section-card__content">{children}</div>
    </Tile>
  );
}
