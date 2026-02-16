package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/keycloak"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/utils"
)

type ImportService struct {
	store db.Store
	kc    *keycloak.KeyAdminClient
}

func NewImportService(store db.Store, kc *keycloak.KeyAdminClient) *ImportService {
	return &ImportService{
		store: store,
		kc:    kc,
	}
}

func (s *ImportService) PreviewCSV(ctx context.Context, r io.Reader, fileName, createdBy string) (*model.PreviewResponse, error) {
	jobID := uuid.New()

	// Create job first (previewed)
	if err := s.store.CreateImportJob(ctx, db.CreateImportJobParams{
		ID:           jobID,
		Filename:     fileName,
		Status:       db.ImportJobStatusPreviewed,
		TotalRows:    0,
		ValidRows:    0,
		SuccessCount: 0,
		FailureCount: 0,
		CreatedBy:    createdBy,
	}); err != nil {
		return nil, fmt.Errorf("create import job: %w", err)
	}

	csvr := csv.NewReader(r)
	csvr.TrimLeadingSpace = true

	header, err := csvr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	col := map[string]int{}
	for i, h := range header {
		col[utils.NormalizeHeader(h)] = i
	}

	required := []string{"username", "email", "first_name", "last_name", "role"}
	for _, k := range required {
		if _, ok := col[k]; !ok {
			return nil, fmt.Errorf("missing required column: %s", k)
		}
	}

	// Optional
	_, hasEnabled := col["enabled"]
	_, hasClientIDs := col["client_ids"]

	var (
		total   = 0
		valid   = 0
		invalid = 0
		rows    []model.ImportUserRow
	)

	// Pre-fetch roles existence (admin/user)
	// If you want to allow more roles, you can validate against Keycloak role endpoint.
	allowedRoles := map[string]bool{"admin": true, "user": true}

	for {
		rec, err := csvr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}

		total++
		rowNum := total + 1 // +1 because header is line 1

		get := func(k string) string {
			i, ok := col[k]
			if !ok || i >= len(rec) {
				return ""
			}
			return strings.TrimSpace(rec[i])
		}

		user := model.ImportUserRow{
			RowNumber: rowNum,
			Username:  get("username"),
			Email:     get("email"),
			FirstName: get("first_name"),
			LastName:  get("last_name"),
			Roles:     []string{strings.ToLower(get("roles"))},
			Enabled:   true,
		}

		if hasEnabled {
			b, berr := utils.ParseBoolDefaultTrue(get("enabled"))
			if berr != nil {
				user.Errors = append(user.Errors, berr.Error())
			} else {
				user.Enabled = b
			}
		}

		var clientIDsRaw string
		if hasClientIDs {
			clientIDsRaw = get("client_ids")
			user.ClientIDs = utils.SplitClientIDs(clientIDsRaw)
		}

		// Basic validations
		if user.Username == "" {
			user.Errors = append(user.Errors, "username is required")
		}
		if user.Email == "" || !strings.Contains(user.Email, "@") {
			user.Errors = append(user.Errors, "valid email is required")
		}
		if user.FirstName == "" {
			user.Errors = append(user.Errors, "first_name is required")
		}
		if user.LastName == "" {
			user.Errors = append(user.Errors, "last_name is required")
		}
		if len(user.Roles) == 0 || !allowedRoles[user.Roles[0]] {
			user.Errors = append(user.Errors, "role must be one of: admin,user")
		}

		// Check duplicates in Keycloak (username OR email)
		// Keep this in preview so admin sees issues before execute.
		if len(user.Errors) == 0 {
			// Search by username
			foundU, _ := s.kc.FindUsers(ctx, user.Username, true)
			for _, fu := range foundU {
				if strings.EqualFold(fu.Username, user.Username) {
					user.Errors = append(user.Errors, "username already exists")
					break
				}
			}
			// Search by email (Keycloak search might include email)
			foundE, _ := s.kc.FindUsers(ctx, user.Email, false)
			for _, fe := range foundE {
				if strings.EqualFold(fe.Email, user.Email) {
					user.Errors = append(user.Errors, "email already exists")
					break
				}
			}
		}

		if len(user.Errors) == 0 {
			user.Status = "valid"
			valid++
		} else {
			user.Status = "invalid"
			invalid++
		}

		rows = append(rows, user)

		// Persist each item for history & later execute
		_ = s.store.InsertImportJobItem(ctx, db.InsertImportJobItemParams{
			ID:        uuid.New(),
			JobID:     jobID,
			RowNumber: int32(user.RowNumber),
			Username: sql.NullString{
				String: user.Username,
				Valid:  user.Username != "",
			},
			Email: sql.NullString{
				String: user.Email,
				Valid:  user.Email != "",
			},
			FirstName: sql.NullString{
				String: user.FirstName,
				Valid:  user.FirstName != "",
			},
			LastName: sql.NullString{
				String: user.LastName,
				Valid:  user.LastName != "",
			},
			Roles: user.Roles,
			Enabled: sql.NullBool{
				Bool:  user.Enabled,
				Valid: true,
			},
			ClientIds: sql.NullString{
				String: clientIDsRaw,
				Valid:  clientIDsRaw != "",
			},
			Status: "pending",
			ErrorMsg: sql.NullString{
				String: strings.Join(user.Errors, "; "),
				Valid:  len(user.Errors) > 0,
			},
		})
	}

	// Update counts
	if err := s.store.UpdateImportJobCounts(ctx, db.UpdateImportJobCountsParams{
		ID:           jobID,
		TotalRows:    int32(total),
		ValidRows:    int32(valid),
		SuccessCount: 0,
		FailureCount: 0,
	}); err != nil {
		return nil, fmt.Errorf("update job counts: %w", err)
	}

	return &model.PreviewResponse{
		JobID:    jobID.String(),
		FileName: fileName,
		Total:    total,
		Valid:    valid,
		Invalid:  invalid,
		Rows:     rows,
	}, nil
}

