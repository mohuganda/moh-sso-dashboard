import { Button, Tag } from "@carbon/react";

import { useDecideChangeRequestMutation } from "@moh-sso/api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacChangePreview, RbacChangeRequest } from "@moh-sso/types";

type ChangeRequestsPanelProps = {
  changeRequests: RbacChangeRequest[];
  changePreview?: RbacChangePreview;
  onPreviewRisk: () => void;
  onCreateRequest: () => void;
};

export function ChangeRequestsPanel({
  changeRequests,
  changePreview,
  onPreviewRisk,
  onCreateRequest,
}: ChangeRequestsPanelProps) {
  const [decideChangeRequest] = useDecideChangeRequestMutation();

  const handleDecision = async (id: string, decision: "approved" | "rejected" | "applied") => {
    await decideChangeRequest({ id, decision, data: { note: `Marked ${decision} from RBAC console` } }).unwrap();
  };

  return (
    <section>
      <h3>Impact & Change Requests</h3>
      <div className="rbac-sync-actions">
        <Button size="sm" onClick={onPreviewRisk}>
          Preview risk
        </Button>
        <PermissionGuard permission={PERMISSIONS.rbacWrite}>
          <Button size="sm" onClick={onCreateRequest}>
            Create request
          </Button>
        </PermissionGuard>
      </div>
      {changePreview && <Tag type={changePreview.highRisk ? "red" : "green"}>{changePreview.riskLevel}</Tag>}
      <small>{changeRequests.length} change requests</small>
      <div className="rbac-effective-stack">
        {changeRequests.slice(0, 5).map((request) => (
          <div className="rbac-request-row" key={request.id}>
            <small>
              {request.status} · {request.action} · {request.resourceType}
            </small>
            {request.status === "pending" && (
              <PermissionGuard permission={PERMISSIONS.rbacWrite}>
                <div className="rbac-sync-actions">
                  <Button size="sm" kind="primary" onClick={() => handleDecision(request.id, "approved")}>
                    Approve
                  </Button>
                  <Button size="sm" kind="danger--tertiary" onClick={() => handleDecision(request.id, "rejected")}>
                    Reject
                  </Button>
                </div>
              </PermissionGuard>
            )}
            {request.status === "approved" && (
              <PermissionGuard permission={PERMISSIONS.rbacWrite}>
                <Button size="sm" onClick={() => handleDecision(request.id, "applied")}>
                  Mark applied
                </Button>
              </PermissionGuard>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
