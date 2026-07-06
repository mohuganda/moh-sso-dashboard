import { Component, type ReactNode } from "react";
import { Tab, TabList, TabPanel, TabPanels, Tabs } from "@carbon/react";

import { TemplatesTab } from "../../components/TemplatesTab";
import { DocumentsTab } from "../../components/DocumentsTab";

class PageErrorBoundary extends Component<{ children: ReactNode }, { error: Error | null }> {
  state = { error: null };
  static getDerivedStateFromError(error: Error) {
    return { error };
  }
  render() {
    if (this.state.error) {
      return (
        <div
          style={{
            padding: "2rem",
            background: "#fff1f1",
            border: "1px solid #ffd7d9",
            borderRadius: 4,
          }}
        >
          <strong>Page crashed:</strong>
          <pre style={{ marginTop: "0.5rem", fontSize: "0.8rem", whiteSpace: "pre-wrap" }}>
            {(this.state.error as Error).message}
            {"\n\n"}
            {(this.state.error as Error).stack}
          </pre>
        </div>
      );
    }
    return this.props.children;
  }
}

export default function DocumentPage() {
  return (
    <PageErrorBoundary>
      <div>
        {/* Page heading */}
        <div style={{ marginBottom: "1.5rem" }}>
          <h2 style={{ margin: 0, marginBottom: "0.25rem" }}>Document Management</h2>
          <p style={{ margin: 0, color: "#6f6f6f" }}>
            Manage reusable upload templates and monitor uploaded documents.
          </p>
        </div>

        {/* Tabs */}
        <Tabs>
          <TabList aria-label="Document management tabs" contained>
            <Tab>Templates</Tab>
            <Tab>Template Data Upload</Tab>
          </TabList>

          <TabPanels>
            <TabPanel style={{ paddingInline: 0, paddingTop: "1.25rem" }}>
              <PageErrorBoundary>
                <TemplatesTab />
              </PageErrorBoundary>
            </TabPanel>

            <TabPanel style={{ paddingInline: 0, paddingTop: "1.25rem" }}>
              <PageErrorBoundary>
                <DocumentsTab />
              </PageErrorBoundary>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </PageErrorBoundary>
  );
}
