import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function WorkplansPage() {
  return (
    <EservicePageShell
      title="Workplans"
      description="Create, review, approve, and monitor departmental or program workplans."
      actionLabel="Create workplan"
      onActionClick={() => {
        console.log("Create workplan");
      }}
    >
      <EserviceEmptyState
        title="No workplans available"
        description="Program and departmental workplans will appear here after they are created or submitted."
        actionLabel="Create first workplan"
        onActionClick={() => {
          console.log("Create first workplan");
        }}
      />
    </EservicePageShell>
  );
}
