package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID                                 string            `json:"id"`
	ClientID                           string            `json:"clientId"`
	Name                               string            `json:"name"`
	Description                        string            `json:"description"`
	RootURL                            string            `json:"rootUrl"`
	BaseURL                            string            `json:"baseUrl"`
	AdminURL                           string            `json:"adminUrl"`
	SurrogateAuthRequired              bool              `json:"surrogateAuthRequired"`
	Enabled                            bool              `json:"enabled"`
	AlwaysDisplayInConsole             bool              `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType            string            `json:"clientAuthenticatorType"`
	RedirectUris                       []string          `json:"redirectUris"`
	WebOrigins                         []string          `json:"webOrigins"`
	NotBefore                          int               `json:"notBefore"`
	BearerOnly                         bool              `json:"bearerOnly"`
	ConsentRequired                    bool              `json:"consentRequired"`
	StandardFlow                       bool              `json:"standardFlowEnabled"`
	ImplicitFlow                       bool              `json:"implicitFlowEnabled"`
	DirectAccess                       bool              `json:"directAccessGrantsEnabled"`
	ServiceAccounts                    bool              `json:"serviceAccountsEnabled"`
	PublicClient                       bool              `json:"publicClient"`
	FrontChannelLogout                 bool              `json:"frontchannelLogout"`
	Protocol                           string            `json:"protocol"`
	FullScopeAllowed                   bool              `json:"fullScopeAllowed"`
	NodeReRegistrationTimeout          int               `json:"nodeReRegistrationTimeout"`
	DefaultClientScopes                []string          `json:"defaultClientScopes"`
	OptionalClientScopes               []string          `json:"optionalClientScopes"`
	Access                             map[string]bool   `json:"access"`
	Secret                             string            `json:"secret"`
	ClientTemplate                     string            `json:"clientTemplate"`
	UseTemplateConfig                  bool              `json:"useTemplateConfig"`
	RootClientRealm                    string            `json:"rootClientRealm"`
	RegistrationAccessToken            string            `json:"registrationAccessToken"`
	AuthorizationServicesEnabled       bool              `json:"authorizationServicesEnabled"`
	ProtocolMappers                    []ProtocolMapper  `json:"protocolMappers"`
	DefaultRoles                       []string          `json:"defaultRoles"`
	AuthorizationURL                   string            `json:"authorizationUrl"`
	TlsRequired                        string            `json:"tlsRequired"`
	Attributes                         map[string]string `json:"attributes"`
	AuthenticationFlowBindingOverrides map[string]string `json:"authenticationFlowBindingOverrides"`
	RegisteredNodes                    map[string]int    `json:"registeredNodes"`
}

type ProtocolMapper struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Protocol        string            `json:"protocol"`
	ProtocolMapper  string            `json:"protocolMapper"`
	ConsentRequired bool              `json:"consentRequired"`
	ConsentText     string            `json:"consentText"`
	Config          map[string]string `json:"config"`
}

type User struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	FullName         string              `json:"fullName"`
	RealmRoles       []string            `json:"realmRoles"`
	IsAdmin          bool                `json:"isAdmin"`
	ClientRoles      map[string][]string `json:"clientRoles"`
	Enabled          bool                `json:"enabled"`
	EmailVerified    bool                `json:"emailVerified"`
	RequirePwdChange bool                `json:"requirePwdChange"`
	LastLoginAt      *time.Time          `json:"lastLoginAt"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
	CreatedBy        string              `json:"createdBy"`
	UpdatedBy        string              `json:"updatedBy"`
}

type UserResponse struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FullName         string              `json:"fullName"`
	IsAdmin          bool                `json:"isAdmin"`
	RealmRoles       []string            `json:"realmRoles"`
	ClientRoles      map[string][]string `json:"clientRoles"`
	IsActive         bool                `json:"isActive"`
	EmailVerified    bool                `json:"emailVerified"`
	RequirePwdChange bool                `json:"requirePwdChange"`
	LastLoginAt      *time.Time          `json:"lastLoginAt"`
	CreatedAt        *time.Time          `json:"createdAt"`
}

type ImportUserRow struct {
	RowNumber int      `json:"rowNumber"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Enabled   bool     `json:"enabled"`
	ClientIDs []string `json:"clientIds"`
	Errors    []string `json:"errors"`
	Status    string   `json:"status"`
	ErrorMsg  string   `json:"errorMsg"`
}

type PreviewResponse struct {
	JobID    string          `json:"jobId"`
	FileName string          `json:"fileName"`
	Total    int             `json:"total"`
	Valid    int             `json:"valid"`
	Invalid  int             `json:"invalid"`
	Rows     []ImportUserRow `json:"rows"`
}

type ExecuteResponse struct {
	JobID        string `json:"jobId"`
	Total        int    `json:"total"`
	SuccessCount int    `json:"successCount"`
	FailureCount int    `json:"failureCount"`
	Status       string `json:"status"`
}

type JobStatusResponse struct {
	JobID        string          `json:"jobId"`
	FileName     string          `json:"fileName"`
	Status       string          `json:"status"`
	Total        int             `json:"total"`
	Valid        int             `json:"valid"`
	SuccessCount int             `json:"successCount"`
	FailureCount int             `json:"failureCount"`
	Rows         []ImportUserRow `json:"rows"`
}

type parsedRow struct {
	row ImportUserRow
	raw struct {
		clientIDs string
	}
}

type CreateClientRoleRequest struct {
	Role        string `json:"role"`
	Description string `json:"description"`
}

type Notification struct {
	ID         uuid.UUID       `json:"id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	Severity   string          `json:"severity"`    // info | warning | critical
	TargetRole string          `json:"target_role"` // admin | super_admin | etc
	ClientID   string          `json:"client_id"`
	UserID     string          `json:"user_id"`
	Metadata   json.RawMessage `json:"metadata"` // JSONB from Postgres
	Read       bool            `json:"read"`
	CreatedAt  time.Time       `json:"created_at"`
}

