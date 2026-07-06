import { Component, type ReactNode } from "react";
import { Tab, TabList, TabPanel, TabPanels, Tabs } from "@carbon/react";

import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";

import { TemplatesTab } from "../../components/TemplatesTab";
import { DocumentsTab } from "../../components/DocumentsTab";

type PageErrorBoundaryState = {
  error: Error | null;
};

class PageErrorBoundary extends Component<{ children: ReactNode }, PageErrorBoundaryState> {
  state: PageErrorBoundaryState = {
    error: null,
  };

  static getDerivedStateFromError(error: Error): PageErrorBoundaryState {
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

          <pre
            style={{
              marginTop: "0.5rem",
              fontSize: "0.8rem",
              whiteSpace: "pre-wrap",
            }}
          >
            {this.state.error.message}
            {"\n\n"}
            {this.state.error.stack}
          </pre>
        </div>
      );
    }

    return this.props.children;
  }
}

function DocumentPageContent() {
  return (
    <PageErrorBoundary>
      <div>
        <div style={{ marginBottom: "1.5rem" }}>
          <h2
            style={{
              margin: 0,
              marginBottom: "0.25rem",
            }}
          >
            Document Management
          </h2>

          <p
            style={{
              margin: 0,
              color: "#6f6f6f",
            }}
          >
            Manage reusable upload templates and monitor uploaded documents.
          </p>
        </div>

        <Tabs>
          <TabList aria-label="Document management tabs" contained>
            <PermissionGuard permission={PERMISSIONS.documentTemplatesRead}>
              <Tab>Templates</Tab>
            </PermissionGuard>
            <PermissionGuard permission={PERMISSIONS.documentsRead}>
              <Tab>Template Data Upload</Tab>
            </PermissionGuard>
          </TabList>

          <TabPanels>
            <PermissionGuard permission={PERMISSIONS.documentTemplatesRead}>
              <TabPanel
                style={{
                  paddingInline: 0,
                  paddingTop: "1.25rem",
                }}
              >
                <PageErrorBoundary>
                  <TemplatesTab />
                </PageErrorBoundary>
              </TabPanel>
            </PermissionGuard>

            <PermissionGuard permission={PERMISSIONS.documentsRead}>
              <TabPanel
                style={{
                  paddingInline: 0,
                  paddingTop: "1.25rem",
                }}
              >
                <PageErrorBoundary>
                  <DocumentsTab />
                </PageErrorBoundary>
              </TabPanel>
            </PermissionGuard>
          </TabPanels>
        </Tabs>
      </div>
    </PageErrorBoundary>
  );
}

export default function DocumentPage() {
  return (
    <PermissionGuard permission={PERMISSIONS.documentsRead}>
      <DocumentPageContent />
    </PermissionGuard>
  );
}
