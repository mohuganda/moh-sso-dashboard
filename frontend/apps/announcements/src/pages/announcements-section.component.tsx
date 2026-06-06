import { Pin } from "@carbon/react/icons";
import { Stack, Tag, OverflowMenu, OverflowMenuItem } from "@carbon/react";
import type { Announcement } from "@/store/types/announcements.types";
import { formatRelativeTime, getStatusTagType, getTagType, truncateText } from "@/utils/utils";

export default function AnnouncementSection({
  title,
  items,
  selectedId,
  onSelect,
  onEdit,
  onPublish,
  onMoveToDraft,
  onArchive,
  onDelete,
  onTogglePin,
}: {
  title: string;
  items: Announcement[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onEdit: (item: Announcement) => void;
  onPublish: (item: Announcement) => void;
  onMoveToDraft: (item: Announcement) => void;
  onArchive: (item: Announcement) => void;
  onDelete: (item: Announcement) => void;
  onTogglePin: (item: Announcement) => void;
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
            <div
              key={item.id}
              style={{
                border: isActive ? "1px solid #0f62fe" : "1px solid #e0e0e0",
                background: isActive ? "#edf5ff" : "#ffffff",
                padding: "0.9rem",
              }}
            >
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr auto",
                  gap: "0.75rem",
                  alignItems: "start",
                }}
              >
                <button
                  type="button"
                  onClick={() => onSelect(item.id)}
                  style={{
                    textAlign: "left",
                    border: "none",
                    background: "transparent",
                    padding: 0,
                    cursor: "pointer",
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
                      {truncateText(item.summary || item.message, 110)}
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
                      <Tag type={getStatusTagType(item.status)}>{item.status}</Tag>
                      {item.is_pinned && <Tag type="warm-gray">Pinned</Tag>}
                    </div>
                  </Stack>
                </button>

                <OverflowMenu flipped>
                  <OverflowMenuItem itemText="Edit" onClick={() => onEdit(item)} />
                  {item.status !== "PUBLISHED" && (
                    <OverflowMenuItem itemText="Publish now" onClick={() => onPublish(item)} />
                  )}
                  {item.status !== "DRAFT" && (
                    <OverflowMenuItem
                      itemText="Move to draft"
                      onClick={() => onMoveToDraft(item)}
                    />
                  )}
                  {item.status !== "ARCHIVED" && (
                    <OverflowMenuItem itemText="Archive" onClick={() => onArchive(item)} />
                  )}
                  <OverflowMenuItem
                    itemText={item.is_pinned ? "Unpin" : "Pin"}
                    onClick={() => onTogglePin(item)}
                  />
                  <OverflowMenuItem isDelete itemText="Delete" onClick={() => onDelete(item)} />
                </OverflowMenu>
              </div>
            </div>
          );
        })}
      </Stack>
    </div>
  );
}
