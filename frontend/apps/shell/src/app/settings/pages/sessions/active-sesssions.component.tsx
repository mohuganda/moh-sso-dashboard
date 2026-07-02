import UserSessionsTable from "@/app/settings/sessions/UserSessionsTable";

export default function ActiveSessionsPage() {
  return (
    <div className="settings-page">
      <div className="settings-page__header">
        <h2>Active sessions</h2>
        <p className="settings-page__description">
          Review active login sessions connected to your account and end sessions you do not
          recognize.
        </p>
      </div>

      <UserSessionsTable />
    </div>
  );
}
