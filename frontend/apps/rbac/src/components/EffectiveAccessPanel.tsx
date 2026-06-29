import { useMemo, useState } from "react";
import {
  Button,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Tag,
  TextInput,
} from "@carbon/react";
import { Search } from "@carbon/react/icons";

import { useGetRbacEffectiveAccessQuery } from "../api";
import type { RbacEffectiveAccess } from "@moh-sso/types";

type LookupMode = "userId" | "username" | "email";

export function EffectiveAccessPanel() {
  const [lookupMode, setLookupMode] = useState<LookupMode>("username");
  const [lookupValue, setLookupValue] = useState("");
  const [submitted, setSubmitted] = useState<{ mode: LookupMode; value: string } | null>(null);

  const queryArgs = useMemo(() => {
    const value = submitted?.value.trim();
    if (!submitted || !value) return undefined;
    if (submitted.mode === "userId") return { userId: value };
    if (submitted.mode === "email") return { email: value };
    return { username: value };
  }, [submitted]);

  const {
    data,
    isFetching,
    error,
  } = useGetRbacEffectiveAccessQuery(queryArgs ?? {}, { skip: !queryArgs });

  const handleSearch = () => {
    const value = lookupValue.trim();
    if (!value) return;
    setSubmitted({ mode: lookupMode, value });
  };

  return (
    <section className="rbac-card">
      <div className="rbac-page__panel-header">
        <div>
          <h2>Effective Access</h2>
          <p>Explain resolved permissions from Keycloak realm roles and client roles.</p>
        </div>
      </div>

      <div className="rbac-effective-search">
        <Select
          id="rbac-effective-access-mode"
          labelText="Lookup type"
          hideLabel
          value={lookupMode}
          onChange={(event) => setLookupMode(event.target.value as LookupMode)}
        >
          <SelectItem value="username" text="Username" />
          <SelectItem value="email" text="Email" />
          <SelectItem value="userId" text="User ID" />
        </Select>
        <TextInput
          id="rbac-effective-access-search"
          hideLabel
          labelText="User lookup"
          placeholder={lookupMode === "userId" ? "Keycloak user UUID" : lookupMode}
          value={lookupValue}
          onChange={(event) => setLookupValue(event.target.value)}
        />
        <Button renderIcon={Search} onClick={handleSearch}>
          Explain access
        </Button>
      </div>

      {isFetching && <InlineLoading description="Resolving effective access..." />}

      {error && (
        <InlineNotification
          kind="error"
          lowContrast
          title="Unable to resolve access"
          subtitle="Check the user identifier and try again."
        />
      )}

      {data && <EffectiveAccessResult access={data} />}
    </section>
  );
}

function EffectiveAccessResult({ access }: { access: RbacEffectiveAccess }) {
  return (
    <div className="rbac-effective-result">
      <div className="rbac-effective-user">
        <div>
          <strong>{access.user.fullName || access.user.username}</strong>
          <small>{access.user.email || access.user.id}</small>
        </div>
        <Tag type={access.user.enabled ? "green" : "red"}>
          {access.user.enabled ? "Enabled" : "Disabled"}
        </Tag>
        {access.user.isAdmin && <Tag type="purple">Admin</Tag>}
      </div>

      <div className="rbac-effective-grid">
        <section>
          <h3>Realm Roles</h3>
          <TagList values={access.realmRoles} type="blue" />
        </section>
        <section>
          <h3>Accessible Systems</h3>
          <div className="rbac-effective-stack">
            {access.accessibleSystems.map((system) => (
              <div className="rbac-effective-card" key={system.clientId}>
                <strong>{system.displayName || system.clientId}</strong>
                <small>{system.clientId}</small>
                <TagList values={system.roles} type="cyan" />
              </div>
            ))}
            {access.accessibleSystems.length === 0 && <small>No accessible systems resolved.</small>}
          </div>
        </section>
        <section>
          <h3>Client Roles</h3>
          <div className="rbac-effective-stack">
            {Object.entries(access.clientRoles).map(([clientId, roles]) => (
              <div className="rbac-effective-card" key={clientId}>
                <strong>{clientId}</strong>
                <TagList values={roles} type="gray" />
              </div>
            ))}
          </div>
        </section>
        <section>
          <h3>Permissions</h3>
          <TagList values={access.permissions.map((permission) => permission.key)} type="green" />
        </section>
      </div>

      <section className="rbac-effective-grants">
        <h3>Granted By</h3>
        {access.grantSources.map((source) => (
          <div
            className="rbac-effective-grant"
            key={`${source.permissionKey}-${source.grantedByType}-${source.systemClientId ?? "realm"}-${source.role}`}
          >
            <strong>{source.permissionKey}</strong>
            <span>{source.grantedByType}</span>
            <span>{source.systemName || source.systemClientId || "realm"}</span>
            <Tag type={source.grantedByType === "realmRole" ? "blue" : "cyan"}>{source.role}</Tag>
          </div>
        ))}
        {access.grantSources.length === 0 && <small>No permission grants resolved.</small>}
      </section>
    </div>
  );
}

function TagList({ values, type }: { values: string[]; type: "blue" | "cyan" | "gray" | "green" }) {
  if (values.length === 0) {
    return <small>None</small>;
  }
  return (
    <div className="rbac-effective-tags">
      {values.map((value) => (
        <Tag key={value} type={type}>
          {value}
        </Tag>
      ))}
    </div>
  );
}
