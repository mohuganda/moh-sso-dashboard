export type RuleType = "builtin" | "custom";
export type Severity = "error" | "warning" | "info";
export type Operator = "contains" | "eq" | "gt" | "gte" | "isnull" | "lt" | "lte" | "ne" | "notnull";

export type ValidationRule = {
  id: string;
  type: RuleType;
  code: string;
  severity?: Severity;
  table?: string;
  column?: string;
  operator?: Operator;
  compareTo?: string;
  value?: string;
  description: string;
};

export type RuleFormState = {
  table: string;
  code: string;
  severity: Severity;
  column: string;
  operator: Operator;
  compareTo: string;
  value: string;
  description: string;
};

export type ValidationRulePayload = {
  table_id?: string;
  category: "custom";
  code: string;
  severity: Severity;
  description: string;
  column: string;
  op: Operator;
  value?: string;
  value_column?: string;
};
