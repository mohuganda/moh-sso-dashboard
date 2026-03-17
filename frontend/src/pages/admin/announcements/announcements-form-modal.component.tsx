import { Modal } from "@carbon/react";
import type {
  Announcement,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from "../../../store/types/announcements.types";
import { AnnouncementForm } from "../../../components/announcements/announcements-form.component";

interface AnnouncementFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  announcement?: Announcement | null;
  isSubmitting?: boolean;
  onRequestClose: () => void;
  onSubmit: (payload: CreateAnnouncementRequest | UpdateAnnouncementRequest) => Promise<void>;
}

export function AnnouncementFormModal({
  open,
  mode,
  announcement,
  isSubmitting = false,
  onRequestClose,
  onSubmit,
}: AnnouncementFormModalProps) {
  return (
    <Modal
      open={open}
      modalHeading={mode === "create" ? "Create announcement" : "Edit announcement"}
      passiveModal
      onRequestClose={onRequestClose}
      size="lg"
    >
      <div style={{ paddingTop: "0.5rem" }}>
        <AnnouncementForm
          mode={mode}
          initialValues={announcement}
          isSubmitting={isSubmitting}
          onSubmit={onSubmit}
          onCancel={onRequestClose}
        />
      </div>
    </Modal>
  );
}
