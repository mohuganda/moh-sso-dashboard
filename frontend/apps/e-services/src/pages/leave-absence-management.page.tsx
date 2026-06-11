import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function LeaveAbsenceManagementPage() {
  return (
    <EservicePageShell
      title="Leave & Absence Management"
      description="Manage leave requests, approvals, absence records, staff availability, and leave balances."
      actionLabel="Request leave"
      onActionClick={() => undefined}
    >
      <EserviceEmptyState
        title="No leave requests found"
        description="Leave requests, absence records, approvals, and staff availability information will appear here."
        actionLabel="Create leave request"
        onActionClick={() => undefined}
      />
    </EservicePageShell>
  );
}
