import { ChevronLeft } from "@carbon/react/icons";
import { Button } from "@carbon/react";

import { AppBreadcrumbs, type AppBreadcrumbItem } from "./AppBreadcrumbs";

import "./breadcrumb-bar.scss";

type BreadcrumbBarProps = {
  items: AppBreadcrumbItem[];
  onNavigate?: (href: string) => void;
  onBack?: () => void;
  showBackButton?: boolean;
  className?: string;
};

export function BreadcrumbBar({
  items,
  onNavigate,
  onBack,
  showBackButton = true,
  className = "",
}: BreadcrumbBarProps) {
  return (
    <div className={["moh-breadcrumb-bar", className].filter(Boolean).join(" ")}>
      {showBackButton && (
        <div className="moh-breadcrumb-bar__back">
          <Button
            kind="ghost"
            size="sm"
            hasIconOnly
            renderIcon={ChevronLeft}
            iconDescription="Go back"
            tooltipPosition="bottom"
            onClick={onBack}
          />
        </div>
      )}

      <div className="moh-breadcrumb-bar__content">
        <AppBreadcrumbs items={items} onNavigate={onNavigate} />
      </div>
    </div>
  );
}
