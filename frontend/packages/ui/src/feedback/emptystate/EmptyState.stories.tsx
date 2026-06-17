import { EmptyState } from "./EmptyState";

export default {
  title: "Feedback/EmptyState",
};

export const Default = () => {
  return (
    <EmptyState
      title="No records found"
      description="Try adjusting your filters or search terms."
    />
  );
};
