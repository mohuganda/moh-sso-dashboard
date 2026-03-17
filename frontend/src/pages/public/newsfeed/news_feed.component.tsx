import { useMemo } from "react";
import { Information, Time, ChevronRight, Pin } from "@carbon/react/icons";
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
import { useListMyAnnouncementsQuery } from "../../../store/api/announcement.api";

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

function mapLevelTag(level: AnnouncementLevel): { label: string; type: CarbonTagType } {
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

function mapStatusTag(status: AnnouncementStatus): { label: string; type: CarbonTagType } | null {
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

function mapCustomTag(tag?: string | null): { label: string; type: CarbonTagType } | null {
  if (!tag?.trim()) return null;

  const normalized = tag.trim().toLowerCase();

  switch (normalized) {
    case "critical":
    case "urgent":
    case "alert":
    case "action":
      return { label: tag, type: "red" };

    case "scheduled":
    case "maintenance":
    case "warning":
      return { label: tag, type: "warm-gray" };

    case "event":
    case "new":
    case "update":
      return { label: tag, type: "blue" };

    case "success":
    case "resolved":
      return { label: tag, type: "green" };

    default:
      return { label: tag, type: "gray" };
  }
}

function formatTimestamp(dateString: string): string {
  const date = new Date(dateString);

  if (Number.isNaN(date.getTime())) {
    return "Recently posted";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function renderAnnouncementLink(item: Announcement) {
  if (!item.link_url) return null;

  return (
    <div className="feed-link-row">
      <Link href={item.link_url} target="_blank" rel="noopener noreferrer" className="feed-link">
        Read more
        <ChevronRight size={16} />
      </Link>
    </div>
  );
}

export default function NewsFeedPage() {
  const navigate = useNavigate();

  const {
    data: announcements = [],
    isLoading,
    isFetching,
    isError,
    error,
    refetch,
  } = useListMyAnnouncementsQuery({ limit: 20, offset: 0 });

  const sortedAnnouncements = useMemo(() => {
    return [...announcements].sort((a, b) => {
      if (a.is_pinned !== b.is_pinned) {
        return a.is_pinned ? -1 : 1;
      }

      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });
  }, [announcements]);

  const isEmpty = !isLoading && !isFetching && sortedAnnouncements.length === 0;

  const errorMessage =
    typeof error === "object" && error !== null && "status" in error
      ? "Unable to load announcements at the moment."
      : "Something went wrong while loading announcements.";

  return (
    <div className="page-container">
      <header className="page-header">
        <h3 className="page-title">News & Updates</h3>
        <p className="page-subtitle">Latest system updates, announcements, and notices.</p>
      </header>

      <section className="page-content">
        <main className="news-feed">
          {(isLoading || isFetching) && <SkeletonText paragraph lineCount={4} />}

          {isError && !isLoading && (
            <ErrorState
              title="Failed to load announcements"
              description={errorMessage}
              primaryAction={{
                label: "Retry",
                onClick: () => {
                  refetch();
                },
              }}
              secondaryAction={{
                label: "Contact support",
                onClick: () => navigate("/support"),
              }}
            />
          )}

          {!isError && isEmpty && (
            <EmptyState
              title="No announcements yet"
              description="System updates and important notices will appear here when available."
            />
          )}

          {!isError &&
            !isLoading &&
            sortedAnnouncements.map((item) => {
              const levelTag = mapLevelTag(item.level);
              const customTag = mapCustomTag(item.tag);
              const statusTag = mapStatusTag(item.status);

              return (
                <Tile key={item.id} className="feed-item">
                  <div className="feed-header">
                    <Information size={16} />
                    <h4>{item.title}</h4>

                    {item.is_pinned && (
                      <Tag type="warm-gray" size="sm">
                        <span style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                          <Pin size={12} />
                          Pinned
                        </span>
                      </Tag>
                    )}

                    <Tag type={levelTag.type} size="sm">
                      {levelTag.label}
                    </Tag>

                    {customTag ? (
                      <Tag type={customTag.type} size="sm">
                        {customTag.label}
                      </Tag>
                    ) : null}

                    {statusTag ? (
                      <Tag type={statusTag.type} size="sm">
                        {statusTag.label}
                      </Tag>
                    ) : null}
                  </div>

                  {item.summary ? <p className="feed-summary">{item.summary}</p> : null}

                  <p className="feed-message">{item.message}</p>

                  {renderAnnouncementLink(item)}

                  <div className="feed-timestamp">
                    <Time size={14} />
                    <span>{formatTimestamp(item.publish_at || item.created_at)}</span>
                  </div>
                </Tile>
              );
            })}
        </main>

        <aside className="news-sidebar">
          <Tile>
            <h4>Case Reporting</h4>

            <ul className="case-reporting-list">
              {CASE_REPORTING.map((item) => (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="case-reporting-link"
                  >
                    {item.label}
                    <ChevronRight size={16} />
                  </Link>
                </li>
              ))}
            </ul>
          </Tile>
        </aside>
      </section>
    </div>
  );
}
