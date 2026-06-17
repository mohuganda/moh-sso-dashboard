import { Link } from "@carbon/react";

import "./app-breadcrumbs.scss";

export type AppBreadcrumbItem = {
  label: string;
  href?: string;
  current?: boolean;
};

export type AppBreadcrumbsProps = {
  items: AppBreadcrumbItem[];
  className?: string;
  separator?: string;
  onNavigate?: (href: string) => void;
};

export function AppBreadcrumbs({
  items,
  className = "",
  separator = ">",
  onNavigate,
}: AppBreadcrumbsProps) {
  const visibleItems = items.filter((item) => item.label.trim().length > 0);

  if (visibleItems.length === 0) {
    return null;
  }

  return (
    <nav
      className={["moh-breadcrumbs", className].filter(Boolean).join(" ")}
      aria-label="Breadcrumb"
    >
      <ol className="moh-breadcrumbs__list">
        {visibleItems.map((item, index) => {
          const isLast = index === visibleItems.length - 1;
          const isCurrent = item.current ?? isLast;

          return (
            <li
              key={`${item.label}-${index}`}
              className="moh-breadcrumbs__item"
              aria-current={isCurrent ? "page" : undefined}
            >
              {index > 0 && (
                <span className="moh-breadcrumbs__separator" aria-hidden="true">
                  {separator}
                </span>
              )}

              {item.href && !isCurrent ? (
                <Link
                  href={item.href}
                  className="moh-breadcrumbs__link"
                  onClick={(event) => {
                    if (!onNavigate) {
                      return;
                    }

                    event.preventDefault();
                    onNavigate(item.href as string);
                  }}
                >
                  {item.label}
                </Link>
              ) : (
                <span className="moh-breadcrumbs__current">{item.label}</span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
