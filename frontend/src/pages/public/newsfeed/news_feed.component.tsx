import { Information, Time, Add, ChevronRight } from "@carbon/react/icons";
import { Tile, Link, Tag, SkeletonText } from "@carbon/react";
import { useNavigate } from "react-router-dom";

import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import "./news-feed.css";
import { useListAnnouncementsQuery } from "../../../store/api/announcement.api";
import type { Announcement } from "../../../store/types/announcements.types";

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

function mapTag(tag: string): { label: string; type: CarbonTagType } {
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

    case "info":
    default:
      return { label: tag || "Info", type: "gray" };
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
  } = useListAnnouncementsQuery({ limit: 20 });

  const isEmpty = !isLoading && !isFetching && announcements.length === 0;

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
              description="System updates and important notices will appear here."
              primaryAction={{
                label: "Create announcement",
                icon: Add,
                onClick: () => navigate("/admin/announcements/new"),
              }}
            />
          )}

          {!isError &&
            !isLoading &&
            announcements.map((item) => {
              const uiTag = mapTag(item.tag);

              return (
                <Tile key={item.id} className="feed-item">
                  <div className="feed-header">
                    <Information size={16} />
                    <h4>{item.title}</h4>
                    <Tag type={uiTag.type} size="sm">
                      {uiTag.label}
                    </Tag>
                  </div>

                  <p className="feed-message">{item.message}</p>

                  {renderAnnouncementLink(item)}

                  <div className="feed-timestamp">
                    <Time size={14} />
                    <span>{formatTimestamp(item.created_at)}</span>
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
