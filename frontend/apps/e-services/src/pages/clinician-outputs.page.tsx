import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function ClinicianOutputsPage() {
  return (
    <EservicePageShell
      title="Clinician Outputs"
      description="Monitor clinician outputs, performance indicators, reports, and service delivery contributions."
      actionLabel="Add output"
      onActionClick={() => {
        console.log("Add clinician output");
      }}
    >
      <EserviceEmptyState
        title="No clinician outputs submitted"
        description="Clinician output records and performance summaries will be displayed here once submitted."
        actionLabel="Add clinician output"
        onActionClick={() => {
          console.log("Add clinician output");
        }}
      />
    </EservicePageShell>
  );
}
