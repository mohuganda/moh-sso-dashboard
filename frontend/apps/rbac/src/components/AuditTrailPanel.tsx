import { TextInput } from "@carbon/react";
import { useState } from "react";

import { useListRbacAuditEventsQuery } from "@moh-sso/api";

export function AuditTrailPanel() {
  const [actor, setActor] = useState("");
  const [systemClientId, setSystemClientId] = useState("");
  const [roleName, setRoleName] = useState("");
  const [permissionKey, setPermissionKey] = useState("");
  const [action, setAction] = useState("");
  const { data: auditEvents = [] } = useListRbacAuditEventsQuery({
    actor,
    systemClientId,
    roleName,
    permissionKey,
    action,
    limit: 100,
  });

  return (
    <section>
      <h3>Audit Trail</h3>
      <div className="rbac-audit-filters">
        <TextInput id="rbac-audit-actor" labelText="Actor" value={actor} onChange={(event) => setActor(event.target.value)} />
        <TextInput id="rbac-audit-system" labelText="System" value={systemClientId} onChange={(event) => setSystemClientId(event.target.value)} />
        <TextInput id="rbac-audit-role" labelText="Role" value={roleName} onChange={(event) => setRoleName(event.target.value)} />
        <TextInput id="rbac-audit-permission" labelText="Permission" value={permissionKey} onChange={(event) => setPermissionKey(event.target.value)} />
        <TextInput id="rbac-audit-action" labelText="Action" value={action} onChange={(event) => setAction(event.target.value)} />
      </div>
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
