import { Stack, TextInput, Button, InlineLoading } from "@carbon/react";
import { useState } from "react";

import { useCreateClientRoleMutation } from "@moh-sso/api";
import { useToast } from "@moh-sso/ui";

type Props = {
  clientId: string;
  onSuccess?: () => void;
};

export function CreateClientRoleForm({ clientId, onSuccess }: Props) {
  const toast = useToast();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const [createRole, { isLoading }] = useCreateClientRoleMutation();

  const isValid = Boolean(name.trim());

  const handleSubmit = async () => {
    if (!isValid) return;

    try {
      await createRole({
        clientId,
        payload: {
          role: name.trim(),
          description: description.trim() || undefined,
        },
      }).unwrap();

      toast.success("Role created", `Role "${name}" added`);
      setName("");
      setDescription("");
      onSuccess?.();
    } catch {
      toast.error("Failed to create role", `Please try again`);
    }
  };

  return (
    <Stack gap={3}>
      <TextInput
        id="role-name"
        labelText="Role name"
        helperText="Format: resource:action (e.g. users:view)"
        value={name}
        onChange={(e) => {
          setName(e.target.value);
        }}
      />

      <TextInput
        id="role-description"
        labelText="Description (optional)"
        value={description}
        onChange={(e) => {
          setDescription(e.target.value);
        }}
      />

      <Button kind="primary" size="sm" disabled={!isValid || isLoading} onClick={handleSubmit}>
        {isLoading ? <InlineLoading description="Creating…" /> : "Add role"}
      </Button>
    </Stack>
  );
}
