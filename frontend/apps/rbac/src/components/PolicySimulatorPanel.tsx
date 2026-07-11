import { Button, Tag, TextInput } from "@carbon/react";

import type { RbacSimulationResult } from "../types";

type PolicySimulatorPanelProps = {
  simulationRoles: string;
  simulationClientRoles: string;
  simulationUser: string;
  addRealmRoles: string;
  removeRealmRoles: string;
  addClientRoles: string;
  removeClientRoles: string;
  addPermissions: string;
  removePermissions: string;
  simulation: RbacSimulationResult | null;
  onSimulationRolesChange: (value: string) => void;
  onSimulationClientRolesChange: (value: string) => void;
  onSimulationUserChange: (value: string) => void;
  onAddRealmRolesChange: (value: string) => void;
  onRemoveRealmRolesChange: (value: string) => void;
  onAddClientRolesChange: (value: string) => void;
  onRemoveClientRolesChange: (value: string) => void;
  onAddPermissionsChange: (value: string) => void;
  onRemovePermissionsChange: (value: string) => void;
  onSimulate: () => void;
};

export function PolicySimulatorPanel({
  simulationRoles,
  simulationClientRoles,
  simulationUser,
  addRealmRoles,
  removeRealmRoles,
  addClientRoles,
  removeClientRoles,
  addPermissions,
  removePermissions,
  simulation,
  onSimulationRolesChange,
  onSimulationClientRolesChange,
  onSimulationUserChange,
  onAddRealmRolesChange,
  onRemoveRealmRolesChange,
  onAddClientRolesChange,
  onRemoveClientRolesChange,
  onAddPermissionsChange,
  onRemovePermissionsChange,
  onSimulate,
}: PolicySimulatorPanelProps) {
  return (
    <section className="rbac-governance-wide">
      <h3>Policy Simulator</h3>
      <TextInput
        id="rbac-sim-user"
        labelText="User ID, username, or email"
        value={simulationUser}
        onChange={(event) => onSimulationUserChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-realm"
        labelText="Realm roles, comma separated"
        value={simulationRoles}
        onChange={(event) => onSimulationRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-client"
        labelText="Client roles, comma separated as client:role"
        value={simulationClientRoles}
        onChange={(event) => onSimulationClientRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-add-realm"
        labelText="Add realm roles"
        value={addRealmRoles}
        onChange={(event) => onAddRealmRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-remove-realm"
        labelText="Remove realm roles"
        value={removeRealmRoles}
        onChange={(event) => onRemoveRealmRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-add-client"
        labelText="Add client roles as client:role"
        value={addClientRoles}
        onChange={(event) => onAddClientRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-remove-client"
        labelText="Remove client roles as client:role"
        value={removeClientRoles}
        onChange={(event) => onRemoveClientRolesChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-add-permissions"
        labelText="Add permissions"
        value={addPermissions}
        onChange={(event) => onAddPermissionsChange(event.target.value)}
      />
      <TextInput
        id="rbac-sim-remove-permissions"
        labelText="Remove permissions"
        value={removePermissions}
        onChange={(event) => onRemovePermissionsChange(event.target.value)}
      />
      <Button size="sm" onClick={onSimulate}>
        Simulate access
      </Button>
      {simulation && (
        <div className="rbac-effective-stack">
          <div className="rbac-effective-tags">
            {simulation.permissions.map((permission) => (
              <Tag key={permission.key} type="green">
                {permission.key}
              </Tag>
            ))}
          </div>
          <small>Added: {simulation.addedPermissions.join(", ") || "none"}</small>
          <small>Removed: {simulation.removedPermissions.join(", ") || "none"}</small>
        </div>
      )}
    </section>
  );
}
