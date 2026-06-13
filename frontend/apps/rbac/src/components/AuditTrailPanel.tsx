import type { RbacAuditEvent } from "@moh-sso/types";

type AuditTrailPanelProps = {
  auditEvents: RbacAuditEvent[];
};

export function AuditTrailPanel({ auditEvents }: AuditTrailPanelProps) {
  return (
    <section>
      <h3>Audit Trail</h3>
      <small>{auditEvents.length} recent RBAC audit events</small>
      <div className="rbac-effective-stack">
        {auditEvents.slice(0, 10).map((event) => (
          <small key={event.id}>
            {event.createdAt} · {event.action}
          </small>
        ))}
      </div>
    </section>
  );
}
