import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function MeetingManagerPage() {
  return (
    <EservicePageShell
      title="Meeting Manager"
      description="Plan meetings, manage agendas, track attendance, and document meeting resolutions."
      actionLabel="Create meeting"
      onActionClick={() => {
        console.log("Create meeting");
      }}
    >
      <EserviceEmptyState
        title="No meetings available"
        description="Create and manage meetings for departments, programs, partners, and internal coordination."
        actionLabel="Create first meeting"
        onActionClick={() => {
          console.log("Create first meeting");
        }}
      />
    </EservicePageShell>
  );
}
