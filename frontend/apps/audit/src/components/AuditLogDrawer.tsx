import { Button, CodeSnippet, Tag } from "@carbon/react";
import { Close } from "@carbon/react/icons";
import type { ReactNode } from "react";

import type { AuditLog } from "../types";

import "./AuditLogDrawer.scss";

type JsonRecord = Record<string, unknown>;

const SENSITIVE_KEY_PATTERN =
  /(authorization|cookie|password|secret|token|verifier|credential|session|refresh|access_token|id_token)/i;

const REQUEST_ID_KEYS = ["request_id", "requestId", "correlation_id", "correlationId", "trace_id", "traceId"];
const PATH_KEYS = ["path", "url", "request_path", "requestPath", "route"];
const METHOD_KEYS = ["method", "request_method", "requestMethod"];
const STATUS_KEYS = ["status", "status_code", "statusCode"];
const LATENCY_KEYS = ["latency_ms", "latencyMs", "duration_ms", "durationMs"];
const BEFORE_KEYS = ["before", "before_state", "previous", "old", "old_value", "oldValue"];
const AFTER_KEYS = ["after", "after_state", "current", "new", "new_value", "newValue"];

export function AuditLogPanel({ log, onClose }: { log: AuditLog | null; onClose: () => void }) {
  if (!log) return null;

  const metadata = toRecord(log.metadata);
  const success = typeof log.success === "boolean" ? log.success : booleanValue(metadata.success);
  const requestId = firstString(metadata, REQUEST_ID_KEYS);
  const requestPath = firstString(metadata, PATH_KEYS);
  const method = firstString(metadata, METHOD_KEYS);
  const status = firstString(metadata, STATUS_KEYS);
  const latency = firstString(metadata, LATENCY_KEYS);
  const before = firstPresent(metadata, BEFORE_KEYS);
  const after = firstPresent(metadata, AFTER_KEYS);
  const changes = diffRecords(before, after);
  const safeMetadata = omitKeys(redactSensitive(metadata), [
    ...BEFORE_KEYS,
    ...AFTER_KEYS,
  ]);

  return (
    <div className="audit-log-panel">
      <div className="audit-log-panel__header">
        <div>
          <p className="audit-log-panel__eyebrow">Audit event</p>
          <h3>{log.action}</h3>
          <p>{formatDate(log.createdAt)}</p>
        </div>

        <Button
          kind="ghost"
          size="sm"
          hasIconOnly
          renderIcon={Close}
          iconDescription="Close"
          onClick={onClose}
        />
      </div>

      <div className="audit-log-panel__tags">
        {success !== null && (
          <Tag type={success ? "green" : "red"}>{success ? "success" : "failure"}</Tag>
        )}
        {log.clientId && <Tag type="cyan">{log.clientId}</Tag>}
        {requestId && <Tag type="purple">request {requestId}</Tag>}
      </div>

      <section className="audit-log-panel__section">
        <h4>Event Summary</h4>
        <div className="audit-log-panel__grid">
          <Field label="ID" value={log.id} />
          <Field label="Actor" value={log.username || "System"} />
          <Field label="User ID" value={log.userId || "-"} />
          <Field label="Client" value={log.clientId || stringValue(metadata.client_id) || "-"} />
          <Field label="IP Address" value={log.ip || stringValue(metadata.ip) || "-"} />
          <Field label="User Agent" value={stringValue(metadata.user_agent) || "-"} />
        </div>
      </section>

      <section className="audit-log-panel__section">
        <h4>Request Context</h4>
        <div className="audit-log-panel__grid">
          <Field label="Request ID" value={requestId || "-"} />
          <Field label="Method" value={method || "-"} />
          <Field label="Path" value={requestPath || "-"} />
          <Field label="Status" value={status || "-"} />
          <Field label="Latency" value={latency ? `${latency} ms` : "-"} />
          <Field
            label="Location"
            value={`${stringValue(metadata.country) || "-"} / ${stringValue(metadata.city) || "-"}`}
          />
        </div>
      </section>

      {(before !== undefined || after !== undefined) && (
        <section className="audit-log-panel__section">
          <h4>Before / After</h4>
          {changes.length > 0 && (
            <div className="audit-log-panel__changes">
              {changes.map((change) => (
                <div className="audit-log-panel__change" key={change.path}>
                  <strong>{change.path}</strong>
                  <span>{stringifyCompact(change.before)}</span>
                  <span>{stringifyCompact(change.after)}</span>
                </div>
              ))}
            </div>
          )}
          <div className="audit-log-panel__diff">
            <DiffBlock title="Before" value={before} />
            <DiffBlock title="After" value={after} />
          </div>
        </section>
      )}

      <section className="audit-log-panel__section">
        <h4>Safe Metadata</h4>
        <CodeSnippet type="multi">{JSON.stringify(safeMetadata, null, 2)}</CodeSnippet>
      </section>
    </div>
  );
}

function Field({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="audit-log-panel__field">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function DiffBlock({ title, value }: { title: string; value: unknown }) {
  return (
    <div className="audit-log-panel__diff-block">
      <span>{title}</span>
      <CodeSnippet type="multi">{JSON.stringify(redactSensitive(value), null, 2)}</CodeSnippet>
    </div>
  );
}

function toRecord(value: unknown): JsonRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  return value as JsonRecord;
}

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "-" : date.toLocaleString();
}

function stringValue(value: unknown): string {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  return "";
}

function booleanValue(value: unknown): boolean | null {
  if (typeof value === "boolean") return value;
  if (value === "true") return true;
  if (value === "false") return false;
  return null;
}

function firstString(record: JsonRecord, keys: string[]) {
  for (const key of keys) {
    const value = stringValue(record[key]);
    if (value) return value;
  }

  return "";
}

function firstPresent(record: JsonRecord, keys: string[]) {
  for (const key of keys) {
    if (record[key] !== undefined && record[key] !== null) {
      return record[key];
    }
  }

  return undefined;
}

function omitKeys(record: unknown, keys: string[]) {
  const source = toRecord(record);
  const blocked = new Set(keys);

  return Object.fromEntries(Object.entries(source).filter(([key]) => !blocked.has(key)));
}

function redactSensitive(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(redactSensitive);
  }

  if (!value || typeof value !== "object") {
    return value;
  }

  return Object.fromEntries(
    Object.entries(value as JsonRecord).map(([key, current]) => [
      key,
      SENSITIVE_KEY_PATTERN.test(key) ? "[redacted]" : redactSensitive(current),
    ]),
  );
}

type Change = {
  path: string;
  before: unknown;
  after: unknown;
};

function diffRecords(before: unknown, after: unknown): Change[] {
  const previous = toRecord(before);
  const current = toRecord(after);
  const keys = Array.from(new Set([...Object.keys(previous), ...Object.keys(current)])).sort();

  return keys
    .filter((key) => JSON.stringify(previous[key]) !== JSON.stringify(current[key]))
    .slice(0, 12)
    .map((key) => ({
      path: key,
      before: redactSensitive(previous[key]),
      after: redactSensitive(current[key]),
    }));
}

function stringifyCompact(value: unknown) {
  if (value === undefined) return "—";
  if (value === null) return "null";
  if (typeof value === "string") return value || "—";
  if (typeof value === "number" || typeof value === "boolean") return String(value);

  return JSON.stringify(value);
}