type NotificationType string

const (
	LoginFailed            NotificationType = "LOGIN_FAILED"
	TokenRefreshFailed     NotificationType = "TOKEN_REFRESH_FAILED"
	LoginSucceeded         NotificationType = "LOGIN_SUCCEEDED"
	SuspiciousLogin        NotificationType = "SUSPICIOUS_LOGIN"
	UserPasswordReset      NotificationType = "USER_PASSWORD_RESET"
	PasswordResetRequested NotificationType = "PASSWORD_RESET_REQUESTED"
	PasswordResetCompleted NotificationType = "PASSWORD_RESET_COMPLETED"
	AccountLocked          NotificationType = "ACCOUNT_LOCKED"
	AccountUnlocked        NotificationType = "ACCOUNT_UNLOCKED"
)
const (
	UserImported    NotificationType = "USER_IMPORTED"
	UserCreated     NotificationType = "USER_CREATED"
	UserUpdated     NotificationType = "USER_UPDATED"
	UserDisabled    NotificationType = "USER_DISABLED"
	UserDeleted     NotificationType = "USER_DELETED"
	UserEnabled     NotificationType = "USER_ENABLED"
	UserRoleChanged NotificationType = "USER_ROLE_CHANGED"
)
const (
	ClientCreated       NotificationType = "CLIENT_CREATED"
	ClientUpdated       NotificationType = "CLIENT_UPDATED"
	ClientDeleted       NotificationType = "CLIENT_DELETED"
	ClientDisabled      NotificationType = "CLIENT_DISABLED"
	ClientEnabled       NotificationType = "CLIENT_ENABLED"
	ClientSecretRotated NotificationType = "CLIENT_SECRET_ROTATED"
)

const (
	DocumentCreated NotificationType = "DOCUMENT_CREATED"
	DocumentUpdated NotificationType = "DOCUMENT_UPDATED"
	DocumentDeleted NotificationType = "DOCUMENT_DELETED"
	DocumentEdited  NotificationType = "DOCUMENT_EDITED"
)

const (
	SystemStartup   NotificationType = "SYSTEM_STARTUP"
	SystemShutdown  NotificationType = "SYSTEM_SHUTDOWN"
	ConfigChanged   NotificationType = "CONFIG_CHANGED"
	BackupCompleted NotificationType = "BACKUP_COMPLETED"
	BackupFailed    NotificationType = "BACKUP_FAILED"
)

const (
	ImportStarted   NotificationType = "IMPORT_STARTED"
	ImportCompleted NotificationType = "IMPORT_COMPLETED"
	ImportFailed    NotificationType = "IMPORT_FAILED"
)

const (
	AuditExported   NotificationType = "AUDIT_EXPORTED"
	PolicyViolation NotificationType = "POLICY_VIOLATION"
)

const (
	ClientRoleCreated  NotificationType = "CLIENT_ROLE_CREATED"
	ClientRoleDeleted  NotificationType = "CLIENT_ROLE_DELETED"
	ClientRoleAssigned NotificationType = "CLIENT_ROLE_ASSIGNED"
	ClientRolesUpdated NotificationType = "CLIENT_TOLE_UPDATED"
	ClientRoleRemoved  NotificationType = "CLIENT_ROLE_REMOVED"
)

const (
	AnnouncementCreated   NotificationType = "ANNOUNCEMENT_CREATED"
	AnnouncementUpdated   NotificationType = "ANNOUNCEMENT_UPDATED"
	AnnouncementDeleted   NotificationType = "ANNOUNCEMENT_DELETED"
	AnnouncementArchived  NotificationType = "ANNOUNCEMENT_ARCHIVED"
	AnnouncementScheduled NotificationType = "ANNOUNCEMENT_SCHEDULED"
	AnnouncementPublished NotificationType = "ANNOUNCEMENT_PUBLISHED"
	AnnouncementDrafted   NotificationType = "ANNOUNCEMENT_DRAFTED"
)

type ProcessStatus string

