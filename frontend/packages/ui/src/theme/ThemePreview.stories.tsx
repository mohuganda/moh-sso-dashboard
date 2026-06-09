import { Button, Tag, Tile } from "@carbon/react";
import MohLogo from "../assets/images/moh-logo.png";

export default {
  title: "Theme/MOH Theme",
};

export const Overview = () => {
  return (
    <div style={{ display: "grid", gap: "1rem", maxWidth: "42rem" }}>
      <Tile>
        <img
          src={MohLogo}
          alt="Ministry of Health Uganda"
          style={{ maxHeight: 56, marginBottom: 16 }}
        />

        <h3 style={{ margin: 0 }}>Integrated Health Portal</h3>
        <p style={{ color: "var(--cds-text-secondary)" }}>
          Shared theme tokens, Carbon overrides, and branding assets.
        </p>

        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <Tag type="blue">Info</Tag>
          <Tag type="green">Success</Tag>
          <Tag type="magenta">Warning</Tag>
          <Tag type="red">Critical</Tag>
        </div>

        <div style={{ marginTop: 16 }}>
          <Button>Primary action</Button>
        </div>
      </Tile>
    </div>
  );
};
