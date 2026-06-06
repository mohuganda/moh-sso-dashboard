import { useModal } from "@moh-sso/ui";

import { EnableClientModal } from "./EnableClientModal";

type OpenEnableClientModalArgs = {
  clientName: string;
  enabled: boolean;
  onConfirm: () => Promise<void> | void;
};

export function useEnableClientModal() {
  const { openModal, closeModal } = useModal();

  function openEnableClientModal({ clientName, enabled, onConfirm }: OpenEnableClientModalArgs) {
    openModal({
      title: enabled ? "Disable client" : "Enable client",
      size: "sm",
      content: (
        <EnableClientModal
          clientName={clientName}
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

  return { openEnableClientModal };
}
