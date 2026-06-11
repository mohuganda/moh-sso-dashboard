import { ComingSoon } from "./ComingSoon";

export default {
  title: "Feedback/ComingSoon",
};

export const Default = () => {
  return <ComingSoon title="Document Upload" />;
};

export const WithCustomDescription = () => {
  return (
    <ComingSoon
      title="Facility Register"
      description="This module is being prepared and will be available soon."
    />
  );
};
