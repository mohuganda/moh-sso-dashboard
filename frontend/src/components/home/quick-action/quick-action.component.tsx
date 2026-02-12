import React from "react";
import { Tile } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import "./quick-action.css";

type QuickActionTone = "default" | "primary" | "warning" | "danger";

type QuickActionProps = {
  icon: React.ReactNode;
  label: string;
  description?: string;

  /** Route navigation (optional) */
  href?: string;

  /** Custom action (e.g. open header panel) */
  onClick?: () => void;

  /** Visual emphasis */
  tone?: QuickActionTone;

  /** Disable interaction */
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

  const isInteractive = !disabled && (onClick || href);

  const handleClick = () => {
    if (disabled) return;

    if (onClick) {
      onClick();
      return;
    }

    if (href) {
      navigate(href);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isInteractive) return;

    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleClick();
    }
  };

  return (
    <Tile
      role={isInteractive ? "button" : undefined}
      tabIndex={isInteractive ? 0 : -1}
      aria-disabled={disabled}
      className={["quick-action", `quick-action--${tone}`, disabled && "quick-action--disabled"]
        .filter(Boolean)
        .join(" ")}
      onClick={isInteractive ? handleClick : undefined}
      onKeyDown={handleKeyDown}
    >
      <div className="quick-action__icon" aria-hidden>
        {icon}
      </div>

      <div className="quick-action__content">
        <strong className="quick-action__label">{label}</strong>

        {description && <p className="quick-action__description">{description}</p>}
      </div>
    </Tile>
  );
}
