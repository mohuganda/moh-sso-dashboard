import { Select, SelectItem } from "@carbon/react";

import { useHealthContext } from "@moh-sso/auth";
import { useToast } from "@moh-sso/ui";

import "./health-context-selector.scss";

export function HealthContextSelector() {
  const { contexts, activeContext, isLoading, isUpdating, selectContext } = useHealthContext();
  const toast = useToast();

  if (!isLoading && contexts.length <= 1) {
    return null;
  }

  return (
    <div className="health-context-selector">
      <Select
        id="global-health-context"
        labelText="Health context"
        hideLabel
        size="sm"
        value={activeContext?.id ?? ""}
        disabled={isLoading || isUpdating || contexts.length === 0}
        onChange={(event) => {
          void selectContext(event.target.value).catch(() => {
            toast.error("Context change failed", "Your previous health context remains active.");
          });
        }}
      >
        {contexts.length === 0 ? (
          <SelectItem value="" text="No health context" />
        ) : null}
        {contexts.map((context) => (
          <SelectItem
            key={context.id}
            value={context.id}
            text={`${context.name} (${context.contextType.replaceAll("_", " ").toLowerCase()})`}
          />
        ))}
      </Select>
    </div>
  );
}
