import {
  Button,
  CodeSnippet,
  InlineNotification,
  Select,
  SelectItem,
  Stack,
  TextArea,
  TextInput,
  Toggle,
} from "@carbon/react";
import { useMemo, useState } from "react";

import { useCompileDQARuleMutation } from "../../api";
import { DQA_RULE_TYPES, type DQARule, type DQARuleInput, type DQARuleType } from "../../dqa.types";

type Props = {
  tableOptions: string[];
  initialRule?: DQARule;
  isSubmitting: boolean;
  onSubmit: (rule: DQARuleInput) => Promise<void> | void;
  onClose: () => void;
};

const DEFINITION_PLACEHOLDER: Record<DQARuleType, string> = {
  row_expression: '{\n  "expr": "total_tested >= total_positive"\n}',
  metric: '{\n  "metric": "missing_percent",\n  "column": "district"\n}',
  row_join: '{\n  "left": "table_a",\n  "right": "table_b",\n  "join_keys": ["vht_id"],\n  "expr": "a.total = b.total"\n}',
  aggregate_compare:
    '{\n  "left": { "table": "table_a", "agg": "SUM(total)" },\n  "right": { "table": "table_b", "agg": "SUM(total)" },\n  "op": "=",\n  "tolerance": 0.5\n}',
  reference: '{\n  "left": "table_a",\n  "right": "table_b",\n  "join_keys": ["vht_id"]\n}',
  raw_sql: '{\n  "fail_query": "SELECT * FROM report.my_table WHERE total < 0"\n}',
};

function safeParse(text: string): { value?: Record<string, unknown>; error?: string } {
  if (!text.trim()) return { value: {} };
  try {
    const parsed = JSON.parse(text);
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
      return { error: "Definition must be a JSON object." };
    }
    return { value: parsed as Record<string, unknown> };
  } catch (e) {
    return { error: e instanceof Error ? e.message : "Invalid JSON" };
  }
}

