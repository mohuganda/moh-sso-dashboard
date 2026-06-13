import { Button, Tag, TextInput } from "@carbon/react";
import { useState } from "react";

import { useCreateRoleFromTemplateMutation } from "@moh-sso/api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacRoleTemplate } from "@moh-sso/types";

type RoleTemplatesPanelProps = {
  templates: RbacRoleTemplate[];
  activeClientId: string;
  onMessage: (message: string) => void;
};

export function RoleTemplatesPanel({ templates, activeClientId, onMessage }: RoleTemplatesPanelProps) {
  const [selectedTemplate, setSelectedTemplate] = useState("");
  const [roleName, setRoleName] = useState("");
  const [createRoleFromTemplate] = useCreateRoleFromTemplateMutation();

  const handleCreateRole = async () => {
    if (!activeClientId || !selectedTemplate) return;
    await createRoleFromTemplate({
      clientId: activeClientId,
      data: { templateName: selectedTemplate, roleName },
    }).unwrap();
    onMessage("Role created from template.");
    setRoleName("");
  };

  return (
    <section>
      <h3>Templates & Bulk</h3>
      <div className="rbac-effective-tags">
        {templates.map((template) => (
          <Tag
            key={template.name}
            type={selectedTemplate === template.name ? "blue" : "purple"}
            onClick={() => setSelectedTemplate(template.name)}
          >
            {template.displayName}
          </Tag>
        ))}
      </div>
      <TextInput
        id="rbac-template-role-name"
        labelText="Role name override"
        placeholder={selectedTemplate || "viewer"}
        value={roleName}
        onChange={(event) => setRoleName(event.target.value)}
      />
      <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
        <Button size="sm" disabled={!activeClientId || !selectedTemplate} onClick={handleCreateRole}>
          Create role from template
        </Button>
      </PermissionGuard>
      <small>
        {activeClientId
          ? `Templates apply to ${activeClientId}.`
          : "Select a system before applying a template."}
      </small>
    </section>
  );
}
