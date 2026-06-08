import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function ActivityReportingPage() {
  return (
    <EservicePageShell
      title="Activity Reporting"
      description="Submit, review, and monitor activity reports, outputs, implementation progress, and supporting evidence."
      actionLabel="Submit report"
      onActionClick={() => {
        console.log("Submit report");
      }}
    >
      <EserviceEmptyState
        title="No activity reports submitted"
        description="Submitted activity reports, implementation updates, and supporting documentation will appear here."
        actionLabel="Submit activity report"
        onActionClick={() => {
          console.log("Submit activity report");
        }}
      />
    </EservicePageShell>
  );
}
