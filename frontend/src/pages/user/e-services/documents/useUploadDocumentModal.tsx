import { useModal } from "../../../../components/modal/modal.context";
import { UploadDocumentModal } from "./UploadDocumentModal";

type UploadDocumentModalArgs = {};

export function useUploadDocumentModal() {
  const { openModal, closeModal } = useModal();

  function openUploadDocumentModal(UploadDocumentModalArgs) {
    openModal({
      title: "Upload Document",
      size: "sm",
      content: <UploadDocumentModal onClose={closeModal} />,
      onClose: closeModal,
    });
  }

  return { openUploadDocumentModal };
}
