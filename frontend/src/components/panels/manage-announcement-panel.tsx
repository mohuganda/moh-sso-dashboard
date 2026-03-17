import { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  DatePicker,
  DatePickerInput,
  Form,
  Select,
  SelectItem,
  Stack,
  TextArea,
  TextInput,
  Tile,
} from "@carbon/react";
import { Add, Send } from "@carbon/react/icons";
import type {
  AnnouncementLevel,
  AnnouncementStatus,
  CreateAnnouncementRequest,
} from "../../store/types/announcements.types";
import { useCreateAnnouncementMutation } from "../../store/api/announcement.api";
import { useToast } from "../notifications/toast/useToast";

interface AnnouncementForm {
  title: string;
  message: string;
  summary: string;
  level: AnnouncementLevel;
  tag: string;
  link_url: string;
  priority: number;
  is_pinned: boolean;
  status: AnnouncementStatus;
  publish_at: string;
}

const initialForm: AnnouncementForm = {
  title: "",
  message: "",
  summary: "",
  level: "INFO",
  tag: "",
  link_url: "",
  priority: 0,
  is_pinned: false,
  status: "DRAFT",
  publish_at: "",
};

export function ManageAnnouncementsPanel() {
  const toast = useToast();

  const [form, setForm] = useState<AnnouncementForm>(initialForm);

  const [createAnnouncement, { isLoading: isSubmitting }] = useCreateAnnouncementMutation();

  const isValid = useMemo(() => {
    return form.title.trim().length > 0 && form.message.trim().length > 0;
  }, [form]);

  const updateForm = <K extends keyof AnnouncementForm>(key: K, value: AnnouncementForm[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const resetForm = () => {
    setForm(initialForm);
  };

  const buildPayload = (): CreateAnnouncementRequest => {
    const payload: CreateAnnouncementRequest = {
      title: form.title.trim(),
      message: form.message.trim(),
      level: form.level,
      priority: Number.isNaN(Number(form.priority)) ? 0 : Number(form.priority),
      is_pinned: form.is_pinned,
      status: form.status,
    };

    if (form.summary.trim()) {
      payload.summary = form.summary.trim();
    }

    if (form.tag.trim()) {
      payload.tag = form.tag.trim();
    }

    if (form.link_url.trim()) {
      payload.link_url = form.link_url.trim();
    }

    if (form.publish_at.trim()) {
      payload.publish_at = form.publish_at;
    }

    return payload;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!isValid) {
      toast.error("Invalid", "Please provide both a title and message.");
      return;
    }

    try {
      const payload = buildPayload();

      await createAnnouncement(payload).unwrap();

      toast.success("Announcement Created", "Announcement created successfully.");
      resetForm();
    } catch (error: any) {
      const message =
        error?.data?.error ||
        error?.data?.message ||
        error?.message ||
        "Failed to create announcement. Please try again.";

      toast.error("Create Failed", message);
    }
  };

  return (
    <Tile style={{ padding: "1.5rem", maxWidth: "880px" }}>
      <Stack gap={6}>
        <div>
          <h3 style={{ margin: 0 }}>Create Announcement</h3>
          <p style={{ marginTop: "0.5rem", color: "#6f6f6f" }}>
            Publish announcements, alerts, and important updates for dashboard users.
          </p>
        </div>

        <Form onSubmit={handleSubmit}>
          <Stack gap={6}>
            <TextInput
              id="announcement-title"
              labelText="Title"
              placeholder="Enter announcement title"
              value={form.title}
              onChange={(e) => updateForm("title", e.target.value)}
              maxLength={255}
            />

            <TextInput
              id="announcement-summary"
              labelText="Summary"
              placeholder="Optional short summary"
              value={form.summary}
              onChange={(e) => updateForm("summary", e.target.value)}
              maxLength={500}
            />

            <TextArea
              id="announcement-message"
              labelText="Message"
              placeholder="Write the announcement message"
              rows={8}
              value={form.message}
              onChange={(e) => updateForm("message", e.target.value)}
              maxCount={2000}
              enableCounter
            />

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
                gap: "1rem",
                alignItems: "start",
              }}
            >
              <Select
                id="announcement-level"
                labelText="Level"
                value={form.level}
                onChange={(e) => updateForm("level", e.target.value as AnnouncementLevel)}
              >
                <SelectItem value="INFO" text="INFO" />
                <SelectItem value="SUCCESS" text="SUCCESS" />
                <SelectItem value="WARNING" text="WARNING" />
                <SelectItem value="CRITICAL" text="CRITICAL" />
              </Select>

              <Select
                id="announcement-status"
                labelText="Status"
                value={form.status}
                onChange={(e) => updateForm("status", e.target.value as AnnouncementStatus)}
              >
                <SelectItem value="DRAFT" text="DRAFT" />
                <SelectItem value="SCHEDULED" text="SCHEDULED" />
                <SelectItem value="PUBLISHED" text="PUBLISHED" />
              </Select>

              <TextInput
                id="announcement-tag"
                labelText="Tag"
                placeholder="Optional tag"
                value={form.tag}
                onChange={(e) => updateForm("tag", e.target.value)}
              />

              <TextInput
                id="announcement-priority"
                labelText="Priority"
                type="number"
                min={0}
                value={String(form.priority)}
                onChange={(e) => updateForm("priority", Number(e.target.value || 0))}
              />

              <TextInput
                id="announcement-link-url"
                labelText="Link URL"
                placeholder="https://example.com"
                value={form.link_url}
                onChange={(e) => updateForm("link_url", e.target.value)}
              />

              <div>
                <DatePicker
                  datePickerType="single"
                  value={form.publish_at ? [form.publish_at] : []}
                  onChange={(dates) => {
                    const selectedDate = dates?.[0];
                    updateForm(
                      "publish_at",
                      selectedDate instanceof Date ? selectedDate.toISOString() : "",
                    );
                  }}
                >
                  <DatePickerInput
                    id="announcement-publish-date"
                    labelText="Publish date"
                    placeholder="yyyy-mm-dd"
                  />
                </DatePicker>
              </div>
            </div>

            <Checkbox
              id="announcement-pinned"
              labelText="Pin this announcement"
              checked={form.is_pinned}
              onChange={(_, { checked }) => updateForm("is_pinned", Boolean(checked))}
            />

            <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
              <Button type="submit" renderIcon={Send} disabled={!isValid || isSubmitting}>
                {isSubmitting ? "Saving..." : "Create announcement"}
              </Button>

              <Button
                type="button"
                kind="secondary"
                renderIcon={Add}
                onClick={resetForm}
                disabled={isSubmitting}
              >
                Reset form
              </Button>
            </div>
          </Stack>
        </Form>
      </Stack>
    </Tile>
  );
}