func (s *ImportService) Execute(ctx context.Context, jobID uuid.UUID) (*model.ExecuteResponse, error) {
	// mark running
	if err := s.store.UpdateImportJobStatus(ctx, db.UpdateImportJobStatusParams{
		ID:     jobID,
		Status: db.ImportJobStatusRunning,
	}); err != nil {
		return nil, err
	}

	job, err := s.store.GetImportJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListImportJobItems(ctx, jobID)
	if err != nil {
		return nil, err
	}

	success := 0
	fail := 0

	// Pre-fetch roles
	adminRole, _ := s.kc.GetRealmRoleByName(ctx, "admin")
	userRole, _ := s.kc.GetRealmRoleByName(ctx, "user")

	for _, it := range items {
		// If preview already had errors, mark failed (skip calling KC)
		if it.ErrorMsg.Valid && strings.TrimSpace(it.ErrorMsg.String) != "" {
			fail++
			_ = s.store.UpdateImportJobItemStatus(ctx, db.UpdateImportJobItemStatusParams{
				ID:       it.ID,
				Status:   "failed",
				ErrorMsg: it.ErrorMsg,
			})
			continue
		}

		username := it.Username.String
		email := it.Email.String
		firstName := it.FirstName.String
		lastName := it.LastName.String
		role := strings.ToLower(it.Roles[0])

		// Idempotency: if username exists now, skip
		foundU, _ := s.kc.FindUsers(ctx, username, true)
		exists := false
		for _, fu := range foundU {
			if strings.EqualFold(fu.Username, username) {
				exists = true
				break
			}
		}
		if exists {
			_ = s.store.UpdateImportJobItemStatus(ctx, db.UpdateImportJobItemStatusParams{
				ID:     it.ID,
				Status: "skipped",
				ErrorMsg: sql.NullString{
					String: "user already exists at execution time",
					Valid:  true,
				},
			})
			continue
		}

		req := keycloak.CreateUserRequest{
			Username:      username,
			Email:         email,
			FirstName:     firstName,
			LastName:      lastName,
			Enabled:       it.Enabled.Bool,
			EmailVerified: true,
		}
		req.Credentials = append(req.Credentials, struct {
			Type      string `json:"type"`
			Value     string `json:"value"`
			Temporary bool   `json:"temporary"`
		}{
			Type:      "password",
			Value:     "",
			Temporary: true,
		})

		user := &model.User{
			Username:  req.Username,
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  lastName,
			Enabled:   it.Enabled.Bool,
		}

		userID, err := s.kc.CreateUser(user)
		if err != nil {
			fail++
			_ = s.store.UpdateImportJobItemStatus(ctx, db.UpdateImportJobItemStatusParams{
				ID:     it.ID,
				Status: "failed",
				ErrorMsg: sql.NullString{
					String: err.Error(),
					Valid:  true,
				},
			})
			continue
		}

		// Assign role
		var rr *keycloak.RoleRep
		if role == "admin" {
			rr = adminRole
		} else {
			rr = userRole
		}
		if rr != nil && rr.Name != "" {
			_ = s.kc.AddRealmRoleToUser(ctx, userID, *rr)
		}

		success++
		_ = s.store.UpdateImportJobItemStatus(ctx, db.UpdateImportJobItemStatusParams{
			ID:     it.ID,
			Status: "success",
			ErrorMsg: sql.NullString{
				String: "",
				Valid:  false,
			},
		})

		// Small throttle if needed
		time.Sleep(30 * time.Millisecond)
	}

	if err := s.store.UpdateImportJobCounts(ctx, db.UpdateImportJobCountsParams{
		ID:           jobID,
		TotalRows:    job.TotalRows,
		ValidRows:    job.ValidRows,
		SuccessCount: int32(success),
		FailureCount: int32(fail),
	}); err != nil {
		return nil, err
	}

	finalStatus := db.ImportJobStatusCompleted
	if err := s.store.UpdateImportJobStatus(ctx, db.UpdateImportJobStatusParams{
		ID:     jobID,
		Status: finalStatus,
	}); err != nil {
		return nil, err
	}

	return &model.ExecuteResponse{
		JobID:        jobID.String(),
		Total:        int(job.TotalRows),
		SuccessCount: success,
		FailureCount: fail,
		Status:       string(finalStatus),
	}, nil
}

