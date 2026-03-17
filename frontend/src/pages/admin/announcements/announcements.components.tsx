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
  Dropdown,
} from "@carbon/react";
import { Notification, Time, Add, Edit, Launch } from "@carbon/react/icons";

import {
  useListAnnouncementsAdminQuery,
  usePublishAnnouncementMutation,
  useArchiveAnnouncementMutation,
  useDeleteAnnouncementMutation,
  useUpdateAnnouncementMutation,
  useCreateAnnouncementMutation,
} from "../../../store/api/announcement.api";
import type {
  Announcement,
  AnnouncementLevel,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from "../../../store/types/announcements.types";
import AnnouncementSection from "./announcements-section.component";
import MetadataItem from "./announcements-metadata-item.components";
import SummaryCard from "./announcements-summary-card.component";
import {
  isAnnouncementActive,
  getTagType,
  getStatusTagType,
  formatRelativeTime,
  formatDateTime,
} from "../../../utils/utils";
import { AnnouncementFormModal } from "./announcements-form-modal.component";

type AudienceFilter = "ALL" | "ALL_USERS" | "ADMINS_ONLY" | "SPECIFIC_ROLES";
type LevelFilter = "ALL" | AnnouncementLevel;
type StatusTab = "ALL" | "PUBLISHED" | "DRAFT" | "SCHEDULED" | "ARCHIVED";

export function AnnouncementsPage() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [statusTab, setStatusTab] = useState<StatusTab>("ALL");
  const [levelFilter, setLevelFilter] = useState<LevelFilter>("ALL");
  const [audienceFilter, setAudienceFilter] = useState<AudienceFilter>("ALL");
  const [pinnedOnly, setPinnedOnly] = useState(false);

  const {
    data: announcements = [],
    isLoading,
    isError,
    error,
    refetch,
  } = useListAnnouncementsAdminQuery({
    limit: 50,
    offset: 0,
  });

  const [publishAnnouncement, publishState] = usePublishAnnouncementMutation();
  const [archiveAnnouncement, archiveState] = useArchiveAnnouncementMutation();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingAnnouncement, setEditingAnnouncement] = useState<Announcement | null>(null);

  const [createAnnouncement, createState] = useCreateAnnouncementMutation();
  const [updateAnnouncement, updateState] = useUpdateAnnouncementMutation();
  const [deleteAnnouncement, deleteState] = useDeleteAnnouncementMutation();

  useEffect(() => {
    if (!selectedId && announcements.length > 0) {
      setSelectedId(announcements[0].id);
    }
  }, [announcements, selectedId]);

  const filteredAnnouncements = useMemo(() => {
    const q = search.trim().toLowerCase();

    return announcements.filter((item) => {
      const matchesSearch =
        !q ||
        item.title.toLowerCase().includes(q) ||
        item.message.toLowerCase().includes(q) ||
        (item.summary?.toLowerCase().includes(q) ?? false) ||
        (item.tag?.toLowerCase().includes(q) ?? false) ||
        item.level.toLowerCase().includes(q) ||
        item.status.toLowerCase().includes(q);

      const matchesStatus = statusTab === "ALL" ? true : item.status === statusTab;
      const matchesLevel = levelFilter === "ALL" ? true : item.level === levelFilter;
      const matchesAudience =
        audienceFilter === "ALL" ? true : item.audience_type === audienceFilter;
      const matchesPinned = pinnedOnly ? item.is_pinned : true;

      return matchesSearch && matchesStatus && matchesLevel && matchesAudience && matchesPinned;
    });
  }, [announcements, search, statusTab, levelFilter, audienceFilter, pinnedOnly]);

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

  const selectedAnnouncement = useMemo(
    () => filteredAnnouncements.find((item) => item.id === selectedId) ?? null,
    [filteredAnnouncements, selectedId],
  );

  const counts = useMemo(() => {
    return {
      total: announcements.length,
      published: announcements.filter((item) => item.status === "PUBLISHED").length,
      drafts: announcements.filter((item) => item.status === "DRAFT").length,
      scheduled: announcements.filter((item) => item.status === "SCHEDULED").length,
      archived: announcements.filter((item) => item.status === "ARCHIVED").length,
      pinned: announcements.filter((item) => item.is_pinned).length,
      active: announcements.filter((item) => isAnnouncementActive(item)).length,
    };
  }, [announcements]);

  const handleSelect = (id: string) => {
    setSelectedId(id);
  };

  const handleCreate = () => {
    setIsCreateModalOpen(true);
  };

  const handleEdit = (item: Announcement) => {
    setEditingAnnouncement(item);
  };

  const handleCreateSubmit = async (
    payload: CreateAnnouncementRequest | UpdateAnnouncementRequest,
  ) => {
    await createAnnouncement(payload as CreateAnnouncementRequest).unwrap();
    setIsCreateModalOpen(false);
  };

  const handleEditSubmit = async (
    payload: CreateAnnouncementRequest | UpdateAnnouncementRequest,
  ) => {
    if (!editingAnnouncement) return;

    await updateAnnouncement({
      id: editingAnnouncement.id,
      body: payload as UpdateAnnouncementRequest,
    }).unwrap();

    setEditingAnnouncement(null);
  };

  const handlePublishNow = async (item: Announcement) => {
    try {
      await publishAnnouncement(item.id).unwrap();
    } catch (err) {
      console.error("Failed to publish announcement", err);
    }
  };

  const handleMoveToDraft = async (item: Announcement) => {
    try {
      await updateAnnouncement({
        id: item.id,
        status: "DRAFT",
      }).unwrap();
    } catch (err) {
      console.error("Failed to move announcement to draft", err);
    }
  };

  const handleArchive = async (item: Announcement) => {
    try {
      await archiveAnnouncement(item.id).unwrap();
    } catch (err) {
      console.error("Failed to archive announcement", err);
    }
  };

  const handleDelete = async (item: Announcement) => {
    const confirmed = window.confirm(
      `Delete "${item.title}"? This action should only be used when you're sure.`,
    );
    if (!confirmed) return;

    try {
      await deleteAnnouncement(item.id).unwrap();
      if (selectedId === item.id) {
        setSelectedId(null);
      }
    } catch (err) {
      console.error("Failed to delete announcement", err);
    }
  };

  const handleTogglePin = async (item: Announcement) => {
    try {
      await updateAnnouncement({
        id: item.id,
        is_pinned: !item.is_pinned,
      }).unwrap();
    } catch (err) {
      console.error("Failed to update pin state", err);
    }
  };

  const isMutating =
    publishState.isLoading ||
    archiveState.isLoading ||
    deleteState.isLoading ||
    updateState.isLoading;

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
    <>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "820px 1fr",
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
                  Manage drafts, scheduled notices, published alerts, and archived updates
                </p>
              </div>

              <Button renderIcon={Add} onClick={handleCreate}>
                New announcement
              </Button>
            </div>

            {isMutating && (
              <InlineNotification
                kind="info"
                lowContrast
                title="Updating announcements"
                subtitle="Your changes are being processed."
              />
            )}

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
                gap: "0.75rem",
              }}
            >
              <SummaryCard label="Total" value={counts.total} />
              <SummaryCard label="Published" value={counts.published} />
              <SummaryCard label="Drafts" value={counts.drafts} />
              <SummaryCard label="Scheduled" value={counts.scheduled} />
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(3, minmax(0, 1fr))",
                gap: "0.75rem",
              }}
            >
              <SummaryCard label="Archived" value={counts.archived} />
              <SummaryCard label="Pinned" value={counts.pinned} />
              <SummaryCard label="Active now" value={counts.active} />
            </div>

            <Search
              id="admin-announcements-search"
              labelText="Search announcements"
              placeholder="Search by title, message, tag, level, status..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              size="lg"
            />

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(3, minmax(0, 1fr)) auto",
                gap: "0.75rem",
                alignItems: "end",
              }}
            >
              <Dropdown
                id="announcement-level-filter"
                titleText="Level"
                label="All levels"
                items={["ALL", "INFO", "SUCCESS", "WARNING", "CRITICAL"]}
                selectedItem={levelFilter}
                onChange={({ selectedItem }) =>
                  setLevelFilter((selectedItem as LevelFilter) ?? "ALL")
                }
              />

              <Dropdown
                id="announcement-audience-filter"
                titleText="Audience"
                label="All audiences"
                items={["ALL", "ALL_USERS", "ADMINS_ONLY", "SPECIFIC_ROLES"]}
                selectedItem={audienceFilter}
                onChange={({ selectedItem }) =>
                  setAudienceFilter((selectedItem as AudienceFilter) ?? "ALL")
                }
              />

              <Dropdown
                id="announcement-pin-filter"
                titleText="Pin state"
                label="All items"
                items={["ALL", "PINNED_ONLY"]}
                selectedItem={pinnedOnly ? "PINNED_ONLY" : "ALL"}
                onChange={({ selectedItem }) => setPinnedOnly(selectedItem === "PINNED_ONLY")}
              />

              <Button
                kind="ghost"
                onClick={() => {
                  setSearch("");
                  setStatusTab("ALL");
                  setLevelFilter("ALL");
                  setAudienceFilter("ALL");
                  setPinnedOnly(false);
                }}
              >
                Reset
              </Button>
            </div>

            <Tabs
              selectedIndex={
                {
                  ALL: 0,
                  PUBLISHED: 1,
                  DRAFT: 2,
                  SCHEDULED: 3,
                  ARCHIVED: 4,
                }[statusTab]
              }
              onChange={({ selectedIndex }) => {
                const next =
                  (["ALL", "PUBLISHED", "DRAFT", "SCHEDULED", "ARCHIVED"][
                    selectedIndex
                  ] as StatusTab) ?? "ALL";
                setStatusTab(next);
              }}
            >
              <TabList aria-label="Admin announcement tabs" contained>
                <Tab>All</Tab>
                <Tab>Published</Tab>
                <Tab>Drafts</Tab>
                <Tab>Scheduled</Tab>
                <Tab>Archived</Tab>
              </TabList>

              <TabPanels>
                <TabPanel style={{ paddingInline: 0 }}>
                  <AnnouncementSection
                    title="All announcements"
                    items={filteredAnnouncements}
                    selectedId={selectedId}
                    onSelect={handleSelect}
                    onEdit={handleEdit}
                    onPublish={handlePublishNow}
                    onMoveToDraft={handleMoveToDraft}
                    onArchive={handleArchive}
                    onDelete={handleDelete}
                    onTogglePin={handleTogglePin}
                  />
                </TabPanel>

                <TabPanel style={{ paddingInline: 0 }}>
                  <AnnouncementSection
                    title="Published announcements"
                    items={filteredAnnouncements}
                    selectedId={selectedId}
                    onSelect={handleSelect}
                    onEdit={handleEdit}
                    onPublish={handlePublishNow}
                    onMoveToDraft={handleMoveToDraft}
                    onArchive={handleArchive}
                    onDelete={handleDelete}
                    onTogglePin={handleTogglePin}
                  />
                </TabPanel>

                <TabPanel style={{ paddingInline: 0 }}>
                  <AnnouncementSection
                    title="Draft announcements"
                    items={filteredAnnouncements}
                    selectedId={selectedId}
                    onSelect={handleSelect}
                    onEdit={handleEdit}
                    onPublish={handlePublishNow}
                    onMoveToDraft={handleMoveToDraft}
                    onArchive={handleArchive}
                    onDelete={handleDelete}
                    onTogglePin={handleTogglePin}
                  />
                </TabPanel>

                <TabPanel style={{ paddingInline: 0 }}>
                  <AnnouncementSection
                    title="Scheduled announcements"
                    items={filteredAnnouncements}
                    selectedId={selectedId}
                    onSelect={handleSelect}
                    onEdit={handleEdit}
                    onPublish={handlePublishNow}
                    onMoveToDraft={handleMoveToDraft}
                    onArchive={handleArchive}
                    onDelete={handleDelete}
                    onTogglePin={handleTogglePin}
                  />
                </TabPanel>

                <TabPanel style={{ paddingInline: 0 }}>
                  <AnnouncementSection
                    title="Archived announcements"
                    items={filteredAnnouncements}
                    selectedId={selectedId}
                    onSelect={handleSelect}
                    onEdit={handleEdit}
                    onPublish={handlePublishNow}
                    onMoveToDraft={handleMoveToDraft}
                    onArchive={handleArchive}
                    onDelete={handleDelete}
                    onTogglePin={handleTogglePin}
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

                    <Tag type={getStatusTagType(selectedAnnouncement.status)}>
                      {selectedAnnouncement.status}
                    </Tag>

                    {selectedAnnouncement.is_pinned && <Tag type="warm-gray">Pinned</Tag>}

                    {selectedAnnouncement.tag && (
                      <Tag type="warm-gray">{selectedAnnouncement.tag}</Tag>
                    )}

                    <Tag type="blue">{selectedAnnouncement.audience_type}</Tag>

                    {isAnnouncementActive(selectedAnnouncement) && (
                      <Tag type="green">Active now</Tag>
                    )}
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

              <div
                style={{
                  display: "flex",
                  gap: "0.75rem",
                  flexWrap: "wrap",
                }}
              >
                <Button
                  size="sm"
                  renderIcon={Edit}
                  onClick={() => handleEdit(selectedAnnouncement)}
                >
                  Edit
                </Button>

                {selectedAnnouncement.status !== "PUBLISHED" && (
                  <Button
                    size="sm"
                    kind="secondary"
                    onClick={() => handlePublishNow(selectedAnnouncement)}
                  >
                    Publish now
                  </Button>
                )}

                {selectedAnnouncement.status !== "DRAFT" && (
                  <Button
                    size="sm"
                    kind="ghost"
                    onClick={() => handleMoveToDraft(selectedAnnouncement)}
                  >
                    Move to draft
                  </Button>
                )}

                <Button
                  size="sm"
                  kind="ghost"
                  onClick={() => handleTogglePin(selectedAnnouncement)}
                >
                  {selectedAnnouncement.is_pinned ? "Unpin" : "Pin"}
                </Button>

                {selectedAnnouncement.status !== "ARCHIVED" && (
                  <Button
                    size="sm"
                    kind="ghost"
                    onClick={() => handleArchive(selectedAnnouncement)}
                  >
                    Archive
                  </Button>
                )}

                <Button
                  size="sm"
                  kind="danger--tertiary"
                  onClick={() => handleDelete(selectedAnnouncement)}
                >
                  Delete
                </Button>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
                  gap: "1rem",
                  padding: "1rem",
                  background: "#f4f4f4",
                }}
              >
                <MetadataItem label="Audience" value={selectedAnnouncement.audience_type} />
                <MetadataItem label="Level" value={selectedAnnouncement.level} />
                <MetadataItem label="Status" value={selectedAnnouncement.status} />
                <MetadataItem
                  label="Pinned"
                  value={selectedAnnouncement.is_pinned ? "Yes" : "No"}
                />
                <MetadataItem
                  label="Publish at"
                  value={formatDateTime(selectedAnnouncement.publish_at)}
                />
                <MetadataItem
                  label="Expires at"
                  value={formatDateTime(selectedAnnouncement.expires_at)}
                />
                <MetadataItem
                  label="Created at"
                  value={formatDateTime(selectedAnnouncement.created_at)}
                />
                <MetadataItem
                  label="Updated at"
                  value={formatDateTime(selectedAnnouncement.updated_at)}
                />
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
                    style={{
                      color: "#0f62fe",
                      textDecoration: "none",
                      display: "inline-flex",
                      alignItems: "center",
                      gap: "0.35rem",
                    }}
                  >
                    Open related link
                    <Launch size={16} />
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

      <AnnouncementFormModal
        open={isCreateModalOpen}
        mode={"create"}
        isSubmitting={createState.isLoading}
        onRequestClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateSubmit}
      />

      <AnnouncementFormModal
        open={Boolean(editingAnnouncement)}
        mode={"create"}
        announcement={editingAnnouncement}
        isSubmitting={updateState.isLoading}
        onRequestClose={() => setEditingAnnouncement(null)}
        onSubmit={handleEditSubmit}
      />
    </>
  );
}
