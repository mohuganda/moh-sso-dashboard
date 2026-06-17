import {Button, Modal, TextArea, Tile} from "@carbon/react";
import {ArrowLeft} from "@carbon/react/icons";
import "./issue-detail.scss";
import React, {useState} from "react";
import {useCreateTransactionMutation, useGetIssuesQuery, useGetTransactionsQuery} from "../../component/issuetracker.api.ts";
import {useSelector} from "react-redux";

import type {Issue} from "../issue-tracker.component.tsx";
import {IssueModal} from "../../component/issue-modal.component.tsx";
import {selectUser} from "@moh-sso/auth";
type ModalMode = 'resolve' | 'close' | 'comment' | null;

const IssueDetail = ({ selectedIssue:initialIssue, goToBack, showBack = true }: { selectedIssue: Issue, goToBack: () => void, showBack?: boolean }) => {
  const { data: latestIssues } = useGetIssuesQuery();
  const selectedIssue = latestIssues?.data?.find(
      (issueItem: Issue) => issueItem?.issue_code === initialIssue?.issue_code
  ) || initialIssue;
  const [comment, setComment] = useState("");
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<ModalMode>(null);
  const [isViewModalResolution, setIsViewModalResolution] = useState(false);
  const { data: transactions, isLoading: isLoadingTransactions } = useGetTransactionsQuery(
      selectedIssue?.issue_code,
      { skip: !selectedIssue?.issue_code }
  );
  const [createTransaction, { isLoading: isResolving }] = useCreateTransactionMutation();
  const user = useSelector(selectUser);

  const handleTextChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setComment(event.target.value);
  };

  const handleUpdateAction = async () => {
    if (!comment.trim()) return;

    const payload = {
      id: selectedIssue?.issue_code,
      body: {
        resolution_action: comment,
        resolved_by: user?.username,
        status: modalMode === 'resolve' ? 'RESOLVED' :
            modalMode === 'close' ? 'CLOSED' :
                selectedIssue?.status
      }
    };

    try {
      await createTransaction(payload).unwrap();

      setIsViewModalResolution(false);
      setModalMode(null);
      setComment("");
      console.log(`${modalMode} action successful!`);
    } catch (err) {
      console.error(`Failed to ${modalMode} issue:`, err);
    }
  };
  return (
      <>
        {showBack && (
            <div style={{ marginBottom: "1rem" }}>
              <Button
                  kind="tertiary"
                  size="lg"
                  renderIcon={ArrowLeft}
                  onClick={goToBack}
              >
                Back to Issues
              </Button>
            </div>
        )}
        <div className="issue-detail-container">
          <div style={{ width: '65%' }}>
            <Tile className="issue-tile">
              <div style={{ padding: '0.5rem' }}>
                <div className="issue-label">Reported Issue: {selectedIssue?.issue_code} </div>
                <p>
                                <span className="issue-p">
                                    <strong>Issue Type:</strong> {selectedIssue?.issue_type}
                                </span>

                  <span className="issue-p">
                                    <strong>Priority:</strong> {selectedIssue?.priority ?? "<<Not set>>"}
                                </span>
                  <strong>Severity:</strong> {selectedIssue?.severity ?? "<<Not set>>"}
                </p>
                <p className="issue-p-top"><strong>Organization Unit:</strong> {selectedIssue?.org_unit}</p>
                <p className="issue-p-top"><strong>Dataset:</strong> {selectedIssue?.dataset}</p>
                <p className="issue-p-top"><strong>DataElement:</strong> {selectedIssue?.data_element}</p>
                <p className="issue-p-top"><strong>Reporting Period:</strong> {selectedIssue?.time_Period}</p>
                <p className="issue-p-top"><strong>Description:</strong></p>
                <p> {selectedIssue?.issue}</p>
              </div>
            </Tile>
            <Tile className="issue-tile issue-history-container">
              <div>
                <div className="issue-label">Issue History / Transactions</div>


                {isLoadingTransactions ? (
                    <p>Loading history...</p>
                ) : (
                    <div className="transaction-list">
                      {(transactions?.data?.length ?? 0) > 0 ? (
                          transactions?.data?.map((trx, index) => (
                              <div key={index} className="transaction-item" style={{ marginBottom: '1rem', fontSize: '0.875rem' }}>
                                <div style={{ fontWeight: 'bold' }}>
                                  {trx?.resolution_date} - {trx?.resolved_by}
                                </div>
                                <div style={{ color: '#525252' }}>{trx?.resolution_action || "No comment provided"}</div>
                              </div>
                          ))
                      ) : (
                          <p style={{ fontStyle: 'italic', color: '#6f6f6f' }}>No transactions found for this issue.</p>
                      )}
                    </div>
                )}
              </div>
            </Tile>
          </div>
          <div className="issue-reporter-details" style={{ width: '30%' }}>
            <Tile className="issue-tile">
              <div  style={{ padding: '0.5rem'}}>
                <p className="issue-p-top"><strong>Issue Status:</strong> {selectedIssue?.status}</p>
                <p><strong>Reported By:</strong> {selectedIssue?.reported_by}</p>
                <p><strong>Date:</strong> {selectedIssue?.date_reported}</p>
              </div>
            </Tile>

            <Tile className="issue-tile issue-actions-container">
              {selectedIssue?.status !== 'CLOSED' && selectedIssue?.status !== 'RESOLVED' && (
                  <Button className="btn-issue" size="md" kind="primary" onClick={() => setIsEditModalOpen(true)}>
                    Edit Issue
                  </Button>
              )}
              <Button className="btn-issue btn-full-width" kind="secondary"
                      onClick={() => { setModalMode('comment'); setIsViewModalResolution(true); setComment(""); }}
              >
                Add Comment
              </Button>
              {selectedIssue?.status !== 'CLOSED' && selectedIssue?.status !== 'RESOLVED' && (
                  <Button className="btn-issue custom-btn-success btn-full-width"
                          onClick={() => { setModalMode('resolve'); setIsViewModalResolution(true); setComment(""); }}
                  >
                    Resolve Issue
                  </Button>
              )}
              {selectedIssue?.status !== 'CLOSED' && (
                  <Button className="btn-issue btn-full-width" kind="danger--tertiary"
                          onClick={() => { setModalMode('close'); setIsViewModalResolution(true); setComment(""); }}
                  >
                    Close Issue
                  </Button>
              )}
            </Tile>
          </div>
        </div>

        {isViewModalResolution && (
            <Modal
                open
                primaryButtonText={
                  isResolving ? "Processing..." :
                      modalMode === 'resolve' ? "Confirm Resolution" :
                          modalMode === 'close' ? "Confirm Closure" : "Add Comment"
                }
                secondaryButtonText="Cancel"
                modalHeading={
                  modalMode === 'resolve' ? "Resolve Issue" :
                      modalMode === 'close' ? "Close Issue" : "Add Comment"
                }
                onRequestClose={() => { setIsViewModalResolution(false); setModalMode(null); }}
                onRequestSubmit={handleUpdateAction}
                primaryButtonDisabled={!comment.trim() || isResolving}
            >
              <p style={{ marginBottom: '1rem' }}>
                {modalMode === 'resolve' && "Describe how this issue was resolved."}
                {modalMode === 'close' && "Provide a reason for closing this issue."}
                {modalMode === 'comment' && "Enter your comment below."}
              </p>
              <TextArea
                  data-modal-primary-focus
                  id="issue-action-comment"
                  labelText=""
                  value={comment}
                  onChange={handleTextChange}
                  rows={4}
              />
            </Modal>
        )}

        {isEditModalOpen && (
            <IssueModal
                onClose={() => setIsEditModalOpen(false)}
                selectedIssue={selectedIssue}
            />
        )}
      </>
  )
}

export default IssueDetail;