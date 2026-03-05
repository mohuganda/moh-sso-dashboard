import { useEffect, useState, useMemo } from "react";
import {
  TextInput,
  Button,
  Form,
  Stack,
  Tile,
  InlineLoading,
  Tag,
  Grid,
  Column,
  Loading,
  ToastNotification,
  ActionableNotification,
} from "@carbon/react";
import { Save, Undo } from "@carbon/icons-react";

import { useMeQuery, useUpdateProfileMutation } from "../../../../store/api/auth.api";

export default function MyProfilePage() {
  const { data: user, isLoading, isError, refetch } = useMeQuery();
  const [updateProfile, { isLoading: isSaving }] = useUpdateProfileMutation();

  const [form, setForm] = useState({ firstName: "", lastName: "" });
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");

  // Initial data sync
  useEffect(() => {
    if (user) {
      setForm({
        firstName: user.firstName || "",
        lastName: user.lastName || "",
      });
    }
  }, [user]);

  // Derived State: Check if the form has changed compared to server data
  const isDirty = useMemo(() => {
    if (!user) return false;
    return form.firstName !== (user.firstName || "") || form.lastName !== (user.lastName || "");
  }, [form, user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value } = e.target;
    setForm((prev) => ({ ...prev, [id]: value }));
  };

  const handleReset = () => {
    if (user) {
      setForm({
        firstName: user.firstName || "",
        lastName: user.lastName || "",
      });
      setStatus("idle");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isDirty) return;

    setStatus("idle");
    try {
      await updateProfile(form).unwrap();
      setStatus("success");
    } catch (err) {
      setStatus("error");
    }
  };

  if (isLoading) return <Loading description="Loading profile..." withOverlay={true} />;

  if (isError) {
    return (
      <ActionableNotification
        kind="error"
        title="Connection Error"
        subtitle="Could not load profile data."
        actionButtonLabel="Retry"
        onActionButtonClick={() => refetch()}
      />
    );
  }

  return (
    <Stack gap={7} className="profile-page-container">
      {status === "success" && (
        <ToastNotification
          kind="success"
          title="Profile updated"
          subtitle="Your changes have been saved."
          onClose={() => setStatus("idle")}
          timeout={3000}
          style={{ position: "fixed", top: "1rem", right: "1rem", zIndex: 9999 }}
        />
      )}

      <Grid narrow>
        <Column lg={16}>
          <h2 style={{ marginBottom: "1.5rem" }}>My Profile</h2>
        </Column>

        {/* Read Only Account Info */}
        <Column lg={8} md={4}>
          <Tile>
            <h5 style={{ marginBottom: "1.5rem" }}>Account Information</h5>
            <Stack gap={6}>
              <TextInput id="username" labelText="Username" value={user?.username || ""} readOnly />
              <TextInput id="email" labelText="Email" value={user?.email || ""} readOnly />
              <TextInput
                id="fullName"
                labelText="Full Name"
                value={`${user?.firstName || ""} ${user?.lastName || ""}`}
                readOnly
              />
            </Stack>
          </Tile>
        </Column>

        {/* Editable Profile */}
        <Column lg={8} md={4}>
          <Tile>
            <h5 style={{ marginBottom: "1.5rem" }}>Update Details</h5>
            <Form onSubmit={handleSubmit}>
              <Stack gap={6}>
                <TextInput
                  id="firstName"
                  labelText="First Name"
                  value={form.firstName}
                  onChange={handleInputChange}
                />
                <TextInput
                  id="lastName"
                  labelText="Last Name"
                  value={form.lastName}
                  onChange={handleInputChange}
                />

                <div style={{ display: "flex", gap: "1rem", marginTop: "1rem" }}>
                  <Button
                    type="submit"
                    disabled={isSaving || !isDirty}
                    renderIcon={isSaving ? () => <InlineLoading /> : Save}
                  >
                    {isSaving ? "Saving..." : "Save Changes"}
                  </Button>

                  {isDirty && !isSaving && (
                    <Button kind="ghost" onClick={handleReset} renderIcon={Undo}>
                      Reset
                    </Button>
                  )}
                </div>
              </Stack>
            </Form>
          </Tile>
        </Column>

        {/* Permission Visualization */}
        <Column lg={16} style={{ marginTop: "1.5rem" }}>
          <Tile>
            <h5 style={{ marginBottom: "1.25rem" }}>Roles & Permissions</h5>
            <Grid condensed>
              <Column lg={4} md={4} sm={4}>
                <Stack gap={3}>
                  <p style={{ fontSize: "0.75rem", fontWeight: "bold" }}>ACCOUNT STATUS</p>
                  <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
                    <Tag type={user?.enabled ? "green" : "red"}>
                      {user?.enabled ? "Active" : "Disabled"}
                    </Tag>
                    {user?.emailVerified && <Tag type="blue">Verified</Tag>}
                  </div>
                </Stack>
              </Column>

              <Column lg={12} md={4} sm={4}>
                <Stack gap={5}>
                  <section>
                    <p style={{ fontSize: "0.75rem", fontWeight: "bold", marginBottom: "0.5rem" }}>
                      ROLES
                    </p>
                    <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                      {user?.realmRoles?.map((role) => (
                        <Tag key={role} type="magenta" size="sm">
                          {role}
                        </Tag>
                      ))}
                    </div>
                  </section>

                  {user?.clientRoles &&
                    Object.entries(user.clientRoles).map(([client, roles]) => (
                      <div key={client}>
                        <p
                          style={{
                            fontSize: "0.875rem",
                            color: "#525252",
                            marginBottom: "0.25rem",
                          }}
                        >
                          {client}
                        </p>
                        <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                          {roles.map((role) => (
                            <Tag key={role} type="cyan" size="sm">
                              {role}
                            </Tag>
                          ))}
                        </div>
                      </div>
                    ))}
                </Stack>
              </Column>
            </Grid>
          </Tile>
        </Column>
      </Grid>
    </Stack>
  );
}
