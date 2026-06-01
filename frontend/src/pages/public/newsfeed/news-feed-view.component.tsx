import { useMemo } from "react";
import {
  Information,
  Time,
  ChevronRight,
  Pin,
  Help,
  Document,
  WarningAlt,
} from "@carbon/react/icons";
import { Tile, Link, Tag, SkeletonText } from "@carbon/react";
import { useNavigate } from "react-router-dom";

import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import "./news-feed.css";
import type {
  Announcement,
  AnnouncementLevel,
  AnnouncementStatus,
} from "../../../store/types/announcements.types";

const CASE_REPORTING = [
  {
    label: "VHF Case Investigation",
    href: "http://localhost:3001/vhf-cif",
  },
  {
    label: "M-Pox Case Investigation",
    href: "http://localhost:3001/mpox-cif",
  },
  {
    label: "Measles Case Investigation",
    href: "http://localhost:3001/measles_cif",
  },
  {
    label: "Polio Case Investigation",
    href: "http://localhost:3001/polio-cif",
  },
  {
    label: "Other Alerts",
    href: "#",
  },
];

const QUICK_LINKS = [
  {
    label: "Support",
    href: "/support",
    icon: Help,
  },
  {
    label: "FAQ",
    href: "/faq",
    icon: Help,
  },
  {
    label: "Guidelines & Documents",
    href: "https://mohuganda.github.io/digital-guidelines-and-documentation",
    icon: Document,
  },
];

type CarbonTagType =
  | "red"
  | "magenta"
  | "purple"
  | "blue"
  | "cyan"
  | "teal"
  | "green"
  | "gray"
  | "cool-gray"
  | "warm-gray"
  | "high-contrast"
  | "outline";

type TagConfig = {
  label: string;
  type: CarbonTagType;
};

function mapLevelTag(level: AnnouncementLevel): TagConfig {
  switch (level) {
    case "CRITICAL":
      return { label: "Critical", type: "red" };
    case "WARNING":
      return { label: "Warning", type: "warm-gray" };
    case "SUCCESS":
      return { label: "Success", type: "green" };
    case "INFO":
    default:
      return { label: "Info", type: "blue" };
  }
}

function mapStatusTag(status: AnnouncementStatus): TagConfig | null {
  switch (status) {
    case "PUBLISHED":
      return null;
    case "SCHEDULED":
      return { label: "Scheduled", type: "purple" };
    case "DRAFT":
      return { label: "Draft", type: "cool-gray" };
    case "ARCHIVED":
      return { label: "Archived", type: "gray" };
    default:
      return null;
  }
}

function mapCustomTag(tag?: string | null): TagConfig | null {
  const label = tag?.trim();

  if (!label) {
    return null;
  }

  switch (label.toLowerCase()) {
    case "critical":
    case "urgent":
    case "alert":
    case "action":
      return { label, type: "red" };

    case "scheduled":
    case "maintenance":
    case "warning":
      return { label, type: "warm-gray" };

    case "event":
    case "new":
    case "update":
      return { label, type: "blue" };

    case "success":
    case "resolved":
      return { label, type: "green" };

    default:
      return { label, type: "gray" };
  }
}

function getAnnouncementTime(item: Announcement): number {
  const rawDate = item.publish_at || item.created_at;
  const time = new Date(rawDate).getTime();

  return Number.isNaN(time) ? 0 : time;
}

