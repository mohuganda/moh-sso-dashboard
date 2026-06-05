package data_quality

type issueResponse struct {
	IssueID      int64   `json:"issue_id"`
	IssueCode    *string `json:"issue_code,omitempty"`
	Dataset      *string `json:"dataset,omitempty"`
	DataElement  *string `json:"data_element,omitempty"`
	OrgUnit      *string `json:"org_unit,omitempty"`
	Issue        *string `json:"issue,omitempty"`
	IssueType    *string `json:"issue_type,omitempty"`
	DateReported string  `json:"date_reported"`
	ReportedBy   *string `json:"reported_by,omitempty"`
	Status       *string `json:"status,omitempty"`
	Priority     *string `json:"priority,omitempty"`
	Severity     *string `json:"severity,omitempty"`
	UpdatedDate  *string `json:"updated_date,omitempty"`
	UpdatedBy    *string `json:"updated_by,omitempty"`
	TimePeriod   *string `json:"time_Period,omitempty"`
}

type createIssueRequest struct {
	Dataset     *string `json:"dataset"`
	DataElement *string `json:"data_element"`
	OrgUnit     *string `json:"org_unit"`
	Issue       string  `json:"issue" binding:"required"`
	IssueType   *string `json:"issue_type"`
	ReportedBy  *string `json:"reported_by"`
	TimePeriod  *string `json:"time_Period"`
}

type updateIssueRequest struct {
	Dataset      *string `json:"dataset"`
	DataElement  *string `json:"data_element"`
	OrgUnit      *string `json:"org_unit"`
	Issue        *string `json:"issue"`
	IssueType    *string `json:"issue_type"`
	DateReported *string `json:"date_reported"`
	ReportedBy   *string `json:"reported_by"`
	Status       *string `json:"status"`
	Priority     *string `json:"priority"`
	Severity     *string `json:"severity"`
	UpdatedBy    *string `json:"updated_by"`
	TimePeriod   *string `json:"time_Period"`
}

type createIssueResolutionRequest struct {
	Stage              *string `json:"stage"`
	Status             *string `json:"status"`
	ResolutionAction   *string `json:"resolution_action"`
	ResolvedBy         *string `json:"resolved_by"`
	ResolutionDate     *string `json:"resolution_date"`
	VerificationStatus *string `json:"verification_status"`
	VerifiedBy         *string `json:"verified_by"`
	VerificationDate   *string `json:"verification_date"`
	PreventiveAction   *string `json:"preventive_action"`
	ProcessChange      *string `json:"process_change"`
	PreventiveOwner    *string `json:"preventive_owner"`
	DueDate            *string `json:"due_date"`
}

type issueStageResponse struct {
	ID                 int64   `json:"id"`
	IssueCode          string  `json:"issue_code"`
	Status             *string `json:"status,omitempty"`
	Stage              *string `json:"stage,omitempty"`
	IsCurrent          bool    `json:"is_current"`
	ResolutionAction   *string `json:"resolution_action,omitempty"`
	ResolvedBy         *string `json:"resolved_by,omitempty"`
	ResolutionDate     *string `json:"resolution_date,omitempty"`
	VerificationStatus *string `json:"verification_status,omitempty"`
	VerifiedBy         *string `json:"verified_by,omitempty"`
	VerificationDate   *string `json:"verification_date,omitempty"`
	PreventiveAction   *string `json:"preventive_action,omitempty"`
	ProcessChange      *string `json:"process_change,omitempty"`
	PreventiveOwner    *string `json:"preventive_owner,omitempty"`
	DueDate            *string `json:"due_date,omitempty"`
}
