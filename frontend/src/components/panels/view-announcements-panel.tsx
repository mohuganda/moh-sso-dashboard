import { useEffect, useMemo, useState } from "react";
import {
  Button,
  Search,
  Tag,
  Tile,
  InlineLoading,
  Stack,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
} from "@carbon/react";
import { Notification, Pin, Time, CheckmarkOutline } from "@carbon/react/icons";

type AnnouncementLevel = "info" | "warning" | "success" | "critical";

interface Announcement {
  id: string;
  title: string;
  message: string;
  level: AnnouncementLevel;
  createdAt: string;
  pinned?: boolean;
  read?: boolean;
}

const MOCK_ANNOUNCEMENTS: Announcement[] = [
  {
    id: "1",
    title: "Scheduled maintenance",
    message:
      "The SSO platform will undergo scheduled maintenance tonight from 11:00 PM to 1:00 AM.",
    level: "warning",
    createdAt: "2026-03-12T08:30:00Z",
    pinned: true,
    read: false,
  },
  {
    id: "2",
    title: "New client onboarding",
    message:
      "A new ministry application has been onboarded into the dashboard and is now available to authorized users.",
    level: "success",
    createdAt: "2026-03-11T15:00:00Z",
    pinned: false,
    read: false,
  },
  {
    id: "3",
    title: "Policy update",
    message: "Password reset and session management policies have been updated for all users.",
    level: "info",
    createdAt: "2026-03-10T10:15:00Z",
    pinned: false,
    read: true,
  },
  {
    id: "4",
    title: "Security alert",
    message:
      "Multiple failed login attempts were detected across selected accounts. Please review audit logs.",
    level: "critical",
    createdAt: "2026-03-09T07:45:00Z",
    pinned: true,
    read: true,
  },
];

function getTagType(level: AnnouncementLevel): "blue" | "green" | "red" | "purple" | "warm-gray" {
  switch (level) {
    case "success":
      return "green";
    case "warning":
      return "purple";
    case "critical":
      return "red";
    case "info":
    default:
      return "blue";
  }
}

function formatRelativeTime(dateString: string) {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();

  const minutes = Math.floor(diffMs / (1000 * 60));
  const hours = Math.floor(diffMs / (1000 * 60 * 60));
  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (minutes < 1) return "Just now";
  if (minutes < 60) return `${minutes} min ago`;
  if (hours < 24) return `${hours} hr ago`;
  if (days === 1) return "Yesterday";
  return `${days} days ago`;
}

export function AnnouncementsPanel() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setAnnouncements(MOCK_ANNOUNCEMENTS);
      setSelectedId(MOCK_ANNOUNCEMENTS[0]?.id ?? null);
      setLoading(false);
    }, 600);

    return () => clearTimeout(timer);
  }, []);

  const unreadCount = useMemo(
    () => announcements.filter((item) => !item.read).length,
    [announcements],
  );

  const filteredAnnouncements = useMemo(() => {
    const q = search.trim().toLowerCase();

    if (!q) return announcements;

    return announcements.filter(
      (item) => item.title.toLowerCase().includes(q) || item.message.toLowerCase().includes(q),
    );
  }, [announcements, search]);

  const pinnedAnnouncements = useMemo(
    () => filteredAnnouncements.filter((item) => item.pinned),
    [filteredAnnouncements],
  );

  const recentAnnouncements = useMemo(
    () => filteredAnnouncements.filter((item) => !item.pinned),
    [filteredAnnouncements],
  );

  const selectedAnnouncement = useMemo(
    () => announcements.find((item) => item.id === selectedId) ?? null,
    [announcements, selectedId],
  );

  const handleSelect = (id: string) => {
    setSelectedId(id);
    setAnnouncements((prev) =>
      prev.map((item) => (item.id === id ? { ...item, read: true } : item)),
    );
  };

  const handleMarkAllAsRead = () => {
    setAnnouncements((prev) => prev.map((item) => ({ ...item, read: true })));
  };

  if (loading) {
    return (
      <Tile style={{ minHeight: "420px" }}>
        <Stack gap={5}>
          <InlineLoading description="Loading announcements..." />
        </Stack>
      </Tile>
    );
  }

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "380px 1fr",
        gap: "1rem",
        alignItems: "start",
      }}
    >
      <Tile style={{ padding: "1rem" }}>
        <Stack gap={5}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "flex-start",
              gap: "1rem",
            }}
          >
            <div>
              <h3
                style={{
                  margin: 0,
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                }}
              >
                <Notification size={20} />
                Announcements
              </h3>
              <p
                style={{
                  margin: "0.35rem 0 0",
                  color: "#6f6f6f",
                  fontSize: "0.875rem",
                }}
              >
                Latest updates, alerts, and notices
              </p>
            </div>

            {unreadCount > 0 && <Tag type="blue">{unreadCount} unread</Tag>}
          </div>

          <Search
            id="announcements-search"
            labelText="Search announcements"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            size="lg"
          />

          <Button
            kind="ghost"
            size="sm"
            renderIcon={CheckmarkOutline}
            onClick={handleMarkAllAsRead}
            disabled={unreadCount === 0}
          >
            Mark all as read
          </Button>

          <Tabs>
            <TabList aria-label="Announcement tabs" contained>
              <Tab>All</Tab>
              <Tab>Pinned</Tab>
              <Tab>Unread</Tab>
            </TabList>

            <TabPanels>
              <TabPanel style={{ paddingInline: 0 }}>
                <AnnouncementList
                  pinnedAnnouncements={pinnedAnnouncements}
                  recentAnnouncements={recentAnnouncements}
                  selectedId={selectedId}
                  onSelect={handleSelect}
                />
              </TabPanel>

              <TabPanel style={{ paddingInline: 0 }}>
                <AnnouncementSection
                  title="Pinned"
                  items={pinnedAnnouncements}
                  selectedId={selectedId}
                  onSelect={handleSelect}
                />
              </TabPanel>

              <TabPanel style={{ paddingInline: 0 }}>
                <AnnouncementSection
                  title="Unread"
                  items={filteredAnnouncements.filter((item) => !item.read)}
                  selectedId={selectedId}
                  onSelect={handleSelect}
                />
              </TabPanel>
            </TabPanels>
          </Tabs>
        </Stack>
      </Tile>

      <Tile style={{ padding: "1.25rem", minHeight: "420px" }}>
        {selectedAnnouncement ? (
          <Stack gap={6}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                gap: "1rem",
                alignItems: "flex-start",
              }}
            >
              <div>
                <h3 style={{ margin: 0 }}>{selectedAnnouncement.title}</h3>
                <div
                  style={{
                    display: "flex",
                    gap: "0.5rem",
                    alignItems: "center",
                    marginTop: "0.75rem",
                    flexWrap: "wrap",
                  }}
                >
                  <Tag type={getTagType(selectedAnnouncement.level)}>
                    {selectedAnnouncement.level}
                  </Tag>

                  {selectedAnnouncement.pinned && <Tag type="warm-gray">Pinned</Tag>}

                  {!selectedAnnouncement.read && <Tag type="blue">Unread</Tag>}
                </div>
              </div>

              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.35rem",
                  color: "#6f6f6f",
                  fontSize: "0.875rem",
                  whiteSpace: "nowrap",
                }}
              >
                <Time size={16} />
                {formatRelativeTime(selectedAnnouncement.createdAt)}
              </div>
            </div>

            <div
              style={{
                fontSize: "0.95rem",
                lineHeight: 1.6,
                color: "#161616",
              }}
            >
              {selectedAnnouncement.message}
            </div>
          </Stack>
        ) : (
          <div
            style={{
              minHeight: "100%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "#6f6f6f",
            }}
          >
            Select an announcement to view details
          </div>
        )}
      </Tile>
    </div>
  );
}

