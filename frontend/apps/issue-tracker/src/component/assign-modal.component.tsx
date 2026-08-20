import { useState } from "react";
import { ComboBox, Modal, TextArea, InlineNotification } from "@carbon/react";
import {
  useGetKeycloakGroupsQuery,
  useGetKeycloakGroupMembersQuery,
  useAssignIssuesMutation,
} from "../api";
import type { KeycloakGroup, KeycloakGroupMember } from "../types";

type AssignModalProps = {
  issueCodes: string[];
  onClose: () => void;
  onSuccess?: () => void;
};

export const AssignModal = ({ issueCodes, onClose, onSuccess }: AssignModalProps) => {
  const [selectedGroup, setSelectedGroup] = useState<KeycloakGroup | null>(null);
  const [selectedMember, setSelectedMember] = useState<KeycloakGroupMember | null>(null);
  const [comment, setComment] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const { data: groups = [], isLoading: isLoadingGroups, error: groupsError } = useGetKeycloakGroupsQuery();

  const selectedGroupId = selectedGroup?.id ?? "";

  const {
    data: members = [],
    isLoading: isLoadingMembers,
    error: membersError,
  } = useGetKeycloakGroupMembersQuery(selectedGroupId, {
    skip: !selectedGroupId,
  });

  const [assignIssues, { isLoading: isAssigning }] = useAssignIssuesMutation();

  const handleGroupChange = ({ selectedItem }: { selectedItem?: KeycloakGroup | null }) => {
    setSelectedGroup(selectedItem ?? null);
    setSelectedMember(null);
  };

  const handleMemberChange = ({ selectedItem }: { selectedItem?: KeycloakGroupMember | null }) => {
    setSelectedMember(selectedItem ?? null);
  };

  const handleSubmit = async () => {
    if (!selectedMember?.email) {
      setErrorMessage("Selected user does not have a valid email address.");
      return;
    }

    if (issueCodes.length === 0) {
      setErrorMessage("No issues selected for assignment.");
      return;
    }

    setErrorMessage("");

    try {
      await assignIssues({
        issue_codes: issueCodes,
        assigned_to: selectedMember.email,
        comment: comment.trim() || undefined,
      }).unwrap();

      onSuccess?.();
      onClose();
    } catch (err: any) {
      console.error("Failed to assign issue(s):", err);
      setErrorMessage(
        err?.data?.message ?? "Failed to assign issue(s). Please check Keycloak connection or try again.",
      );
    }
  };

  const isMultiple = issueCodes.length > 1;
  const heading = isMultiple ? `Assign ${issueCodes.length} Issues` : `Assign Issue [${issueCodes[0] ?? ""}]`;

  return (
    <Modal
      open
      modalHeading={heading}
      primaryButtonText={isAssigning ? "Assigning..." : "Assign Issue"}
      secondaryButtonText="Cancel"
      onRequestClose={onClose}
      onRequestSubmit={handleSubmit}
      primaryButtonDisabled={!selectedMember?.email || isAssigning}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: "1rem", marginTop: "1rem" }}>
        {errorMessage && (
          <InlineNotification
            kind="error"
            title="Assignment Error"
            subtitle={errorMessage}
            onCloseButtonClick={() => setErrorMessage("")}
          />
        )}

        <p style={{ color: "#525252" }}>
          Select a user group and choose a team member to assign {isMultiple ? "these issues" : "this issue"} to.
        </p>

        <ComboBox
          id="user-group-select"
          titleText="User Group"
          placeholder={isLoadingGroups ? "Loading user groups..." : "Select User Group"}
          items={groups}
          selectedItem={selectedGroup}
          itemToString={(item: KeycloakGroup | null) => (item ? `${item.name} (${item.path})` : "")}
          onChange={handleGroupChange}
          disabled={isLoadingGroups || isAssigning}
          warn={Boolean(groupsError)}
          warnText={groupsError ? "Failed to load User groups" : undefined}
        />

        <ComboBox
          id="keycloak-member-select"
          titleText="Assignee (User)"
          placeholder={
            !selectedGroupId
              ? "First select a user group"
              : isLoadingMembers
                ? "Loading group members..."
                : "Select User"
          }
          items={members}
          selectedItem={selectedMember}
          itemToString={(item: KeycloakGroupMember | null) => {
            if (!item) return "";
            const name = [item.firstName, item.lastName].filter(Boolean).join(" ");
            return name ? `${name} (${item.email || item.username})` : item.email || item.username;
          }}
          onChange={handleMemberChange}
          disabled={!selectedGroupId || isLoadingMembers || isAssigning}
          warn={Boolean(membersError)}
          warnText={membersError ? "Failed to load group members" : undefined}
        />

        {selectedMember && (
          <div
            style={{
              padding: "0.75rem 1rem",
              backgroundColor: "#f4f4f4",
              borderRadius: "4px",
              borderLeft: "4px solid #0f62fe",
              fontSize: "0.875rem",
            }}
          >
            <strong>Selected User Email:</strong> {selectedMember.email || "No email available"}
          </div>
        )}

        <TextArea
          id="assign-comment"
          labelText="Assignment Comment / Notes (Optional)"
          placeholder="Add optional instructions or comments for this assignment"
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          rows={3}
          disabled={isAssigning}
        />
      </div>
    </Modal>
  );
};
