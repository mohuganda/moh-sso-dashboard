import { Button, Tile, Stack } from "@carbon/react";

export default function SecurityPage() {
  return (
    <Tile>
      <h3>Security</h3>

      <Stack gap={5}>
        <Button kind="primary" href="https://portal.local/realms/moh-realm/account" target="_blank">
          Manage Account (Password, MFA)
        </Button>
      </Stack>
    </Tile>
  );
}