const (
	ProcessStatusPENDING    ProcessStatus = "PENDING"
	ProcessStatusPROCESSING ProcessStatus = "PROCESSING"
	ProcessStatusCOMPLETED  ProcessStatus = "COMPLETED"
	ProcessStatusFAILED     ProcessStatus = "FAILED"
	ProcessStatusCANCELLED  ProcessStatus = "CANCELLED"
)

type ProcessType string

const (
	ProcessTypeCSVImport      ProcessType = "CSV_IMPORT"
	ProcessTypeExcelImport    ProcessType = "EXCEL_IMPORT"
	ProcessTypeFHIRImport     ProcessType = "FHIR_IMPORT"
	ProcessTypeUserBulkImport ProcessType = "USER_BULK_IMPORT"
)

type NotificationCategory string

const (
	SystemNotification       NotificationCategory = "SYSTEM"
	SecurityNotification     NotificationCategory = "SECURITY"
	AnnouncementNotification NotificationCategory = "ANNOUNCEMENT"
)

type AnnouncementStatus string

const (
	AnnouncementStatusDRAFT     AnnouncementStatus = "DRAFT"
	AnnouncementStatusSCHEDULED AnnouncementStatus = "SCHEDULED"
	AnnouncementStatusPUBLISHED AnnouncementStatus = "PUBLISHED"
	AnnouncementStatusARCHIVED  AnnouncementStatus = "ARCHIVED"
)

type AnnouncementLevel string

const (
	AnnouncementLevelINFO     AnnouncementLevel = "INFO"
	AnnouncementLevelSUCCESS  AnnouncementLevel = "SUCCESS"
	AnnouncementLevelWARNING  AnnouncementLevel = "WARNING"
	AnnouncementLevelCRITICAL AnnouncementLevel = "CRITICAL"
)

type AnnouncementAudienceType string

const (
	AnnouncementAudienceTypeALLUSERS        AnnouncementAudienceType = "ALL_USERS"
	AnnouncementAudienceTypeADMINSONLY      AnnouncementAudienceType = "ADMINS_ONLY"
	AnnouncementAudienceTypeSPECIFICCLIENTS AnnouncementAudienceType = "SPECIFIC_CLIENTS"
	AnnouncementAudienceTypeSPECIFICROLES   AnnouncementAudienceType = "SPECIFIC_ROLES"
	AnnouncementAudienceTypeSPECIFICUSERS   AnnouncementAudienceType = "SPECIFIC_USERS"
)

func (t NotificationType) Severity() string {
	switch t {

	// 🔴 Critical
	case SuspiciousLogin,
		AccountLocked,
		ClientDisabled,
		ImportFailed,
		BackupFailed,
		PolicyViolation:
		return "critical"

	// 🟠 Warning
	case LoginFailed,
		UserDisabled,
		UserRoleChanged,
		ConfigChanged:
		return "warning"

		// ⚪ Info
	case SystemStartup,
		BackupCompleted:
		return "info"

	// ⚪ Info
	default:
		return "info"
	}
}

func (t NotificationType) Title() string {
	switch t {
	case SystemStartup:
		return "System started"
	case SystemShutdown:
		return "System shutdown"
	case ConfigChanged:
		return "Configuration changed"
	case BackupCompleted:
		return "Backup completed"
	case BackupFailed:
		return "Backup failed"
	case LoginFailed:
		return "Failed login attempt"
	case SuspiciousLogin:
		return "Suspicious login detected"
	case UserImported:
		return "Users imported"
	case UserDisabled:
		return "User account disabled"
	case ClientDisabled:
		return "Client application disabled"
	case ImportFailed:
		return "Import job failed"
	default:
		return "System notification"
	}
}

type ProcessFilters struct {
	Status      *ProcessStatus
	ProcessType *string
}

type Pagination struct {
	Limit  int32
	Offset int32
}

//  visualiser

type Dashboard struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	URL         *string   `json:"url,omitempty"`
	Image       *string   `json:"image,omitempty"`
	ThematicID  int64     `json:"thematicId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DashboardWithThematic includes thematic area information
type DashboardWithThematic struct {
	Dashboard
	Thematic *Thematic `json:"thematic,omitempty"`
}

type Feedback struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Screenshot    *string   `json:"screenshot,omitempty"`
	ReportID      *int64    `json:"reportId,omitempty"`
	ReportName    *string   `json:"reportName,omitempty"`
	DashboardID   *int64    `json:"dashboardId,omitempty"`
	DashboardName *string   `json:"dashboardName,omitempty"`
	Message       string    `json:"message"`
	Status        string    `json:"status"`   // 'pending', 'reviewed', 'resolved'
	Priority      string    `json:"priority"` // 'low', 'medium', 'high'
	AdminNotes    *string   `json:"adminNotes,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Post struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Report struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Period     string    `json:"period"`
	Type       string    `json:"type"`
	Format     *string   `json:"format,omitempty"`
	Reportname string    `json:"reportname"`
	IsDefault  bool      `json:"isDefault"`
	ThematicID int64     `json:"thematicId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ReportWithThematic includes thematic area information
type ReportWithThematic struct {
	Report
	Thematic *Thematic `json:"thematic,omitempty"`
}

type Thematic struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
