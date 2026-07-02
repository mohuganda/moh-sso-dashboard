import { Stack, Tag } from "@carbon/react";

import type { ValidationRule } from "../types";
import "./validation-rule-details-panel.scss";

type ValidationRuleDetailsPanelProps = {
  rule: ValidationRule;
};

function DetailRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="validation-rule-details__row">
      <dt>{label}</dt>
      <dd>{value || "—"}</dd>
    </div>
  );
}

export function ValidationRuleDetailsPanel({ rule }: ValidationRuleDetailsPanelProps) {
  return (
    <Stack gap={6} className="validation-rule-details">
      <div>
        <Tag type={rule.type === "builtin" ? "gray" : "blue"}>{rule.type}</Tag>
        {rule.severity ? (
          <Tag
            type={
              rule.severity === "error"
                ? "red"
                : rule.severity === "warning"
                  ? "magenta"
                  : "blue"
            }
          >
            {rule.severity}
          </Tag>
        ) : null}
      </div>

      <dl>
        <DetailRow label="Code / Category" value={rule.code} />
        <DetailRow label="Table" value={rule.table} />
        <DetailRow label="Column" value={rule.column} />
        <DetailRow label="Operator" value={rule.operator} />
        <DetailRow label="Compare to" value={rule.compareTo} />
        <DetailRow label="Value" value={rule.value} />
        <DetailRow label="Description" value={rule.description} />
      </dl>
    </Stack>
  );
}
