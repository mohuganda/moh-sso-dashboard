import { Information, Time, Add, ChevronRight } from "@carbon/react/icons";
import { Tile, Link, Tag, SkeletonText } from "@carbon/react";
import { useNavigate } from "react-router-dom";

import { EmptyState } from "../../../components/emptystate/EmptyState";
import { ErrorState } from "../../../components/errorstate/ErrorState";
import "./news-feed.css";

/* --------------------------------
 * Types
 * -------------------------------- */
interface FeedItem {
  id: string;
  title: string;
  message: string;
  tag: {
    label: string;
    type: "blue" | "red" | "green" | "gray";
  };
  timestamp: string;
}

/* --------------------------------
 * Mock data
 * -------------------------------- */
const FEED_ITEMS: FeedItem[] = [
  {
    id: "1",
    title: "System Updates",
    message:
      "We have upgraded our authentication engine to support faster logins. You may now manage active sessions directly from your profile settings.",
    tag: {
      label: "New",
      type: "blue",
    },
    timestamp: "Posted today",
  },
  {
    id: "2",
    title: "Maintenance Notice",
    message:
      "To ensure database stability, the MOH Portal will undergo routine optimization this Saturday. Access to the Reporting module may be intermittent between 10:00 PM and 12:00 AM.",
    tag: {
      label: "Scheduled",
      type: "red",
    },
    timestamp: "Scheduled",
  },
  {
    id: "3",
    title: "Upcoming Events",
    message:
      "Join us for the National Digital Health Summit next month. We will be discussing the future of EMR integration across regional referral hospitals.",
    tag: {
      label: "Event",
      type: "blue",
    },
    timestamp: "Posted today",
  },
  {
    id: "4",
    title: "MOH Activity Listing",
    message:
      "The Ministry of Health will conduct a public sickle cell screening for children aged 10-15 starting next Monday at all district health centers.",
    tag: {
      label: "Action",
      type: "red",
    },
    timestamp: "Scheduled",
  },
  {
    id: "5",
    title: "Daily Feeds",
    message:
      "Did you know? Consistent use of the iIHMS system has reduced patient wait times by 15% this quarter. Check the Insights tab for more performance data.",
    tag: {
      label: "Info",
      type: "blue",
    },
    timestamp: "Posted today",
  },
];

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

export default function NewsFeedPage() {
  const navigate = useNavigate();

  const loading = false;
  const error: string | null = null;
  const isEmpty = FEED_ITEMS.length === 0;

  return (
    <div className="page-container">
      {/* ------------------------------
       * Header
       * ------------------------------ */}
      <header className="page-header">
        <h3 className="page-title">News & Updates</h3>
        <p className="page-subtitle">Latest system updates, announcements, and notices.</p>
      </header>

      {/* ------------------------------
       * Main content
       * ------------------------------ */}
      <section className="page-content">
        {/* ---------------- Feed ---------------- */}
        <main className="news-feed">
          {loading && <SkeletonText paragraph lineCount={4} />}

          {error && (
            <ErrorState
              title="Failed to load announcements"
              description={error}
              primaryAction={{
                label: "Retry",
                onClick: () => {
                  window.location.reload();
                },
              }}
              secondaryAction={{
                label: "Contact support",
                onClick: () => navigate("/support"),
              }}
            />
          )}

          {!error && !loading && isEmpty && (
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

          {!error &&
            !loading &&
            !isEmpty &&
            FEED_ITEMS.map((item) => (
              <Tile key={item.id} className="feed-item">
                <div className="feed-header">
                  <Information size={16} />
                  <h4>{item.title}</h4>
                  <Tag type={item.tag.type} size="sm">
                    {item.tag.label}
                  </Tag>
                </div>

                <p className="feed-message">{item.message}</p>

                <div className="feed-timestamp">
                  <Time size={14} />
                  <span>{item.timestamp}</span>
                </div>
              </Tile>
            ))}
        </main>

        {/* ---------------- Sidebar ---------------- */}
        <aside className="news-sidebar">
          <Tile>
            <h4>Case Reporting</h4>

            <ul className="case-reporting-list">
              {CASE_REPORTING.map((item) => (
                <li key={item.href}>
                  <Link
                    onClick={() => window.open(item.href, "_blank", "noopener,noreferrer")}
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
