import { EserviceEmptyState } from "../components/eservice-empty-state.component";
import { EservicePageShell } from "../components/eservice-page-shell.component";

export default function BudgetTrackerPage() {
  return (
    <EservicePageShell
      title="Budget Tracker"
      description="Track budgets, allocations, expenditure, balances, and financial performance by activity or program."
      actionLabel="Add budget item"
      onActionClick={() => {
        console.log("Add budget item");
      }}
    >
      <EserviceEmptyState
        title="No budget records found"
        description="Budget allocations, expenditure summaries, and financial tracking records will appear here."
        actionLabel="Add budget item"
        onActionClick={() => {
          console.log("Add budget item");
        }}
      />
    </EservicePageShell>
  );
}