func (s *ImportService) GetJob(ctx context.Context, jobID uuid.UUID) (*model.JobStatusResponse, error) {
	job, err := s.store.GetImportJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	items, err := s.store.ListImportJobItems(ctx, jobID)
	if err != nil {
		return nil, err
	}

	rows := make([]model.ImportUserRow, 0, len(items))
	for _, it := range items {
		enabled := true
		if it.Enabled.Valid {
			enabled = it.Enabled.Bool
		}
		row := model.ImportUserRow{
			RowNumber: int(it.RowNumber),
			Username:  it.Username.String,
			Email:     it.Email.String,
			FirstName: it.FirstName.String,
			LastName:  it.LastName.String,
			Roles:     it.Roles,
			Enabled:   enabled,
			ClientIDs: utils.SplitClientIDs(it.ClientIds.String),
			Status:    it.Status,
		}
		if it.ErrorMsg.Valid && it.ErrorMsg.String != "" {
			row.ErrorMsg = it.ErrorMsg.String
		}
		rows = append(rows, row)
	}

	return &model.JobStatusResponse{
		JobID:        job.ID.String(),
		FileName:     job.Filename,
		Status:       string(job.Status),
		Total:        int(job.TotalRows),
		Valid:        int(job.ValidRows),
		SuccessCount: int(job.SuccessCount),
		FailureCount: int(job.FailureCount),
		Rows:         rows,
	}, nil
}

func (s *ImportService) BuildErrorCSV(rows []model.ImportUserRow) string {
	// keep it simple: generate CSV string
	var b strings.Builder
	b.WriteString("row_number,username,email,reason\n")
	for _, r := range rows {
		if r.Status == "failed" || (r.ErrorMsg != "" && r.Status != "success") {
			reason := r.ErrorMsg
			if reason == "" && len(r.Errors) > 0 {
				reason = strings.Join(r.Errors, "; ")
			}
			b.WriteString(strconv.Itoa(r.RowNumber))
			b.WriteString(",")
			b.WriteString(utils.EscapeCSV(r.Username))
			b.WriteString(",")
			b.WriteString(utils.EscapeCSV(r.Email))
			b.WriteString(",")
			b.WriteString(utils.EscapeCSV(reason))
			b.WriteString("\n")
		}
	}
	return b.String()
}