function AnnouncementList({
  pinnedAnnouncements,
  recentAnnouncements,
  selectedId,
  onSelect,
}: {
  pinnedAnnouncements: Announcement[];
  recentAnnouncements: Announcement[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  if (pinnedAnnouncements.length === 0 && recentAnnouncements.length === 0) {
    return <div style={{ padding: "1rem 0", color: "#6f6f6f" }}>No announcements found.</div>;
  }

  return (
    <Stack gap={6}>
      {pinnedAnnouncements.length > 0 && (
        <AnnouncementSection
          title="Pinned"
          items={pinnedAnnouncements}
          selectedId={selectedId}
          onSelect={onSelect}
        />
      )}

      {recentAnnouncements.length > 0 && (
        <AnnouncementSection
          title="Recent"
          items={recentAnnouncements}
          selectedId={selectedId}
          onSelect={onSelect}
        />
      )}
    </Stack>
  );
}

function AnnouncementSection({
  title,
  items,
  selectedId,
  onSelect,
}: {
  title: string;
  items: Announcement[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  if (items.length === 0) {
    return (
      <div>
        <h5 style={{ marginBottom: "0.75rem" }}>{title}</h5>
        <div style={{ color: "#6f6f6f" }}>No items available.</div>
      </div>
    );
  }

  return (
    <div>
      <h5 style={{ marginBottom: "0.75rem" }}>{title}</h5>

      <Stack gap={3}>
        {items.map((item) => {
          const isActive = selectedId === item.id;

          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onSelect(item.id)}
              style={{
                textAlign: "left",
                border: isActive ? "1px solid #0f62fe" : "1px solid #e0e0e0",
                background: isActive ? "#edf5ff" : "#ffffff",
                padding: "0.9rem",
                cursor: "pointer",
                borderRadius: 0,
              }}
            >
              <Stack gap={3}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    gap: "0.75rem",
                    alignItems: "flex-start",
                  }}
                >
                  <div style={{ fontWeight: 600, color: "#161616" }}>{item.title}</div>

                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "0.35rem",
                      color: "#6f6f6f",
                      fontSize: "0.75rem",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {item.pinned && <Pin size={14} />}
                    {formatRelativeTime(item.createdAt)}
                  </div>
                </div>

                <div
                  style={{
                    fontSize: "0.875rem",
                    color: "#525252",
                    lineHeight: 1.4,
                  }}
                >
                  {item.message.length > 110 ? `${item.message.slice(0, 110)}...` : item.message}
                </div>

                <div
                  style={{
                    display: "flex",
                    gap: "0.5rem",
                    flexWrap: "wrap",
                    alignItems: "center",
                  }}
                >
                  <Tag type={getTagType(item.level)}>{item.level}</Tag>
                  {!item.read && <Tag type="blue">Unread</Tag>}
                  {item.pinned && <Tag type="warm-gray">Pinned</Tag>}
                </div>
              </Stack>
            </button>
          );
        })}
      </Stack>
    </div>
  );
}
