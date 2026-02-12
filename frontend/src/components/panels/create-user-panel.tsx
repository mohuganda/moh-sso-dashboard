// components/users/panels/UserFormPanel.tsx
import {
  Stack,
  TextInput,
  Button,
  Checkbox,
  InlineLoading,
  MultiSelect,
  Form,
  FormGroup,
} from "@carbon/react";
import { useMemo, useState } from "react";

import { useCreateUserMutation, useUpdateUserMutation } from "../../store/api/users.api";
import type { User } from "../../store/types/user.types";
import { useToast } from "../notifications/toast/useToast";
import { FormInlineAlert } from "../notifications/in-line-alerts/FormInlineAlert";

export type UserFormMode = "create" | "edit";

/**
 * UI form state (NOT API payload)
 */
type UserFormState = {
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  realmRoles: string[];
  enabled: boolean;
  emailVerified: boolean;
};

/**
 * Temporary realm roles
 * (replace with API-driven roles later)
 */
const REALM_ROLES = [
  { id: "admin", text: "Admin" },
  { id: "manager", text: "Manager" },
  { id: "user", text: "User" },
];

type Props = {
  mode: UserFormMode;
  initialUser?: User;
  onSuccess?: () => void;
};

export function UserFormPanel({ mode, initialUser, onSuccess }: Props) {
  const toast = useToast();

  const [form, setForm] = useState<UserFormState>({
    username: initialUser?.username ?? "",
    email: initialUser?.email ?? "",
    firstName: initialUser?.firstName ?? "",
    lastName: initialUser?.lastName ?? "",
    realmRoles: initialUser?.realmRoles ?? [],
    enabled: initialUser?.enabled ?? true,
    emailVerified: initialUser?.emailVerified ?? true,
  });

  const [error, setError] = useState<string | null>(null);

  const [createUser, { isLoading: creating }] = useCreateUserMutation();
  const [updateUser, { isLoading: updating }] = useUpdateUserMutation();

  const submitting = creating || updating;

  /* -----------------------------
   * Helpers
   * ----------------------------- */
  const isValidEmail = (email: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

  const buildPayload = () => ({
    email: form.email.trim().toLowerCase(),
    firstName: form.firstName.trim(),
    lastName: form.lastName.trim(),
    realmRoles: form.realmRoles,
    enabled: form.enabled,
    emailVerified: form.emailVerified,
  });

  /* -----------------------------
   * Validation
   * ----------------------------- */
  const isValid = useMemo(() => {
    if (!form.username.trim()) return false;
    if (!form.firstName.trim()) return false;
    if (!form.lastName.trim()) return false;
    if (!isValidEmail(form.email)) return false;
    return true;
  }, [form]);

  const handleChange = <K extends keyof UserFormState>(field: K, value: UserFormState[K]) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  /* -----------------------------
   * Submit
   * ----------------------------- */
  const handleSubmit = async () => {
    if (!isValid || submitting) return;

    setError(null);

    try {
      if (mode === "create") {
        await createUser({
          username: form.username.trim().toLowerCase(),
          ...buildPayload(),
        }).unwrap();

        toast.success("User created", `User "${form.username}" was created successfully`);
      } else if (initialUser?.id) {
        await updateUser({
          id: initialUser.id,
          data: buildPayload(),
        }).unwrap();

        toast.success("User updated", `Changes to "${form.username}" were saved`);
      }

      onSuccess?.();
    } catch (err: any) {
      const message =
        err?.data?.message ??
        (mode === "create" ? "Failed to create user" : "Failed to update user");

      setError(message);
      toast.error("Operation failed", "Please review the form and try again");
    }
  };

  return (
    <Form>
      <Stack gap={6}>
        {error && <FormInlineAlert title="Unable to save user" subtitle={error} />}

        {/* -----------------------------
         * User details
         * ----------------------------- */}
        <FormGroup legendText="User details">
          <Stack gap={4}>
            <TextInput
              id="username"
              labelText="Username"
              required
              disabled={mode === "edit"}
              helperText={mode === "edit" ? "Username cannot be changed" : undefined}
              value={form.username}
              onChange={(e) => handleChange("username", e.target.value.toLowerCase())}
            />

            <TextInput
              id="email"
              labelText="Email"
              type="email"
              required
              value={form.email}
              invalid={Boolean(form.email) && !isValidEmail(form.email)}
              invalidText="Enter a valid email address"
              onChange={(e) => handleChange("email", e.target.value)}
            />

            <TextInput
              id="firstName"
              labelText="First name"
              required
              value={form.firstName}
              onChange={(e) => handleChange("firstName", e.target.value)}
            />

            <TextInput
              id="lastName"
              labelText="Last name"
              required
              value={form.lastName}
              onChange={(e) => handleChange("lastName", e.target.value)}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Access & Roles
         * ----------------------------- */}
        <FormGroup legendText="Access control">
          <Stack gap={4}>
            <MultiSelect
              id="realmRoles"
              titleText="Realm roles"
              label="Realm roles"
              items={REALM_ROLES}
              itemToString={(item) => item?.text ?? ""}
              selectedItems={REALM_ROLES.filter((r) => form.realmRoles.includes(r.id))}
              onChange={({ selectedItems }) =>
                handleChange(
                  "realmRoles",
                  (selectedItems ?? []).map((r) => r.id),
                )
              }
            />

            <Checkbox
              id="enabled"
              labelText="User enabled"
              checked={form.enabled}
              onChange={(_, { checked }) => handleChange("enabled", checked)}
            />

            <Checkbox
              id="emailVerified"
              labelText="Email verified"
              checked={form.emailVerified}
              onChange={(_, { checked }) => handleChange("emailVerified", checked)}
            />
          </Stack>
        </FormGroup>

        {/* -----------------------------
         * Actions
         * ----------------------------- */}
        <Stack orientation="horizontal" gap={3}>
          <Button kind="primary" disabled={!isValid || submitting} onClick={handleSubmit}>
            {submitting ? (
              <InlineLoading
                description={mode === "create" ? "Creating user…" : "Saving changes…"}
              />
            ) : mode === "create" ? (
              "Create user"
            ) : (
              "Save changes"
            )}
          </Button>
        </Stack>
      </Stack>
    </Form>
  );
}
