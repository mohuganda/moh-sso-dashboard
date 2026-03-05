import { Button, Tile, Stack, Section, Heading } from "@carbon/react";
import { Launch, Security } from "@carbon/icons-react";

export default function SecurityPage() {
  // Tip: In production, pull this from your config or env variables
  const KEYCLOAK_ACCOUNT_URL = "http://localhost:8081/realms/moh-realm/account";

  return (
    <Tile className="security-tile">
      <Stack gap={6}>
        <Section>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.75rem",
              marginBottom: "0.5rem",
            }}
          >
            <Security size={24} aria-label="Security icon" />
            <Heading style={{ fontSize: "1.25rem", fontWeight: "400" }}>Security Settings</Heading>
          </div>

          <p style={{ color: "#525252", maxWidth: "32rem", lineHeight: "1.5" }}>
            To ensure your data remains protected, sensitive actions like changing your password or
            managing Two-Factor Authentication (2FA) are handled through our centralized identity
            provider.
          </p>
        </Section>

        <hr style={{ border: "none", borderTop: "1px solid #e0e0e0", margin: "0" }} />

        <Section>
          <Stack gap={4}>
            <div>
              <p style={{ fontWeight: "600", marginBottom: "0.25rem" }}>
                Account Management Portal
              </p>
              <p style={{ fontSize: "0.875rem", color: "#525252" }}>
                Redirects to the secure MOH Identity Management console.
              </p>
            </div>

            <div style={{ marginTop: "0.5rem" }}>
              <Button
                kind="tertiary"
                href={KEYCLOAK_ACCOUNT_URL}
                target="_blank"
                rel="noopener noreferrer"
                renderIcon={Launch}
              >
                Manage Password & MFA
              </Button>
            </div>
          </Stack>
        </Section>
      </Stack>
    </Tile>
  );
}
