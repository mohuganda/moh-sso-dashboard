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
import { useEffect, useMemo, useState } from "react";

import {
  useCreateUserMutation,
  useUpdateUserMutation,
} from "../api";
import { useGetAssignableUserAccessQuery } from "@moh-sso/rbac";
import type { User } from "@moh-sso/types";
import { FormInlineAlert, useToast } from "@moh-sso/ui";

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

const createFormState = (initialUser?: User): UserFormState => ({
  username: initialUser?.username ?? "",
  email: initialUser?.email ?? "",
  firstName: initialUser?.firstName ?? "",
  lastName: initialUser?.lastName ?? "",
  realmRoles: initialUser?.realmRoles ?? [],
  enabled: initialUser?.enabled ?? initialUser?.isActive ?? true,
  emailVerified: initialUser?.emailVerified ?? true,
});

type Props = {
  mode: UserFormMode;
  initialUser?: User;
  onSuccess?: () => void;
};

export function UserFormPanel({ mode, initialUser, onSuccess }: Props) {
  const toast = useToast();

  const [form, setForm] = useState<UserFormState>(() => createFormState(initialUser));

  const [error, setError] = useState<string | null>(null);

  const [createUser, { isLoading: creating }] = useCreateUserMutation();
  const [updateUser, { isLoading: updating }] = useUpdateUserMutation();
  const { data: assignableAccess, isLoading: loadingAssignableAccess } =
    useGetAssignableUserAccessQuery();

  const submitting = creating || updating;

  const realmRoleItems = useMemo(
    () =>
      (assignableAccess?.realmRoles ?? []).map((role) => ({
        id: role.name,
        text: role.displayName || role.name,
      })),
    [assignableAccess?.realmRoles],
  );

  useEffect(() => {
    setForm(createFormState(initialUser));
    setError(null);
  }, [initialUser, mode]);

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
    } catch (err) {
      const message =
        getApiErrorMessage(err) ??
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
              onChange={(e) => {
                handleChange("username", e.target.value.toLowerCase());
              }}
            />

            <TextInput
              id="email"
              labelText="Email"
              type="email"
              required
              value={form.email}
              invalid={Boolean(form.email) && !isValidEmail(form.email)}
              invalidText="Enter a valid email address"
              onChange={(e) => {
                handleChange("email", e.target.value);
              }}
            />

            <TextInput
              id="firstName"
              labelText="First name"
              required
              value={form.firstName}
              onChange={(e) => {
                handleChange("firstName", e.target.value);
              }}
            />

            <TextInput
              id="lastName"
              labelText="Last name"
              required
              value={form.lastName}
              onChange={(e) => {
                handleChange("lastName", e.target.value);
              }}
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
              label={loadingAssignableAccess ? "Loading realm roles..." : "Realm roles"}
              items={realmRoleItems}
              itemToString={(item) => item?.text ?? ""}
              selectedItems={realmRoleItems.filter((r) => form.realmRoles.includes(r.id))}
              disabled={loadingAssignableAccess}
              onChange={({ selectedItems }) => {
                handleChange(
                  "realmRoles",
                  (selectedItems ?? []).map((r) => r.id),
                );
              }}
            />

            <Checkbox
              id="enabled"
              labelText="User enabled"
              checked={form.enabled}
              onChange={(_, { checked }) => {
                handleChange("enabled", checked);
              }}
            />

            <Checkbox
              id="emailVerified"
              labelText="Email verified"
              checked={form.emailVerified}
              onChange={(_, { checked }) => {
                handleChange("emailVerified", checked);
              }}
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

function getApiErrorMessage(error: unknown): string | undefined {
  if (!error || typeof error !== "object") {
    return undefined;
  }

  const maybeError = error as {
    data?: {
      message?: unknown;
      error?: {
        message?: unknown;
      };
    };
    error?: unknown;
  };

  if (typeof maybeError.data?.message === "string") {
    return maybeError.data.message;
  }

  if (typeof maybeError.data?.error?.message === "string") {
    return maybeError.data.error.message;
  }

  if (typeof maybeError.error === "string") {
    return maybeError.error;
  }

  return undefined;
}
