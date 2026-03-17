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
} from "@carbon/react";
import { Add, Save, Send } from "@carbon/react/icons";

import type {
  Announcement,
  AnnouncementAudienceType,
  AnnouncementLevel,
  AnnouncementStatus,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from "../../store/types/announcements.types";
import { useToast } from "../notifications/toast/useToast";

export interface AnnouncementFormValues {
  title: string;
  message: string;
  summary: string;
  level: AnnouncementLevel;
  tag: string;
  link_url: string;
  link_label: string;
  priority: number;
  is_pinned: boolean;
  status: AnnouncementStatus;
  audience_type: AnnouncementAudienceType;
  publish_at: string;
  expires_at: string;
}

interface AnnouncementFormProps {
  mode: "create" | "edit";
  initialValues?: Partial<AnnouncementFormValues> | Announcement | null;
  isSubmitting?: boolean;
  onSubmit: (
    payload: CreateAnnouncementRequest | UpdateAnnouncementRequest,
  ) => Promise<void> | void;
  onCancel?: () => void;
  submitLabel?: string;
}

const initialForm: AnnouncementFormValues = {
  title: "",
  message: "",
  summary: "",
  level: "INFO",
  tag: "",
  link_url: "",
  link_label: "",
  priority: 0,
  is_pinned: false,
  status: "DRAFT",
  audience_type: "ALL_USERS",
  publish_at: "",
  expires_at: "",
};

function toIsoString(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}

function normalizeInitialValues(
  values?: Partial<AnnouncementFormValues> | Announcement | null,
): AnnouncementFormValues {
  if (!values) return initialForm;

  return {
    title: values.title ?? "",
    message: values.message ?? "",
    summary: values.summary ?? "",
    level: (values.level as AnnouncementLevel) ?? "INFO",
    tag: values.tag ?? "",
    link_url: values.link_url ?? "",
    link_label: values.link_label ?? "",
    priority: typeof values.priority === "number" ? values.priority : 0,
    is_pinned: Boolean(values.is_pinned),
    status: (values.status as AnnouncementStatus) ?? "DRAFT",
    audience_type: (values.audience_type as AnnouncementAudienceType) ?? "ALL_USERS",
    publish_at: toIsoString(values.publish_at),
    expires_at: toIsoString(values.expires_at),
  };
}

export function AnnouncementForm({
  mode,
  initialValues,
  isSubmitting = false,
  onSubmit,
  onCancel,
  submitLabel,
}: AnnouncementFormProps) {
  const toast = useToast();
  const [form, setForm] = useState<AnnouncementFormValues>(normalizeInitialValues(initialValues));

  const isValid = useMemo(() => {
    return form.title.trim().length > 0 && form.message.trim().length > 0;
  }, [form]);

  const updateForm = <K extends keyof AnnouncementFormValues>(
    key: K,
    value: AnnouncementFormValues[K],
  ) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const resetForm = () => {
    setForm(normalizeInitialValues(initialValues));
  };

  const buildPayload = (): CreateAnnouncementRequest | UpdateAnnouncementRequest => {
    const payload: CreateAnnouncementRequest | UpdateAnnouncementRequest = {
      title: form.title.trim(),
      message: form.message.trim(),
      level: form.level,
      priority: Number.isNaN(Number(form.priority)) ? 0 : Number(form.priority),
      is_pinned: form.is_pinned,
      status: form.status,
      audience_type: form.audience_type,
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

    if (form.link_label.trim()) {
      payload.link_label = form.link_label.trim();
    }

    if (form.publish_at.trim()) {
      payload.publish_at = form.publish_at;
    }

    if (form.expires_at.trim()) {
      payload.expires_at = form.expires_at;
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
      await onSubmit(buildPayload());

      if (mode === "create") {
        toast.success("Announcement Created", "Announcement created successfully.");
        setForm(initialForm);
      } else {
        toast.success("Announcement Updated", "Announcement updated successfully.");
      }
    } catch (error: any) {
      const message =
        error?.data?.error ||
        error?.data?.message ||
        error?.message ||
        `Failed to ${mode === "create" ? "create" : "update"} announcement.`;

      toast.error(mode === "create" ? "Create Failed" : "Update Failed", message);
    }
  };

  return (
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
            <SelectItem value="ARCHIVED" text="ARCHIVED" />
          </Select>

          <Select
            id="announcement-audience-type"
            labelText="Audience"
            value={form.audience_type}
            onChange={(e) =>
              updateForm("audience_type", e.target.value as AnnouncementAudienceType)
            }
          >
            <SelectItem value="ALL_USERS" text="ALL_USERS" />
            <SelectItem value="ADMINS_ONLY" text="ADMINS_ONLY" />
            <SelectItem value="SPECIFIC_ROLES" text="SPECIFIC_ROLES" />
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

          <TextInput
            id="announcement-link-label"
            labelText="Link label"
            placeholder="Optional link label"
            value={form.link_label}
            onChange={(e) => updateForm("link_label", e.target.value)}
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

          <div>
            <DatePicker
              datePickerType="single"
              value={form.expires_at ? [form.expires_at] : []}
              onChange={(dates) => {
                const selectedDate = dates?.[0];
                updateForm(
                  "expires_at",
                  selectedDate instanceof Date ? selectedDate.toISOString() : "",
                );
              }}
            >
              <DatePickerInput
                id="announcement-expiry-date"
                labelText="Expiry date"
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
          <Button
            type="submit"
            renderIcon={mode === "create" ? Send : Save}
            disabled={!isValid || isSubmitting}
          >
            {isSubmitting
              ? mode === "create"
                ? "Saving..."
                : "Updating..."
              : (submitLabel ?? (mode === "create" ? "Create announcement" : "Save changes"))}
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

          {onCancel ? (
            <Button type="button" kind="ghost" onClick={onCancel} disabled={isSubmitting}>
              Cancel
            </Button>
          ) : null}
        </div>
      </Stack>
    </Form>
  );
}
