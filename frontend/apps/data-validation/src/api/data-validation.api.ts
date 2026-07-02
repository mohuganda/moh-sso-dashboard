import { baseApi } from "@moh-sso/api";
import { API } from "@moh-sso/config";

import type { ValidationRule, ValidationRulePayload } from "../types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

type ValidationRuleResponse = {
  id: number;
  table_id?: string;
  program?: string;
  category?: string;
  code: string;
  severity: ValidationRule["severity"];
  description: string;
  column: string;
  op: ValidationRule["operator"];
  value?: string;
  value_column?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

type ImportValidationRulesResult = {
  imported: number;
  created: number;
  updated: number;
  skipped: number;
  errors: Array<{
    index: number;
    code?: string;
    message: string;
  }>;
  rules: ValidationRuleResponse[];
};

function toValidationRule(rule: ValidationRuleResponse): ValidationRule {
  return {
    id: `custom-${rule.id}`,
    type: "custom",
    code: rule.code,
    severity: rule.severity,
    table: rule.table_id ?? "",
    column: rule.column,
    operator: rule.op,
    compareTo: rule.value_column ? "column" : "value",
    value: rule.value_column ?? rule.value ?? "",
    description: rule.description,
  };
}

export const dataValidationApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listValidationRules: builder.query<ValidationRule[], void>({
      query: () => ({
        url: API.dataValidation.rules.list(),
        method: "GET",
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<ValidationRuleResponse[]>) =>
        res.data.map(toValidationRule),
      providesTags: ["DataValidationRules"],
    }),

    importValidationRules: builder.mutation<ImportValidationRulesResult, ValidationRulePayload[]>({
      query: (body) => ({
        url: API.dataValidation.rules.import(),
        method: "POST",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<ImportValidationRulesResult>) => res.data,
      invalidatesTags: ["DataValidationRules"],
    }),
  }),
});

export const { useListValidationRulesQuery, useImportValidationRulesMutation } = dataValidationApi;
