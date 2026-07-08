// src/components/footer/PublicFooter.tsx
import "./PublicFooter.scss";

export type FooterVersionInfo = {
  frontendVersion?: string;
  backendVersion?: string;
  backendUnavailable?: boolean;
};

function formatVersion(version?: string) {
  if (!version) {
    return undefined;
  }

  return version === "dev" || version.startsWith("v") ? version : `v${version}`;
}

export function PublicFooter({
  frontendVersion,
  backendVersion,
  backendUnavailable = false,
}: FooterVersionInfo = {}) {
  const frontendLabel = formatVersion(frontendVersion);
  const backendLabel = backendUnavailable ? "unavailable" : formatVersion(backendVersion);
  const hasVersionInfo = frontendLabel || backendLabel;

  return (
    <footer className="public-footer">
      <div className="footer-bottom">
        <span>© {new Date().getFullYear()} Ministry of Health – Uganda</span>
        {hasVersionInfo && (
          <span className="footer-version" aria-label="Application version information">
            {frontendLabel && <span>Frontend {frontendLabel}</span>}
            {frontendLabel && backendLabel && <span aria-hidden="true">·</span>}
            {backendLabel && <span>Backend {backendLabel}</span>}
          </span>
        )}
      </div>
    </footer>
  );
}
