import { ErrorState } from "./ErrorState";

export default {
  title: "Feedback/ErrorState",
};

export const Default = () => {
  return (
    <ErrorState
      title="Failed to load records"
      description="Something went wrong while loading this data."
      primaryAction={{
        label: "Retry",
        onClick: () => alert("Retry clicked"),
      }}
    />
  );
};
