import { InlineNotification } from "@carbon/react";

type Props = {
  kind?: "error" | "warning" | "info" | "success";
  title: string;
  subtitle?: string;
};

export function FormInlineAlert({ kind = "error", title, subtitle }: Props) {
  return (
    <InlineNotification kind={kind} title={title} subtitle={subtitle} lowContrast hideCloseButton />
  );
}
