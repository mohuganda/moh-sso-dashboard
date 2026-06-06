import { useEffect, useState, useMemo } from "react";
import { TextInput, Button, Form, Stack, Tile, Tag, InlineLoading } from "@carbon/react";
import { Save } from "@carbon/react/icons";

import { useMeQuery, useUpdateProfileMutation } from "@/store/api/auth.api";
import UserSessionsTable from "@/features/sessions/components/UserSessionsTable";

export default function MyProfilePage() {
  const { data: user, isLoading } = useMeQuery();
  const [updateProfile, { isLoading: isSaving }] = useUpdateProfileMutation();

  const [form, setForm] = useState({ firstName: "", lastName: "" });

  useEffect(() => {
    if (user) {
      setForm({
        firstName: user.firstName || "",
        lastName: user.lastName || "",
      });
    }
  }, [user]);

  const isDirty = useMemo(
    () => form.firstName !== (user?.firstName || "") || form.lastName !== (user?.lastName || ""),
    [form, user],
  );

  const formatDate = (date?: string) =>
    date
      ? new Date(date).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })
      : "—";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateProfile(form).unwrap();
    } catch (err) {
      console.error(err);
    }
  };

  if (isLoading) return <InlineLoading />;

  const initials =
    user?.firstName && user?.lastName
      ? `${user.firstName[0]}${user.lastName[0]}`
      : user?.username?.slice(0, 2).toUpperCase();

  return (
    <Stack gap={5} style={{ paddingBottom: "2rem" }}>
      {/* PROFILE HEADER */}
      <Tile style={{ padding: "1rem" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: "50%",
              background: "#0f62fe",
              color: "white",
              fontWeight: 600,
              fontSize: 20,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            {initials}
          </div>

          <Stack gap={1}>
            <h4 style={{ lineHeight: 1 }}>{user?.fullName || user?.username}</h4>
            <div style={{ fontSize: 14 }}>{user?.email}</div>
            <div style={{ fontSize: 12, opacity: 0.6 }}>
              Last login: {formatDate(user?.lastLoginAt)}
            </div>
          </Stack>
        </div>
      </Tile>

      {/* DASHBOARD SUMMARY */}
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
        <Tile style={{ padding: "12px", flex: "1 1 250px" }}>
          <p style={{ fontWeight: 600, fontSize: "0.875rem", marginBottom: "8px" }}>
            Activity Summary
          </p>
          <Stack gap={2} style={{ fontSize: "0.875rem" }}>
            <div>Created: {formatDate(user?.createdAt)}</div>
            <div>
              Total Roles:{" "}
              <Tag size="sm" type="cool-gray" style={{ margin: 0 }}>
                {user?.realmRoles?.length || 0}
              </Tag>
            </div>
          </Stack>
        </Tile>

        <Tile style={{ padding: "12px", flex: "1 1 250px" }}>
          <p style={{ fontWeight: 600, fontSize: "0.875rem", marginBottom: "8px" }}>
            Security Alerts
          </p>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
            {!user?.emailVerified && (
              <Tag size="sm" type="red">
                Email Unverified
              </Tag>
            )}
            {user?.requirePwdChange && (
              <Tag size="sm" type="purple">
                PW Change Due
              </Tag>
            )}
            {user?.enabled && user?.emailVerified && !user?.requirePwdChange && (
              <Tag size="sm" type="green">
                Secure
              </Tag>
            )}
          </div>
        </Tile>

        <Tile style={{ padding: "12px", flex: "1 1 250px" }}>
          <p style={{ fontWeight: 600, fontSize: "0.875rem", marginBottom: "8px" }}>Session</p>
          <Stack gap={3}>
            <div style={{ fontSize: "0.875rem" }}>Active: Current Browser</div>
            <Button size="sm" kind="ghost" style={{ padding: 0, minHeight: "unset" }}>
              Manage Sessions
            </Button>
          </Stack>
        </Tile>
      </div>

      {/* ACCOUNT + PROFILE */}
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
        <Tile style={{ padding: "1rem", flex: "1 1 320px" }}>
          <h5 style={{ marginBottom: "1rem" }}>Account Details</h5>
          <Stack gap={4}>
            <TextInput
              id="username"
              labelText="Username"
              value={user?.username || ""}
              readOnly
              size="sm"
            />
            <TextInput id="email" labelText="Email" value={user?.email || ""} readOnly size="sm" />
            <TextInput
              id="fullName"
              labelText="Full Name"
              value={user?.fullName || ""}
              readOnly
              size="sm"
            />
          </Stack>
        </Tile>

        <Tile style={{ padding: "1rem", flex: "1 1 320px" }}>
          <h5 style={{ marginBottom: "1rem" }}>Update Profile</h5>
          <Form onSubmit={handleSubmit}>
            <Stack gap={4}>
              <TextInput
                id="firstName"
                labelText="First Name"
                value={form.firstName}
                size="sm"
                onChange={(e) => setForm({ ...form, firstName: e.target.value })}
              />

              <TextInput
                id="lastName"
                labelText="Last Name"
                value={form.lastName}
                size="sm"
                onChange={(e) => setForm({ ...form, lastName: e.target.value })}
              />

              <Button
                size="sm"
                type="submit"
                disabled={isSaving || !isDirty}
                renderIcon={isSaving ? undefined : Save}
              >
                {isSaving ? "Saving..." : "Save Changes"}
              </Button>
            </Stack>
          </Form>
        </Tile>
      </div>

      {/* ROLES */}
      <Tile style={{ padding: "1rem" }}>
        <h5 style={{ marginBottom: "1rem" }}>Roles & Access</h5>

        <div style={{ display: "flex", gap: "2rem", flexWrap: "wrap" }}>
          <div style={{ minWidth: 200 }}>
            <p style={{ fontSize: "12px", fontWeight: "bold", color: "#525252" }}>REALM ROLES</p>
            <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginTop: 8 }}>
              {user?.realmRoles?.map((role) => (
                <Tag key={role} size="sm">
                  {role}
                </Tag>
              ))}
            </div>
          </div>

          <div style={{ flex: 1 }}>
            <p style={{ fontSize: "12px", fontWeight: "bold", color: "#525252" }}>CLIENT ROLES</p>
            {user?.clientRoles &&
              Object.entries(user.clientRoles).map(([client, roles]) => (
                <div key={client} style={{ marginTop: 8 }}>
                  <span style={{ fontSize: 12 }}>{client}:</span>
                  <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginTop: 4 }}>
                    {(roles as string[]).map((role) => (
                      <Tag key={role} type="cyan" size="sm">
                        {role}
                      </Tag>
                    ))}
                  </div>
                </div>
              ))}
          </div>
        </div>
      </Tile>

      {/* SESSIONS */}
      <Tile>
        <UserSessionsTable />
      </Tile>
    </Stack>
  );
}
