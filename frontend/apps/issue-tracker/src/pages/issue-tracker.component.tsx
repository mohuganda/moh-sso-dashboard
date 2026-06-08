import { useMemo, useState } from "react";
import { Button, InlineLoading, InlineNotification, Tile } from "@carbon/react";
import { Add } from "@carbon/react/icons";
import { DataList } from "@moh-sso/data-visualizer";

import "./issue-tracker.scss";

import { headers } from "../lib/constants";
import IssueDetail from "./issue-detail/issue-detail.component";
import type { Issue } from "@moh-sso/types";
import { IssueModal } from "../component/issue-modal.component";
import { useGetIssuesQuery } from "@moh-sso/api";

type IssueRow = Issue & {
  id: string;
};

type DataListRow = {
  id?: string;
  [key: string]: unknown;
};

const IssueTracker = () => {
  const [showModal, setShowModal] = useState(false);
  const [selectedIssue, setSelectedIssue] = useState<IssueRow | null>(null);

  const { data, isLoading, isFetching, isError, error, refetch } = useGetIssuesQuery();

  const issues = useMemo<IssueRow[]>(() => {
    return (data?.data ?? []).map((item: Issue) => ({
      ...item,
      id: String(item.issue_id),
    }));
  }, [data]);

  const handleOpenModal = () => {
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
  };

  const handleBackToList = () => {
    setSelectedIssue(null);
  };

  const handleIssueClick = (row: DataListRow) => {
    if (!row?.id) return;

    const issue = issues.find((item) => item.id === String(row.id));

    if (!issue) return;

    setSelectedIssue(issue);
  };

  if (selectedIssue) {
    return <IssueDetail selectedIssue={selectedIssue} goToBack={handleBackToList} />;
  }

  return (
    <>
      <div className="issue-tracker-page">
        <div className="dv-toolbar issue-label-container">
          <div>
            <h3 className="issue-label">Registered Issues</h3>
            <p className="issue-subtitle">
              Track reported data quality issues, priorities, severity, and resolution status.
            </p>
          </div>

          <div className="issue-toolbar-actions">
            {isFetching && !isLoading && <InlineLoading description="Refreshing issues…" />}

            <Button
              size="md"
              kind="primary"
              renderIcon={Add}
              className="dwh-btn-width"
              onClick={handleOpenModal}
            >
              New Issue
            </Button>
          </div>
        </div>

        {isLoading && (
          <Tile className="issue-container">
            <InlineLoading description="Loading issues…" />
          </Tile>
        )}

        {isError && (
          <Tile className="issue-container">
            <InlineNotification
              kind="error"
              lowContrast
              title="Failed to load issues"
              subtitle={
                (error as any)?.data?.message ||
                (error as any)?.error ||
                "An unexpected error occurred while fetching issues."
              }
            />

            <div style={{ marginTop: "1rem" }}>
              <Button kind="secondary" onClick={() => refetch()}>
                Retry
              </Button>
            </div>
          </Tile>
        )}

        {!isLoading && !isError && issues.length === 0 && (
          <Tile className="issue-container issue-empty-state">
            <h4>No issues found</h4>
            <p>There are currently no registered issues. Create a new issue to start tracking.</p>

            <Button kind="secondary" renderIcon={Add} onClick={handleOpenModal}>
              Create first issue
            </Button>
          </Tile>
        )}

        {!isLoading && !isError && issues.length > 0 && (
          <div className="issue-container">
            <DataList columns={headers} data={issues} handleIssueClick={handleIssueClick} />
          </div>
        )}
      </div>

      {showModal && <IssueModal onClose={handleCloseModal} />}
    </>
  );
};

export default IssueTracker;
