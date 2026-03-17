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
  InlineNotification,
} from "@carbon/react";
import { Notification, Pin, Time, CheckmarkOutline } from "@carbon/react/icons";

import { useListAnnouncementsAdminQuery } from "../../../store/api/announcement.api";
import type { Announcement, AnnouncementLevel } from "../../../store/types/announcements.types";

type AnnouncementWithReadState = Announcement & {
  read?: boolean;
};

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

function truncateText(text: string, limit = 110) {
  if (!text) return "";
  return text.length > limit ? `${text.slice(0, limit)}...` : text;
}

export function AnnouncementsPage() {
  const [announcements, setAnnouncements] = useState<AnnouncementWithReadState[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [search, setSearch] = useState("");

  const {
    data: fetchedAnnouncements = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useListAnnouncementsAdminQuery({
    limit: 20,
    offset: 0,
  });

  useEffect(() => {
    const nextAnnouncements = fetchedAnnouncements.map((item) => ({
      ...item,
      read: false,
    }));

    setAnnouncements(nextAnnouncements);
    setSelectedId((prev) => prev ?? nextAnnouncements[0]?.id ?? null);
  }, [fetchedAnnouncements]);

  const filteredAnnouncements = useMemo(() => {
    const q = search.trim().toLowerCase();

    if (!q) return announcements;

    return announcements.filter((item) => {
      return (
        item.title.toLowerCase().includes(q) ||
        item.message.toLowerCase().includes(q) ||
        (item.summary?.toLowerCase().includes(q) ?? false) ||
        (item.tag?.toLowerCase().includes(q) ?? false) ||
        item.level.toLowerCase().includes(q) ||
        item.status.toLowerCase().includes(q)
      );
    });
  }, [announcements, search]);

  const pinnedAnnouncements = useMemo(
    () => filteredAnnouncements.filter((item) => item.is_pinned),
    [filteredAnnouncements],
  );

  const unreadAnnouncements = useMemo(
    () => filteredAnnouncements.filter((item) => !item.read),
    [filteredAnnouncements],
  );

  const recentAnnouncements = useMemo(
    () => filteredAnnouncements.filter((item) => !item.is_pinned),
    [filteredAnnouncements],
  );

  const totalCount = announcements.length;
  const pinnedCount = announcements.filter((item) => item.is_pinned).length;
  const unreadCount = announcements.filter((item) => !item.read).length;

  const selectedAnnouncement = useMemo(
    () => filteredAnnouncements.find((item) => item.id === selectedId) ?? null,
    [filteredAnnouncements, selectedId],
  );

  useEffect(() => {
    if (filteredAnnouncements.length === 0) {
      setSelectedId(null);
      return;
    }

    const stillExists = filteredAnnouncements.some((item) => item.id === selectedId);
    if (!stillExists) {
      setSelectedId(filteredAnnouncements[0].id);
    }
  }, [filteredAnnouncements, selectedId]);

  const handleSelect = (id: string) => {
    setSelectedId(id);
    setAnnouncements((prev) =>
      prev.map((item) => (item.id === id ? { ...item, read: true } : item)),
    );
  };

  const handleMarkAllAsRead = () => {
    setAnnouncements((prev) => prev.map((item) => ({ ...item, read: true })));
  };

  if (isLoading) {
    return (
      <Tile style={{ minHeight: "420px" }}>
        <Stack gap={5}>
          <InlineLoading description="Loading admin announcements..." />
        </Stack>
      </Tile>
    );
  }

  if (isError) {
    return (
      <Tile style={{ minHeight: "420px", padding: "1rem" }}>
        <Stack gap={5}>
          <InlineNotification
            kind="error"
            lowContrast
            title="Failed to load announcements"
            subtitle={
              (error as any)?.data?.message ||
              (error as any)?.error ||
              "An unexpected error occurred while fetching announcements."
            }
          />
          <div>
            <Button kind="secondary" onClick={() => refetch()}>
              Retry
            </Button>
          </div>
        </Stack>
      </Tile>
    );
  }

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "400px 1fr",
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
                Admin Announcements
              </h3>
              <p
                style={{
                  margin: "0.35rem 0 0",
                  color: "#6f6f6f",
                  fontSize: "0.875rem",
                }}
              >
                Manage published notices, alerts, and internal updates
              </p>
            </div>

            {unreadCount > 0 && <Tag type="blue">{unreadCount} unread</Tag>}
          </div>

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3, minmax(0, 1fr))",
              gap: "0.75rem",
            }}
          >
            <SummaryCard label="Total" value={totalCount} />
            <SummaryCard label="Pinned" value={pinnedCount} />
            <SummaryCard label="Unread" value={unreadCount} />
          </div>

          <Search
            id="admin-announcements-search"
            labelText="Search announcements"
            placeholder="Search by title, message, tag, level..."
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
            <TabList aria-label="Admin announcement tabs" contained>
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
                  title="Pinned announcements"
                  items={pinnedAnnouncements}
                  selectedId={selectedId}
                  onSelect={handleSelect}
                />
              </TabPanel>

              <TabPanel style={{ paddingInline: 0 }}>
                <AnnouncementSection
                  title="Unread announcements"
                  items={unreadAnnouncements}
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

                  {selectedAnnouncement.is_pinned && <Tag type="warm-gray">Pinned</Tag>}

                  {!selectedAnnouncement.read && <Tag type="blue">Unread</Tag>}

                  {selectedAnnouncement.tag && (
                    <Tag type="warm-gray">{selectedAnnouncement.tag}</Tag>
                  )}

                  <Tag type="warm-gray">{selectedAnnouncement.status}</Tag>
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
                {formatRelativeTime(selectedAnnouncement.created_at)}
              </div>
            </div>

            {selectedAnnouncement.summary ? (
              <div
                style={{
                  fontSize: "0.9rem",
                  color: "#525252",
                  lineHeight: 1.5,
                  padding: "0.9rem 1rem",
                  background: "#f4f4f4",
                  borderLeft: "4px solid #0f62fe",
                }}
              >
                {selectedAnnouncement.summary}
              </div>
            ) : null}

            <div
              style={{
                fontSize: "0.95rem",
                lineHeight: 1.7,
                color: "#161616",
              }}
            >
              {selectedAnnouncement.message}
            </div>

            {selectedAnnouncement.link_url ? (
              <div>
                <a
                  href={selectedAnnouncement.link_url}
                  target="_blank"
                  rel="noreferrer"
                  style={{ color: "#0f62fe", textDecoration: "none" }}
                >
                  Open related link
                </a>
              </div>
            ) : null}
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
            No announcement selected
          </div>
        )}
      </Tile>
    </div>
  );
}

