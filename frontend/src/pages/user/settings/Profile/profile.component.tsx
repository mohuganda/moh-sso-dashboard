import { TextInput, Button, Form, Stack, Tile } from "@carbon/react";

export default function MyProfilePage() {
  const { data: user } = useGetMeQuery();
  const [updateProfile, { isLoading }] = useUpdateMeMutation();

  const [form, setForm] = useState({
    firstName: user?.firstName || "",
    lastName: user?.lastName || "",
  });

  const handleSubmit = async () => {
    await updateProfile(form);
  };

  return (
    <Tile>
      <h3>My Profile</h3>

      <Form>
        <Stack gap={5}>
          <TextInput id="username" labelText="Username" value={user?.username} readOnly />

          <TextInput id="email" labelText="Email" value={user?.email} readOnly />

          <TextInput
            id="firstName"
            labelText="First Name"
            value={form.firstName}
            onChange={(e) => setForm({ ...form, firstName: e.target.value })}
          />

          <TextInput
            id="lastName"
            labelText="Last Name"
            value={form.lastName}
            onChange={(e) => setForm({ ...form, lastName: e.target.value })}
          />

          <Button onClick={handleSubmit} disabled={isLoading}>
            Save Changes
          </Button>
        </Stack>
      </Form>
    </Tile>
  );
}
