import { Button, Tile } from "@carbon/react";
import { Link } from "react-router-dom";

export function ForbiddenPage() {
  return (
    <main className="public-page public-page--centered">
      <Tile>
        <h1>Access denied</h1>
        <p>You do not have permission to open this area.</p>
        <Button as={Link} to="/apps/news" kind="primary">
          Back to portal
        </Button>
      </Tile>
    </main>
  );
}