function SummaryCard({ label, value }: { label: string; value: number }) {
  return (
    <div
      style={{
        border: "1px solid #e0e0e0",
        padding: "0.85rem",
        background: "#ffffff",
      }}
    >
      <div
        style={{
          fontSize: "0.75rem",
          color: "#6f6f6f",
          marginBottom: "0.35rem",
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontSize: "1.25rem",
          fontWeight: 600,
          color: "#161616",
        }}
      >
        {value}
      </div>
    </div>
  );
}

function AnnouncementList({
  pinnedAnnouncements,
  recentAnnouncements,
  selectedId,
  onSelect,
}: {
  pinnedAnnouncements: AnnouncementWithReadState[];
  recentAnnouncements: AnnouncementWithReadState[];
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
  items: AnnouncementWithReadState[];
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
                width: "100%",
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
                    {item.is_pinned && <Pin size={14} />}
                    {formatRelativeTime(item.created_at)}
                  </div>
                </div>

                <div
                  style={{
                    fontSize: "0.875rem",
                    color: "#525252",
                    lineHeight: 1.4,
                  }}
                >
                  {truncateText(item.message, 110)}
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
                  {item.is_pinned && <Tag type="warm-gray">Pinned</Tag>}
                </div>
              </Stack>
            </button>
          );
        })}
      </Stack>
    </div>
  );
}