function formatTimestamp(dateString?: string | null): string {
  if (!dateString) {
    return "Recently posted";
  }

  const date = new Date(dateString);

  if (Number.isNaN(date.getTime())) {
    return "Recently posted";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatRelativeTimestamp(dateString?: string | null): string {
  if (!dateString) {
    return "Recently posted";
  }

  const date = new Date(dateString);

  if (Number.isNaN(date.getTime())) {
    return "Recently posted";
  }

  const diffMs = Date.now() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) {
    return "Just now";
  }

  if (diffMinutes < 60) {
    return `${diffMinutes} minute${diffMinutes === 1 ? "" : "s"} ago`;
  }

  if (diffHours < 24) {
    return `${diffHours} hour${diffHours === 1 ? "" : "s"} ago`;
  }

  if (diffDays < 7) {
    return `${diffDays} day${diffDays === 1 ? "" : "s"} ago`;
  }

  return formatTimestamp(dateString);
}

function isExternalHref(href: string): boolean {
  return /^https?:\/\//i.test(href);
}

function isPlaceholderHref(href: string): boolean {
  return href.trim() === "#";
}

interface AppLinkProps {
  href: string;
  className?: string;
  children: React.ReactNode;
}

function AppLink({ href, className, children }: AppLinkProps) {
  const navigate = useNavigate();

  if (isPlaceholderHref(href)) {
    return (
      <button type="button" className={className} disabled>
        {children}
      </button>
    );
  }

  if (isExternalHref(href)) {
    return (
      <Link href={href} target="_blank" rel="noopener noreferrer" className={className}>
        {children}
      </Link>
    );
  }

  return (
    <button type="button" className={className} onClick={() => navigate(href)}>
      {children}
    </button>
  );
}

function AnnouncementLink({ item }: { item: Announcement }) {
  if (!item.link_url?.trim()) {
    return null;
  }

  return (
    <div className="feed-link-row">
      <AppLink href={item.link_url} className="feed-link">
        Read more
        <ChevronRight size={16} />
      </AppLink>
    </div>
  );
}

interface AnnouncementCardProps {
  item: Announcement;
  showStatusTags?: boolean;
}

function AnnouncementCard({ item, showStatusTags = false }: AnnouncementCardProps) {
  const levelTag = mapLevelTag(item.level);
  const customTag = mapCustomTag(item.tag);
  const statusTag = showStatusTags ? mapStatusTag(item.status) : null;
  const timestampSource = item.publish_at || item.created_at;

  return (
    <Tile className={`feed-item ${item.is_pinned ? "feed-item--pinned" : ""}`}>
      <div className="feed-item__top">
        <div className="feed-item__title-row">
          <span
            className={`feed-item__level-icon feed-item__level-icon--${item.level.toLowerCase()}`}
          >
            <Information size={16} />
          </span>

          <h4 className="feed-item__title">{item.title}</h4>
        </div>

        <div className="feed-item__tags">
          {item.is_pinned && (
            <Tag type="warm-gray" size="sm">
              <span className="feed-tag-with-icon">
                <Pin size={12} />
                Pinned
              </span>
            </Tag>
          )}

          <Tag type={levelTag.type} size="sm">
            {levelTag.label}
          </Tag>

          {customTag && (
            <Tag type={customTag.type} size="sm">
              {customTag.label}
            </Tag>
          )}

          {statusTag && (
            <Tag type={statusTag.type} size="sm">
              {statusTag.label}
            </Tag>
          )}
        </div>
      </div>

      {item.summary && <p className="feed-summary">{item.summary}</p>}

      <p className="feed-message">{item.message}</p>

      <AnnouncementLink item={item} />

      <div className="feed-timestamp" title={formatTimestamp(timestampSource)}>
        <Time size={14} />
        <span>{formatRelativeTimestamp(timestampSource)}</span>
      </div>
    </Tile>
  );
}

function LoadingFeed() {
  return (
    <div className="feed-section-list" aria-label="Loading announcements">
      {Array.from({ length: 3 }).map((_, index) => (
        <Tile className="feed-item feed-item--loading" key={index}>
          <SkeletonText width="45%" />
          <SkeletonText paragraph lineCount={3} />
          <SkeletonText width="25%" />
        </Tile>
      ))}
    </div>
  );
}

interface FeedSectionProps {
  title: string;
  subtitle?: string;
  icon: React.ElementType;
  count?: number;
  children: React.ReactNode;
}

function FeedSection({ title, subtitle, icon: Icon, count, children }: FeedSectionProps) {
  return (
    <section className="feed-section">
      <div className="feed-section-header">
        <div>
          <div className="feed-section-title">
            <Icon size={18} />
            <h4>{title}</h4>

            {typeof count === "number" && (
              <Tag type="gray" size="sm">
                {count}
              </Tag>
            )}
          </div>

          {subtitle && <p className="feed-section-subtitle">{subtitle}</p>}
        </div>
      </div>

      <div className="feed-section-list">{children}</div>
    </section>
  );
}

export interface NewsFeedViewProps {
  announcements: Announcement[];
  isLoading: boolean;
  isFetching?: boolean;
  hasError: boolean;
  errorMessage?: string;
  onRetry: () => void;
  showStatusTags?: boolean;
}

export default function NewsFeedView({
  announcements,
  isLoading,
  isFetching = false,
  hasError,
  errorMessage = "Unable to load announcements at the moment.",
  onRetry,
  showStatusTags = false,
}: NewsFeedViewProps) {
  const navigate = useNavigate();

  const sortedAnnouncements = useMemo(() => {
    return [...announcements].sort((a, b) => {
      if (a.is_pinned !== b.is_pinned) {
        return a.is_pinned ? -1 : 1;
      }

      return getAnnouncementTime(b) - getAnnouncementTime(a);
    });
  }, [announcements]);

  const pinnedAnnouncements = useMemo(
    () => sortedAnnouncements.filter((item) => item.is_pinned),
    [sortedAnnouncements],
  );

  const regularAnnouncements = useMemo(
    () => sortedAnnouncements.filter((item) => !item.is_pinned),
    [sortedAnnouncements],
  );

  const shouldShowInitialLoading = isLoading && announcements.length === 0;
  const shouldShowRefreshing = isFetching && announcements.length > 0;

  const isEmpty =
    !isLoading &&
    !isFetching &&
    !hasError &&
    pinnedAnnouncements.length === 0 &&
    regularAnnouncements.length === 0;

  return (
    <div className="page-container news-page">
      <header className="page-header news-page__header">
        <div>
          <h3 className="page-title">News & Updates</h3>
          <p className="page-subtitle">Latest system updates, announcements, and notices.</p>
        </div>

        {shouldShowRefreshing && (
          <Tag type="blue" size="sm">
            Refreshing
          </Tag>
        )}
      </header>

      <section className="page-content news-page__content">
        <main className="news-feed">
          {shouldShowInitialLoading && <LoadingFeed />}

          {hasError && !isLoading && (
            <ErrorState
              title="Failed to load announcements"
              description={errorMessage}
              primaryAction={{
                label: "Retry",
                onClick: onRetry,
              }}
              secondaryAction={{
                label: "Contact support",
                onClick: () => navigate("/support"),
              }}
            />
          )}

          {isEmpty && (
            <EmptyState
              title="No announcements yet"
              description="System updates and important notices will appear here when available."
            />
          )}

          {!hasError && !shouldShowInitialLoading && pinnedAnnouncements.length > 0 && (
            <FeedSection
              title="Pinned Announcements"
              subtitle="Important notices highlighted for quick access."
              icon={Pin}
              count={pinnedAnnouncements.length}
            >
              {pinnedAnnouncements.map((item) => (
                <AnnouncementCard key={item.id} item={item} showStatusTags={showStatusTags} />
              ))}
            </FeedSection>
          )}

          {!hasError && !shouldShowInitialLoading && regularAnnouncements.length > 0 && (
            <FeedSection
              title="All Announcements"
              icon={Information}
              count={regularAnnouncements.length}
            >
              {regularAnnouncements.map((item) => (
                <AnnouncementCard key={item.id} item={item} showStatusTags={showStatusTags} />
              ))}
            </FeedSection>
          )}
        </main>

        <aside className="news-sidebar">
          <Tile className="news-sidebar__tile">
            <div className="news-sidebar__heading">
              <WarningAlt size={18} />
              <h4>Case Reporting</h4>
            </div>

            <ul className="sidebar-link-list">
              {CASE_REPORTING.map((item) => (
                <li key={item.label}>
                  <AppLink href={item.href} className="sidebar-link">
                    <span>{item.label}</span>
                    <ChevronRight size={16} />
                  </AppLink>
                </li>
              ))}
            </ul>
          </Tile>

          <Tile className="news-sidebar__tile">
            <div className="news-sidebar__heading">
              <Document size={18} />
              <h4>Quick Links</h4>
            </div>

            <ul className="sidebar-link-list">
              {QUICK_LINKS.map(({ label, href, icon: Icon }) => (
                <li key={label}>
                  <AppLink href={href} className="sidebar-link">
                    <span className="sidebar-link__label">
                      <Icon size={16} />
                      {label}
                    </span>
                    <ChevronRight size={16} />
                  </AppLink>
                </li>
              ))}
            </ul>
          </Tile>

          <Tile className="news-sidebar__tile news-sidebar__help">
            <div className="news-sidebar__heading">
              <Help size={18} />
              <h4>Need Help?</h4>
            </div>

            <p className="sidebar-help-text">
              Reach out to support for account issues, access requests, or reporting assistance.
            </p>

            <AppLink href="/support" className="feed-link">
              Contact support
              <ChevronRight size={16} />
            </AppLink>
          </Tile>
        </aside>
      </section>
    </div>
  );
}
