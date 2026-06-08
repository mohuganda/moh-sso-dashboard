import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function ActionTrackerPage() {
  return (
    <EservicePageShell
      title="Action Tracker"
      description="Track action points, responsible persons, deadlines, progress, and completion status."
      actionLabel="New action"
      onActionClick={() => {
        console.log("New action");
      }}
    >
      <EserviceEmptyState
        title="No action points found"
        description="Action points from meetings, supervision visits, reviews, and program activities will appear here."
        actionLabel="Create action point"
        onActionClick={() => {
          console.log("Create action point");
        }}
      />
    </EservicePageShell>
  );
}
