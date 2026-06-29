import { Button, Tag } from "@carbon/react";

import { useDecideAccessRequestMutation } from "../api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacAccessRequest } from "../types";

type AccessRequestsPanelProps = {
  accessRequests: RbacAccessRequest[];
};

export function AccessRequestsPanel({ accessRequests }: AccessRequestsPanelProps) {
  const [decideAccessRequest] = useDecideAccessRequestMutation();

  const handleDecision = async (id: string, decision: "approved" | "rejected" | "cancelled") => {
    await decideAccessRequest({ id, decision, data: { note: `Marked ${decision} from RBAC console` } }).unwrap();
  };

  return (
    <section>
      <h3>Access Requests</h3>
      <small>{accessRequests.length} requests awaiting review or history.</small>
      <div className="rbac-effective-stack">
        {accessRequests.slice(0, 8).map((request) => (
          <div className="rbac-request-row" key={request.id}>
            <Tag type={request.status === "pending" ? "blue" : "gray"}>
              {request.systemClientId}:{request.requestedRole}
            </Tag>
            <small>{request.username || request.email || request.userId || "Unknown user"}</small>
            {request.status === "pending" && (
              <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                <div className="rbac-sync-actions">
                  <Button size="sm" kind="primary" onClick={() => handleDecision(request.id, "approved")}>
                    Approve
                  </Button>
                  <Button size="sm" kind="danger--tertiary" onClick={() => handleDecision(request.id, "rejected")}>
                    Reject
                  </Button>
                  <Button size="sm" kind="ghost" onClick={() => handleDecision(request.id, "cancelled")}>
                    Cancel
                  </Button>
                </div>
              </PermissionGuard>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
