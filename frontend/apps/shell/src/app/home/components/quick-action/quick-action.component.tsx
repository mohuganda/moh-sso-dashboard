import { useId, type ReactNode } from "react";
import { ArrowRight } from "@carbon/react/icons";
import { useNavigate } from "react-router-dom";
import "./quick-action.scss";

type QuickActionTone = "default" | "primary" | "warning" | "danger";
type QuickActionProps = {
  icon: ReactNode;
  label: string;
  description?: string;
  href?: string;
  onClick?: () => void;
  tone?: QuickActionTone;
  disabled?: boolean;
};

export function QuickAction({
  icon,
  label,
  description,
  href,
  onClick,
  tone = "default",
  disabled = false,
}: QuickActionProps) {
  const navigate = useNavigate();
  const descriptionId = useId();
  return (
    <button
      type="button"
      className={`quick-action quick-action--${tone}`}
      disabled={disabled || (!onClick && !href)}
      aria-label={label}
      aria-describedby={description ? descriptionId : undefined}
      onClick={() => {
        if (onClick) onClick();
        else if (href) navigate(href);
      }}
    >
      <span className="quick-action__icon" aria-hidden="true">
        {icon}
      </span>
      <span className="quick-action__content">
        <span className="quick-action__label">{label}</span>
        {description && (
          <span id={descriptionId} className="quick-action__description">
            {description}
          </span>
        )}
      </span>
      <ArrowRight size={16} className="quick-action__arrow" aria-hidden="true" />
    </button>
  );
}
