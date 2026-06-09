import type { ReactNode } from "react";

import "./page-shell.scss";

type PageShellProps = {
  title?: string;
  subtitle?: string;
  actions?: ReactNode;
  children: ReactNode;
};

export function PageShell({ title, subtitle, actions, children }: PageShellProps) {
  return (
    <div className="moh-page-shell">
      {(title || subtitle || actions) && (
        <header className="moh-page-shell__header">
          <div>
            {title && <h3 className="moh-page-shell__title">{title}</h3>}
            {subtitle && <p className="moh-page-shell__subtitle">{subtitle}</p>}
          </div>

          {actions && <div className="moh-page-shell__actions">{actions}</div>}
        </header>
      )}

      <div className="moh-page-shell__content">{children}</div>
    </div>
  );
}
