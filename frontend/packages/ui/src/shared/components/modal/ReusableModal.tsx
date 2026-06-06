import { ComposedModal, ModalHeader, ModalBody, ModalFooter, Button } from "@carbon/react";
import React from "react";

export type ModalSize = "sm" | "md" | "lg";

export type ReusableModalProps = {
  open: boolean;
  title?: string;
  content: React.ReactNode;
  primaryAction?: {
    label: string;
    onClick: () => void;
    kind?: "primary" | "danger";
    disabled?: boolean;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  size?: ModalSize;
  onClose: () => void;
};

export const ReusableModal: React.FC<ReusableModalProps> = ({
  open,
  title,
  content,
  primaryAction,
  secondaryAction,
  size = "md",
  onClose,
}) => {
  return (
    <ComposedModal open={open} onClose={onClose} size={size}>
      {title && <ModalHeader title={title} />}

      <ModalBody>{content}</ModalBody>

      {(primaryAction || secondaryAction) && (
        <ModalFooter>
          {secondaryAction && (
            <Button kind="secondary" onClick={secondaryAction.onClick}>
              {secondaryAction.label}
            </Button>
          )}

          {primaryAction && (
            <Button
              kind={primaryAction.kind ?? "primary"}
              onClick={primaryAction.onClick}
              disabled={primaryAction.disabled}
            >
              {primaryAction.label}
            </Button>
          )}
        </ModalFooter>
      )}
    </ComposedModal>
  );
};
