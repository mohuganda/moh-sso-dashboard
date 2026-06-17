import { InlineLoading, Tile } from "@carbon/react";

type LoadingStateProps = {
  description?: string;
};

export function LoadingState({ description = "Loading..." }: LoadingStateProps) {
  return (
    <Tile className="moh-loading-state">
      <InlineLoading description={description} />
    </Tile>
  );
}
