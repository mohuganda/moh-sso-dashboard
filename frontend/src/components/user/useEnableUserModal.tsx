import { useModal } from "../modal/modal.context";

import { EnableUserModal } from "./EnableUserModal";

type OpenEnableUserModalArgs = {
  username: string;
  enabled: boolean;
  onConfirm: () => Promise<void> | void;
};

export function useEnableUserModal() {
  const { openModal, closeModal } = useModal();

  function openEnableUserModal({ username, enabled, onConfirm }: OpenEnableUserModalArgs) {
    openModal({
      title: enabled ? "Disable user" : "Enable user",
      size: "sm",
      content: (
        <EnableUserModal
          username={username}
          enabled={enabled}
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

  return { openEnableUserModal };
}
