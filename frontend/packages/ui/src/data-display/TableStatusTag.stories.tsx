import { TableStatusTag } from "./TableStatusTag";

export default {
  title: "Data Display/TableStatusTag",
};

export const Statuses = () => {
  return (
    <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
      <TableStatusTag status="Active" />
      <TableStatusTag status="Disabled" />
      <TableStatusTag status="Pending" />
      <TableStatusTag status="Unknown" />
    </div>
  );
};