export function DQARuleForm({ tableOptions, initialRule, isSubmitting, onSubmit, onClose }: Props) {
  const [code, setCode] = useState(initialRule?.code ?? "");
  const [tableId, setTableId] = useState(initialRule?.table_id ?? tableOptions[0] ?? "");
  const [type, setType] = useState<DQARuleType>(initialRule?.type ?? "row_expression");
  const [category, setCategory] = useState(initialRule?.category ?? "custom");
  const [name, setName] = useState(initialRule?.name ?? "");
  const [rowFilter, setRowFilter] = useState(initialRule?.row_filter ?? "");
  const [warn, setWarn] = useState(initialRule?.zones?.warn ?? "");
  const [fail, setFail] = useState(initialRule?.zones?.fail ?? "when > 0");
  const [enabled, setEnabled] = useState(initialRule?.enabled ?? true);
  const [definitionText, setDefinitionText] = useState(
    initialRule ? JSON.stringify(initialRule.definition ?? {}, null, 2) : DEFINITION_PLACEHOLDER[type],
  );

  const [compileDQARule, { isLoading: isCompiling }] = useCompileDQARuleMutation();
  const [preview, setPreview] = useState<{ sql?: string; sample?: string; error?: string } | null>(null);

  const activeTypeMeta = useMemo(() => DQA_RULE_TYPES.find((t) => t.value === type), [type]);

  const buildInput = (): { input?: DQARuleInput; error?: string } => {
    if (!code.trim()) return { error: "Code is required." };
    if (!tableId.trim()) return { error: "Table is required." };
    const parsed = safeParse(definitionText);
    if (parsed.error) return { error: `Definition: ${parsed.error}` };
    return {
      input: {
        code: code.trim(),
        table_id: tableId.trim(),
        type,
        category: category.trim() || "custom",
        name: name.trim(),
        enabled,
        row_filter: rowFilter.trim() || undefined,
        zones: { warn: warn.trim(), fail: fail.trim() },
        definition: parsed.value ?? {},
      },
    };
  };

  const handleCompile = async () => {
    const { input, error } = buildInput();
    if (error || !input) {
      setPreview({ error });
      return;
    }
    try {
      const result = await compileDQARule(input).unwrap();
      setPreview({ sql: result.measure_sql, sample: result.sample_sql });
    } catch (e) {
      const message =
        e && typeof e === "object" && "data" in e
          ? // eslint-disable-next-line @typescript-eslint/no-explicit-any
            ((e as any).data?.error?.message ?? "Compile failed")
          : "Compile failed";
      setPreview({ error: String(message) });
    }
  };

  const handleSave = async () => {
    const { input, error } = buildInput();
    if (error || !input) {
      setPreview({ error });
      return;
    }
    await onSubmit(input);
  };

  return (
    <Stack gap={5}>
      <TextInput
        id="dqa-rule-code"
        labelText="Code"
        placeholder="e.g. MAL-04"
        value={code}
        disabled={!!initialRule}
        onChange={(e) => setCode(e.target.value)}
      />

      {tableOptions.length > 0 ? (
        <Select
          id="dqa-rule-table"
          labelText="Table"
          value={tableId}
          disabled={!!initialRule}
          onChange={(e) => setTableId(e.target.value)}
        >
          {tableOptions.map((t) => (
            <SelectItem key={t} value={t} text={t} />
          ))}
        </Select>
      ) : (
        <TextInput
          id="dqa-rule-table-text"
          labelText="Table ID"
          helperText="No tables are registered yet — register one in the Tables tab first."
          value={tableId}
          disabled={!!initialRule}
          onChange={(e) => setTableId(e.target.value)}
        />
      )}

      <Select
        id="dqa-rule-type"
        labelText="Rule type"
        value={type}
        disabled={!!initialRule}
        onChange={(e) => {
          const next = e.target.value as DQARuleType;
          setType(next);
          if (!initialRule) setDefinitionText(DEFINITION_PLACEHOLDER[next]);
        }}
      >
        {DQA_RULE_TYPES.map((t) => (
          <SelectItem key={t.value} value={t.value} text={t.label} />
        ))}
      </Select>
      {activeTypeMeta ? (
        <p style={{ fontSize: "0.75rem", color: "var(--cds-text-secondary)", marginTop: "-1rem" }}>
          {activeTypeMeta.description}
        </p>
      ) : null}

      <TextInput
        id="dqa-rule-category"
        labelText="Category"
        placeholder="e.g. malaria"
        value={category}
        onChange={(e) => setCategory(e.target.value)}
      />

      <TextInput
        id="dqa-rule-name"
        labelText="Name / description"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />

      <TextInput
        id="dqa-rule-row-filter"
        labelText="Row filter (optional WHERE predicate)"
        placeholder="e.g. district IS NOT NULL"
        value={rowFilter}
        onChange={(e) => setRowFilter(e.target.value)}
      />

      <TextArea
        id="dqa-rule-definition"
        labelText="Definition (JSON)"
        rows={6}
        value={definitionText}
        onChange={(e) => setDefinitionText(e.target.value)}
      />

      <Stack orientation="horizontal" gap={5}>
        <TextInput
          id="dqa-rule-warn"
          labelText="Warn zone"
          placeholder="when > 5"
          value={warn}
          onChange={(e) => setWarn(e.target.value)}
        />
        <TextInput
          id="dqa-rule-fail"
          labelText="Fail zone"
          placeholder="when > 0"
          value={fail}
          onChange={(e) => setFail(e.target.value)}
        />
      </Stack>

      <Toggle
        id="dqa-rule-enabled"
        labelText="Enabled"
        toggled={enabled}
        onToggle={(v) => setEnabled(v)}
      />

      <Button kind="tertiary" onClick={handleCompile} disabled={isCompiling}>
        {isCompiling ? "Compiling…" : "Preview compiled SQL"}
      </Button>

      {preview?.error ? (
        <InlineNotification kind="error" title="Compile failed" subtitle={preview.error} lowContrast hideCloseButton />
      ) : null}
      {preview?.sql ? (
        <div>
          <p style={{ fontSize: "0.75rem", marginBottom: "0.25rem" }}>Measure SQL</p>
          <CodeSnippet type="multi" feedback="Copied">
            {preview.sql}
          </CodeSnippet>
          {preview.sample ? (
            <>
              <p style={{ fontSize: "0.75rem", margin: "0.75rem 0 0.25rem" }}>Sample SQL (failing rows)</p>
              <CodeSnippet type="multi" feedback="Copied">
                {preview.sample}
              </CodeSnippet>
            </>
          ) : null}
        </div>
      ) : null}

      <Stack orientation="horizontal" gap={4}>
        <Button kind="primary" onClick={handleSave} disabled={isSubmitting}>
          {isSubmitting ? "Saving…" : initialRule ? "Save changes" : "Create rule"}
        </Button>
        <Button kind="ghost" onClick={onClose}>
          Cancel
        </Button>
      </Stack>
    </Stack>
  );
}
