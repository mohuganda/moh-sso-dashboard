import { useModal } from "@/shared/components/modal/modal.context";

import { ResetPasswordModal } from "./ResetPasswordModal";

type OpenResetPasswordModalArgs = {
  username: string;
  email?: string | null;
  onConfirm: () => Promise<void> | void;
};

export function useResetPasswordModal() {
  const { openModal, closeModal } = useModal();

  function openResetPasswordModal({ username, email, onConfirm }: OpenResetPasswordModalArgs) {
    openModal({
      title: "Reset password",
      size: "sm",
      content: (
        <ResetPasswordModal
          username={username}
          email={email}
          onConfirm={async () => {
            await onConfirm();
            closeModal();
          }}
          onCancel={closeModal}
        />
      ),
      onClose: closeModal,
    });
  }

  return { openResetPasswordModal };
}
