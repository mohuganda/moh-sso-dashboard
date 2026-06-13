import { Button, Tag, TextInput } from "@carbon/react";

import type { RbacSimulationResult } from "@moh-sso/types";

type PolicySimulatorPanelProps = {
  simulationRoles: string;
  simulationClientRoles: string;
  simulation: RbacSimulationResult | null;
  onSimulationRolesChange: (value: string) => void;
  onSimulationClientRolesChange: (value: string) => void;
  onSimulate: () => void;
};

export function PolicySimulatorPanel({
  simulationRoles,
  simulationClientRoles,
  simulation,
  onSimulationRolesChange,
  onSimulationClientRolesChange,
  onSimulate,
}: PolicySimulatorPanelProps) {
  return (
    <section className="rbac-governance-wide">
      <h3>Policy Simulator</h3>
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
      <Button size="sm" onClick={onSimulate}>
        Simulate access
      </Button>
      {simulation && (
        <div className="rbac-effective-tags">
          {simulation.permissions.map((permission) => (
            <Tag key={permission.key} type="green">
              {permission.key}
            </Tag>
          ))}
        </div>
      )}
    </section>
  );
}
