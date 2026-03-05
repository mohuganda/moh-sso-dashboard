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
} from "@carbon/react";
import { Save } from "@carbon/icons-react";

import { useMeQuery, useUpdateProfileMutation } from "../../../../store/api/auth.api";

export default function MyProfilePage() {
  const { data: user, isLoading, isFetching } = useMeQuery();
  const [updateProfile, { isLoading: isSaving }] = useUpdateProfileMutation();

  const [form, setForm] = useState({ firstName: "", lastName: "" });
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");

  // Sync form state when user data arrives
  useEffect(() => {
    if (user) {
      setForm({
        firstName: user.firstName || "",
        lastName: user.lastName || "",
      });
    }
  }, [user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value } = e.target;
    setForm((prev) => ({ ...prev, [id]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatus("idle");
    try {
      await updateProfile(form).unwrap();
      setStatus("success");
    } catch (err) {
      setStatus("error");
    }
  };

  const formatDate = (date?: string) =>
    date
      ? new Date(date).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })
      : "N/A";

  if (isLoading) return <Loading description="Loading profile..." withOverlay={true} />;

  return (
    <Stack gap={7} className="profile-page-container">
      {/* Notifications */}
      {status === "success" && (
        <ToastNotification
          kind="success"
          title="Profile updated"
          subtitle="Your changes have been saved successfully."
          onClose={() => setStatus("idle")}
          timeout={3000}
        />
      )}

      <Grid narrow>
        <Column lg={16} md={8} sm={4}>
          <h2 style={{ marginBottom: "1.5rem" }}>My Profile</h2>
        </Column>

        {/* Account Info (Read Only) */}
        <Column lg={8} md={4} sm={4}>
          <Tile>
            <h5 style={{ marginBottom: "1.5rem" }}>Account Information</h5>
            <Stack gap={6}>
              <TextInput id="username" labelText="Username" value={user?.username || ""} readOnly />
              <TextInput id="email" labelText="Email" value={user?.email || ""} readOnly />
              <TextInput
                id="createdAt"
                labelText="Account Created"
                value={formatDate(user?.createdAt)}
                readOnly
              />
            </Stack>
          </Tile>
        </Column>

        {/* Profile Details (Editable) */}
        <Column lg={8} md={4} sm={4}>
          <Tile>
            <h5 style={{ marginBottom: "1.5rem" }}>Update Profile</h5>
            <Form onSubmit={handleSubmit}>
              <Stack gap={6}>
                <TextInput
                  id="firstName"
                  labelText="First Name"
                  value={form.firstName}
                  onChange={handleInputChange}
                  placeholder="Enter first name"
                />
                <TextInput
                  id="lastName"
                  labelText="Last Name"
                  value={form.lastName}
                  onChange={handleInputChange}
                  placeholder="Enter last name"
                />
                <div style={{ marginTop: "1rem" }}>
                  <Button
                    type="submit"
                    disabled={isSaving || isFetching}
                    renderIcon={isSaving ? () => <InlineLoading /> : Save}
                  >
                    {isSaving ? "Saving..." : "Save Changes"}
                  </Button>
                </div>
              </Stack>
            </Form>
          </Tile>
        </Column>

        {/* Roles & Access */}
        <Column lg={16} md={8} sm={4} style={{ marginTop: "1.5rem" }}>
          <Tile>
            <h5 style={{ marginBottom: "1.25rem" }}>Roles & Access</h5>
            <Stack gap={6}>
              <section>
                <p style={{ fontSize: "0.75rem", fontWeight: "bold", marginBottom: "0.5rem" }}>
                  STATUS
                </p>
                <div style={{ display: "flex", gap: "0.5rem" }}>
                  <Tag type={user?.enabled ? "green" : "red"}>
                    {user?.enabled ? "Active" : "Disabled"}
                  </Tag>
                  <Tag type={user?.emailVerified ? "blue" : "outline"}>
                    {user?.emailVerified ? "Verified" : "Unverified"}
                  </Tag>
                </div>
              </section>

              <section>
                <p style={{ fontSize: "0.75rem", fontWeight: "bold", marginBottom: "0.5rem" }}>
                  REALM ROLES
                </p>
                <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                  {user?.realmRoles?.map((role) => (
                    <Tag key={role} type="magenta">
                      {role}
                    </Tag>
                  ))}
                </div>
              </section>

              {user?.clientRoles && (
                <section>
                  <p style={{ fontSize: "0.75rem", fontWeight: "bold", marginBottom: "0.5rem" }}>
                    CLIENT PERMISSIONS
                  </p>
                  {Object.entries(user.clientRoles).map(([client, roles]) => (
                    <div key={client} style={{ marginBottom: "1rem" }}>
                      <span style={{ fontSize: "0.875rem", color: "#525252" }}>{client}</span>
                      <div
                        style={{
                          marginTop: "0.25rem",
                          display: "flex",
                          gap: "0.5rem",
                          flexWrap: "wrap",
                        }}
                      >
                        {roles.map((role) => (
                          <Tag key={role} type="cyan">
                            {role}
                          </Tag>
                        ))}
                      </div>
                    </div>
                  ))}
                </section>
              )}
            </Stack>
          </Tile>
        </Column>
      </Grid>
    </Stack>
  );
}
