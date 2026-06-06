import { useModal } from "@/shared/components/modal/modal.context";
import { UploadDocumentModal } from "./UploadDocumentModal";

export function useUploadDocumentModal() {
  const { openModal, closeModal } = useModal();

  function openUploadDocumentModal() {
    openModal({
      title: "Upload Document",
      size: "sm",
      content: <UploadDocumentModal onClose={closeModal} />,
      onClose: closeModal,
    });
  }

  return { openUploadDocumentModal };
}
