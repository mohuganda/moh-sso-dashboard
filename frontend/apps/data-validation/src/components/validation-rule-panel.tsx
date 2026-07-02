import {
  Button,
  Form,
  FormGroup,
  Select,
  SelectItem,
  Stack,
  TextArea,
  TextInput,
} from "@carbon/react";
import { useMemo, useState } from "react";

import type { Operator, RuleFormState, Severity, ValidationRule } from "../types";

type ValidationRulePanelProps = {
  mode?: "create" | "edit";
  initialCode?: string;
  initialRule?: ValidationRule;
  onSubmit: (rule: ValidationRule) => void;
  onClose: () => void;
};

const tableOptions = ["cht_form_097b", "facility_weekly_metrics", "case_investigations", "stock_status"];
const columnOptions = [
  "org_unit",
  "reporting_period",
  "malaria",
  "diarrhoea",
  "pneumonia",
  "danger_signs",
  "pregnancy",
  "family_planning",
  "activity",
  "commodities",
  "submitted_at",
];

const defaultFormState: RuleFormState = {
  table: "",
  code: "",
  severity: "error",
  column: "",
  operator: "contains",
  compareTo: "value",
  value: "",
  description: "",
};

function createFormState(initialCode?: string, initialRule?: ValidationRule): RuleFormState {
  return {
    ...defaultFormState,
    table: initialRule?.table ?? "",
    code: initialRule?.code ?? initialCode ?? "",
    severity: initialRule?.severity ?? "error",
    column: initialRule?.column ?? "",
    operator: initialRule?.operator ?? "contains",
    compareTo: initialRule?.compareTo ?? "value",
    value: initialRule?.value ?? "",
    description: initialRule?.description ?? "",
  };
}

function createRule(form: RuleFormState, initialRule?: ValidationRule): ValidationRule {
  return {
    id: initialRule?.id ?? `custom-${crypto.randomUUID()}`,
    type: initialRule?.type ?? "custom",
    code: form.code.trim(),
    severity: form.severity,
    table: form.table,
    column: form.column,
    operator: form.operator,
    compareTo: form.compareTo,
    value: form.value.trim(),
    description: form.description.trim(),
  };
}

export function ValidationRulePanel({
  mode = "create",
  initialCode,
  initialRule,
  onSubmit,
  onClose,
}: ValidationRulePanelProps) {
  const [form, setForm] = useState<RuleFormState>(() => createFormState(initialCode, initialRule));

  const isValid = useMemo(() => {
    return Boolean(form.code.trim() && form.table && form.column && form.description.trim());
  }, [form]);

  const updateForm = <Key extends keyof RuleFormState>(key: Key, value: RuleFormState[Key]) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const handleSubmit = () => {
    if (!isValid) {
      return;
    }

    onSubmit(createRule(form, initialRule));
    onClose();
  };

  return (
    <Form className="data-validation-panel">
      <Stack gap={7}>
        <FormGroup legendText="Rule definition">
          <Stack gap={4}>
            <Select
              id="validation-table"
              labelText="Table"
              value={form.table}
              onChange={(event) => updateForm("table", event.target.value)}
            >
              <SelectItem value="" text="Select table..." />
              {tableOptions.map((table) => (
                <SelectItem key={table} value={table} text={table} />
              ))}
            </Select>

            <TextInput
              id="validation-code"
              labelText="Code"
              value={form.code}
              onChange={(event) => updateForm("code", event.target.value)}
            />

            <Select
              id="validation-severity"
              labelText="Severity"
              value={form.severity}
              onChange={(event) => updateForm("severity", event.target.value as Severity)}
            >
              <SelectItem value="error" text="error" />
              <SelectItem value="warning" text="warning" />
              <SelectItem value="info" text="info" />
            </Select>

            <Select
              id="validation-column"
              labelText="Column"
              value={form.column}
              onChange={(event) => updateForm("column", event.target.value)}
            >
              <SelectItem value="" text="Select column..." />
              {columnOptions.map((column) => (
                <SelectItem key={column} value={column} text={column} />
              ))}
            </Select>
          </Stack>
        </FormGroup>

        <FormGroup legendText="Condition">
          <Stack gap={4}>
            <Select
              id="validation-operator"
              labelText="Operator"
              value={form.operator}
              onChange={(event) => updateForm("operator", event.target.value as Operator)}
            >
              <SelectItem value="contains" text="contains" />
              <SelectItem value="eq" text="equals" />
              <SelectItem value="gt" text="greater than" />
              <SelectItem value="gte" text="greater or equal" />
              <SelectItem value="isnull" text="is null" />
              <SelectItem value="lt" text="less than" />
              <SelectItem value="lte" text="less or equal" />
              <SelectItem value="ne" text="not equals" />
              <SelectItem value="notnull" text="is not null" />
            </Select>

            <Select
              id="validation-compare-to"
              labelText="Compare to"
              value={form.compareTo}
              onChange={(event) => updateForm("compareTo", event.target.value)}
            >
              <SelectItem value="value" text="Value" />
              <SelectItem value="column" text="Column" />
            </Select>

            <TextInput
              id="validation-value"
              labelText="Value"
              placeholder="e.g. 0, 100, text"
              value={form.value}
              onChange={(event) => updateForm("value", event.target.value)}
            />

            <TextArea
              id="validation-description"
              labelText="Description"
              placeholder="Short description shown in flags"
              rows={4}
              value={form.description}
              onChange={(event) => updateForm("description", event.target.value)}
            />

            <p className="data-validation-panel__hint">
              Operators: contains=string contains; eq=equals; gt=greater than; gte=greater
              or equal; isnull=value is null; lt=less than; lte=less or equal; ne=not
              equals; notnull=value is not null.
            </p>
          </Stack>
        </FormGroup>

        <Stack orientation="horizontal" gap={3}>
          <Button type="button" disabled={!isValid} onClick={handleSubmit}>
            {mode === "edit" ? "Save changes" : "Add rule"}
          </Button>
          <Button kind="secondary" type="button" onClick={onClose}>
            Cancel
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
