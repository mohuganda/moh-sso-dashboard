import { Tile, Grid, Column } from "@carbon/react";
import { useNavigate } from "react-router-dom";

const apps = [
  {
    id: "audit",
    name: "Audit Logs",
    description: "System security & activity",
    route: "/admin/audit-logs",
    adminOnly: true,
  },
];

export default function AppLauncherPage() {
  const navigate = useNavigate();

  return (
    <Grid condensed>
      {apps.map((app) => (
        <Column lg={4} md={4} sm={4} key={app.id}>
          <Tile style={{ cursor: "pointer" }} onClick={() => navigate(app.route)}>
            <h4>{app.name}</h4>
            <p>{app.description}</p>
          </Tile>
        </Column>
      ))}
    </Grid>
  );
}
