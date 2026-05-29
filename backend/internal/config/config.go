package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// ==================================================
	// Global
	// ==================================================
	Environment string `mapstructure:"ENVIRONMENT"`
	GinMode     string `mapstructure:"GIN_MODE"`

	// ==================================================
	// Frontend
	// ==================================================
	FrontendBaseURL     string `mapstructure:"FRONTEND_BASE_URL"`
	FrontendRedirectURI string `mapstructure:"FRONTEND_REDIRECT_URI"`
	LoginURL            string `mapstructure:"LOGIN_URL"`

	// ==================================================
	// Keycloak (Infrastructure)
	// ==================================================
	KeycloakVersion   string `mapstructure:"KEYCLOAK_VERSION"`
	KeycloakDB        string `mapstructure:"KEYCLOAK_DB"`
	KeycloakDBName    string `mapstructure:"KEYCLOAK_DB_NAME"`
	KeycloakDBUser    string `mapstructure:"KEYCLOAK_DB_USER"`
	KeycloakDBPass    string `mapstructure:"KEYCLOAK_DB_PASSWORD"`
	KeycloakAdmin     string `mapstructure:"KEYCLOAK_ADMIN"`
	KeycloakAdminPass string `mapstructure:"KEYCLOAK_ADMIN_PASSWORD"`
	KeycloakHostname  string `mapstructure:"KEYCLOAK_HOSTNAME"`

	// ==================================================
	// Keycloak (Application usage)
	// ==================================================
	KeycloakInternalURL string `mapstructure:"KEYCLOAK_INTERNAL_URL"`
	KeycloakExternalURL string `mapstructure:"KEYCLOAK_EXTERNAL_URL"`
	KeycloakBaseURL     string `mapstructure:"KEYCLOAK_BASE_URL"`
	KeycloakRealm       string `mapstructure:"KEYCLOAK_REALM"`

	KeycloakAdminClientID     string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_ID"`
	KeycloakAdminClientSecret string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_SECRET"`

	KeycloakWebClientID     string `mapstructure:"KEYCLOAK_WEB_CLIENT_ID"`
	KeycloakWebClientSecret string `mapstructure:"KEYCLOAK_WEB_CLIENT_SECRET"`
	KeycloakRedirectURI     string `mapstructure:"KEYCLOAK_REDIRECT_URI"`

	// ==================================================
	// Backend
	// ==================================================
	ServerPort string `mapstructure:"SERVER_PORT"`
	AppBaseURL string `mapstructure:"APP_BASE_URL"`

	// ==================================================
	// TLS
	// ==================================================
	EnableTLS             bool   `mapstructure:"ENABLE_TLS"`
	TLSCert               string `mapstructure:"TLS_CERT"`
	TLSKey                string `mapstructure:"TLS_KEY"`
	TLSKeycloakCACertPath string `mapstructure:"TLS_KEYCLOAK_CA_CERT_PATH"`

	// ==================================================
	// Database
	// ==================================================
	DBDriver    string `mapstructure:"DB_DRIVER"`
	DBHost      string `mapstructure:"DB_HOST"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	DBPort      string `mapstructure:"DB_PORT"`
	DBEnableSSL bool   `mapstructure:"DB_ENABLE_SSL"`

	// ==================================================
	// Redis
	// ==================================================
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	// ==================================================
	// JWT
	// ==================================================
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`

	// ==================================================
	// Storage
	// ==================================================
	StorageProvider string `mapstructure:"STORAGE_PROVIDER"`
	LocalBasePath   string `mapstructure:"LOCAL_BASE_PATH"`
	NFSBasePath     string `mapstructure:"NFS_BASE_PATH"`

	S3Bucket          string `mapstructure:"S3_BUCKET"`
	S3Region          string `mapstructure:"S3_REGION"`
	S3AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	S3SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`

	MinioEndpoint        string `mapstructure:"MINIO_ENDPOINT"`
	MinioRegion          string `mapstructure:"MINIO_REGION"`
	MinioBucket          string `mapstructure:"MINIO_BUCKET"`
	MinioAccessKeyID     string `mapstructure:"MINIO_ACCESS_KEY_ID"`
	MinioSecretAccessKey string `mapstructure:"MINIO_SECRET_ACCESS_KEY"`

	// ==================================================
	// Remote DB
	// ==================================================
	RemoteDBHost     string `mapstructure:"REMOTE_DB_HOST"`
	RemoteDBPort     string `mapstructure:"REMOTE_DB_PORT"`
	RemoteDBUser     string `mapstructure:"REMOTE_DB_USER"`
	RemoteDBPassword string `mapstructure:"REMOTE_DB_PASSWORD"`
	RemoteDBName     string `mapstructure:"REMOTE_DB_NAME"`

	// ==================================================
	// DWH
	// ==================================================
	DWHHost     string `mapstructure:"DWH_HOST"`
	DWHPort     string `mapstructure:"DWH_PORT"`
	DWHUsername string `mapstructure:"DWH_USERNAME"`
	DWHPassword string `mapstructure:"DWH_PASSWORD"`
	DWHDBName   string `mapstructure:"DWH_DB"`

	// ==================================================
	// SMTP / Retry / Notifications
	// ==================================================
	SMTP         SMTPConfig         `mapstructure:",squash"`
	Retry        RetryConfig        `mapstructure:",squash"`
	Notification NotificationConfig `mapstructure:",squash"`
}

type SMTPConfig struct {
	Host           string        `mapstructure:"SMTP_HOST"`
	Port           int           `mapstructure:"SMTP_PORT"`
	Username       string        `mapstructure:"SMTP_USERNAME"`
	Password       string        `mapstructure:"SMTP_PASSWORD"`
	FromEmail      string        `mapstructure:"SMTP_FROM_EMAIL"`
	FromName       string        `mapstructure:"SMTP_FROM_NAME"`
	ConnectTimeout time.Duration `mapstructure:"SMTP_CONNECT_TIMEOUT"`
	SendTimeout    time.Duration `mapstructure:"SMTP_SEND_TIMEOUT"`
}

type RetryConfig struct {
	MaxAttempts int           `mapstructure:"RETRY_MAX_ATTEMPTS"`
	BaseDelay   time.Duration `mapstructure:"RETRY_BASE_DELAY"`
}

type NotificationConfig struct {
	PlatformName      string `mapstructure:"PLATFORM_NAME"`
	SystemAdminName   string `mapstructure:"SYSTEM_ADMIN_NAME"`
	SystemAdminEmail  string `mapstructure:"SYSTEM_ADMIN_EMAIL"`
	AdminDashboardURL string `mapstructure:"ADMIN_DASHBOARD_URL"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)

	for key, env := range envBindings() {
		_ = viper.BindEnv(key, env)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	normalizeConfig(&config)

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// ==================================================
	// SMTP defaults for local MailHog
	// ==================================================
	viper.SetDefault("SMTP_HOST", "mailhog")
	viper.SetDefault("SMTP_PORT", 1025)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM_EMAIL", "noreply@moh.go.ug")
	viper.SetDefault("SMTP_FROM_NAME", "MOH Integrated Health Portal")
	viper.SetDefault("SMTP_CONNECT_TIMEOUT", "10s")
	viper.SetDefault("SMTP_SEND_TIMEOUT", "15s")

	// ==================================================
	// Retry defaults
	// ==================================================
	viper.SetDefault("RETRY_MAX_ATTEMPTS", 3)
	viper.SetDefault("RETRY_BASE_DELAY", "2s")

	// ==================================================
	// Notification defaults
	// ==================================================
	viper.SetDefault("PLATFORM_NAME", "MOH Integrated Health Portal")
	viper.SetDefault("SYSTEM_ADMIN_NAME", "System Administrator")
	viper.SetDefault("SYSTEM_ADMIN_EMAIL", "admin@example.com")
	viper.SetDefault("ADMIN_DASHBOARD_URL", "http://localhost:3000/admin/home")
}

func envBindings() map[string]string {
	return map[string]string{
		"ENVIRONMENT":                  "ENVIRONMENT",
		"GIN_MODE":                     "GIN_MODE",
		"FRONTEND_BASE_URL":            "FRONTEND_BASE_URL",
		"FRONTEND_REDIRECT_URI":        "FRONTEND_REDIRECT_URI",
		"LOGIN_URL":                    "LOGIN_URL",
		"KEYCLOAK_VERSION":             "KEYCLOAK_VERSION",
		"KEYCLOAK_DB":                  "KEYCLOAK_DB",
		"KEYCLOAK_DB_NAME":             "KEYCLOAK_DB_NAME",
		"KEYCLOAK_DB_USER":             "KEYCLOAK_DB_USER",
		"KEYCLOAK_DB_PASSWORD":         "KEYCLOAK_DB_PASSWORD",
		"KEYCLOAK_ADMIN":               "KEYCLOAK_ADMIN",
		"KEYCLOAK_ADMIN_PASSWORD":      "KEYCLOAK_ADMIN_PASSWORD",
		"KEYCLOAK_HOSTNAME":            "KEYCLOAK_HOSTNAME",
		"KEYCLOAK_INTERNAL_URL":        "KEYCLOAK_INTERNAL_URL",
		"KEYCLOAK_EXTERNAL_URL":        "KEYCLOAK_EXTERNAL_URL",
		"KEYCLOAK_BASE_URL":            "KEYCLOAK_BASE_URL",
		"KEYCLOAK_REALM":               "KEYCLOAK_REALM",
		"KEYCLOAK_ADMIN_CLIENT_ID":     "KEYCLOAK_ADMIN_CLIENT_ID",
		"KEYCLOAK_ADMIN_CLIENT_SECRET": "KEYCLOAK_ADMIN_CLIENT_SECRET",
		"KEYCLOAK_WEB_CLIENT_ID":       "KEYCLOAK_WEB_CLIENT_ID",
		"KEYCLOAK_WEB_CLIENT_SECRET":   "KEYCLOAK_WEB_CLIENT_SECRET",
		"KEYCLOAK_REDIRECT_URI":        "KEYCLOAK_REDIRECT_URI",
		"SERVER_PORT":                  "SERVER_PORT",
		"APP_BASE_URL":                 "APP_BASE_URL",
		"ENABLE_TLS":                   "ENABLE_TLS",
		"TLS_CERT":                     "TLS_CERT",
		"TLS_KEY":                      "TLS_KEY",
		"TLS_KEYCLOAK_CA_CERT_PATH":    "TLS_KEYCLOAK_CA_CERT_PATH",
		"DB_DRIVER":                    "DB_DRIVER",
		"DB_HOST":                      "DB_HOST",
		"DB_PORT":                      "DB_PORT",
		"DB_NAME":                      "DB_NAME",
		"DB_USER":                      "DB_USER",
		"DB_PASSWORD":                  "DB_PASSWORD",
		"DB_ENABLE_SSL":                "DB_ENABLE_SSL",
		"REDIS_HOST":                   "REDIS_HOST",
		"REDIS_PORT":                   "REDIS_PORT",
		"REDIS_PASSWORD":               "REDIS_PASSWORD",
		"TOKEN_SYMMETRIC_KEY":          "TOKEN_SYMMETRIC_KEY",
		"ACCESS_TOKEN_DURATION":        "ACCESS_TOKEN_DURATION",
		"REFRESH_TOKEN_DURATION":       "REFRESH_TOKEN_DURATION",
		"STORAGE_PROVIDER":             "STORAGE_PROVIDER",
		"LOCAL_BASE_PATH":              "LOCAL_BASE_PATH",
		"NFS_BASE_PATH":                "NFS_BASE_PATH",
		"S3_BUCKET":                    "S3_BUCKET",
		"S3_REGION":                    "S3_REGION",
		"S3_ACCESS_KEY_ID":             "S3_ACCESS_KEY_ID",
		"S3_SECRET_ACCESS_KEY":         "S3_SECRET_ACCESS_KEY",
		"MINIO_ENDPOINT":               "MINIO_ENDPOINT",
		"MINIO_REGION":                 "MINIO_REGION",
		"MINIO_BUCKET":                 "MINIO_BUCKET",
		"MINIO_ACCESS_KEY_ID":          "MINIO_ACCESS_KEY_ID",
		"MINIO_SECRET_ACCESS_KEY":      "MINIO_SECRET_ACCESS_KEY",
		"REMOTE_DB_HOST":               "REMOTE_DB_HOST",
		"REMOTE_DB_PORT":               "REMOTE_DB_PORT",
		"REMOTE_DB_USER":               "REMOTE_DB_USER",
		"REMOTE_DB_PASSWORD":           "REMOTE_DB_PASSWORD",
		"REMOTE_DB_NAME":               "REMOTE_DB_NAME",
		"DWH_HOST":                     "DWH_HOST",
		"DWH_PORT":                     "DWH_PORT",
		"DWH_USERNAME":                 "DWH_USERNAME",
		"DWH_PASSWORD":                 "DWH_PASSWORD",
		"DWH_DB":                       "DWH_DB",
		"SMTP_HOST":                    "SMTP_HOST",
		"SMTP_PORT":                    "SMTP_PORT",
		"SMTP_USERNAME":                "SMTP_USERNAME",
		"SMTP_PASSWORD":                "SMTP_PASSWORD",
		"SMTP_FROM_EMAIL":              "SMTP_FROM_EMAIL",
		"SMTP_FROM_NAME":               "SMTP_FROM_NAME",
		"SMTP_CONNECT_TIMEOUT":         "SMTP_CONNECT_TIMEOUT",
		"SMTP_SEND_TIMEOUT":            "SMTP_SEND_TIMEOUT",
		"RETRY_MAX_ATTEMPTS":           "RETRY_MAX_ATTEMPTS",
		"RETRY_BASE_DELAY":             "RETRY_BASE_DELAY",
		"PLATFORM_NAME":                "PLATFORM_NAME",
		"SYSTEM_ADMIN_NAME":            "SYSTEM_ADMIN_NAME",
		"SYSTEM_ADMIN_EMAIL":           "SYSTEM_ADMIN_EMAIL",
		"ADMIN_DASHBOARD_URL":          "ADMIN_DASHBOARD_URL",
	}
}

func normalizeConfig(c *Config) {
	c.Environment = strings.TrimSpace(c.Environment)
	c.GinMode = strings.TrimSpace(c.GinMode)

	c.FrontendBaseURL = strings.TrimRight(strings.TrimSpace(c.FrontendBaseURL), "/")
	c.FrontendRedirectURI = strings.TrimSpace(c.FrontendRedirectURI)
	c.LoginURL = strings.TrimSpace(c.LoginURL)

	c.ServerPort = strings.TrimSpace(c.ServerPort)
	c.AppBaseURL = strings.TrimRight(strings.TrimSpace(c.AppBaseURL), "/")

	c.DBDriver = strings.TrimSpace(c.DBDriver)
	c.DBHost = strings.TrimSpace(c.DBHost)
	c.DBUser = strings.TrimSpace(c.DBUser)
	c.DBName = strings.TrimSpace(c.DBName)
	c.DBPort = strings.TrimSpace(c.DBPort)

	c.StorageProvider = strings.ToLower(strings.TrimSpace(c.StorageProvider))
	c.LocalBasePath = strings.TrimSpace(c.LocalBasePath)
	c.NFSBasePath = strings.TrimSpace(c.NFSBasePath)

	c.SMTP.Host = strings.TrimSpace(c.SMTP.Host)
	c.SMTP.Username = strings.TrimSpace(c.SMTP.Username)
	c.SMTP.FromEmail = strings.TrimSpace(c.SMTP.FromEmail)
	c.SMTP.FromName = strings.TrimSpace(c.SMTP.FromName)

	c.Notification.PlatformName = strings.TrimSpace(c.Notification.PlatformName)
	c.Notification.SystemAdminName = strings.TrimSpace(c.Notification.SystemAdminName)
	c.Notification.SystemAdminEmail = strings.TrimSpace(c.Notification.SystemAdminEmail)
	c.Notification.AdminDashboardURL = strings.TrimSpace(c.Notification.AdminDashboardURL)

	if c.Notification.PlatformName == "" {
		c.Notification.PlatformName = "MOH Integrated Health Portal"
	}
	if c.Notification.SystemAdminName == "" {
		c.Notification.SystemAdminName = "System Administrator"
	}
	if c.Notification.AdminDashboardURL == "" {
		if c.AppBaseURL != "" {
			c.Notification.AdminDashboardURL = strings.TrimRight(c.AppBaseURL, "/") + "/admin/home"
		} else {
			c.Notification.AdminDashboardURL = "http://localhost:3000/admin/home"
		}
	}
}

func (c *Config) DbSource() string {
	if c.DBUser == "" {
		panic("DB_USER is empty")
	}

	dbPassEscaped := url.QueryEscape(c.DBPassword)

	sslMode := "disable"
	if c.DBEnableSSL {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		dbPassEscaped,
		c.DBHost,
		c.DBPort,
		c.DBName,
		sslMode,
	)
}

func (c *Config) RemoteDbSource() string {
	if c.RemoteDBUser == "" {
		panic("REMOTE_DB_USER is empty")
	}

	dbPassEscaped := url.QueryEscape(c.RemoteDBPassword)

	sslMode := "disable"
	if c.DBEnableSSL {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.RemoteDBUser,
		dbPassEscaped,
		c.RemoteDBHost,
		c.RemoteDBPort,
		c.RemoteDBName,
		sslMode,
	)
}

func (c *Config) DwhDbSource() string {
	if c.DWHUsername == "" {
		panic("DWH_USERNAME is empty")
	}

	dbPassEscaped := url.QueryEscape(c.DWHPassword)

	sslMode := "disable"
	if c.DBEnableSSL {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DWHUsername,
		dbPassEscaped,
		c.DWHHost,
		c.DWHPort,
		c.DWHDBName,
		sslMode,
	)
}

func validateConfig(c *Config) error {
	if strings.TrimSpace(c.DBUser) == "" {
		return errors.New("DB_USER is required")
	}
	if strings.TrimSpace(c.DBPassword) == "" {
		return errors.New("DB_PASSWORD is required")
	}
	if strings.TrimSpace(c.DBHost) == "" {
		return errors.New("DB_HOST is required")
	}

	if c.StorageProvider != "" {
		switch c.StorageProvider {
		case "local":
			if strings.TrimSpace(c.LocalBasePath) == "" {
				return errors.New("LOCAL_BASE_PATH is required when STORAGE_PROVIDER=local")
			}
			if strings.TrimSpace(c.AppBaseURL) == "" {
				return errors.New("APP_BASE_URL is required when STORAGE_PROVIDER=local")
			}
		case "nfs":
			if strings.TrimSpace(c.NFSBasePath) == "" {
				return errors.New("NFS_BASE_PATH is required when STORAGE_PROVIDER=nfs")
			}
			if strings.TrimSpace(c.AppBaseURL) == "" {
				return errors.New("APP_BASE_URL is required when STORAGE_PROVIDER=nfs")
			}
		case "s3":
			if strings.TrimSpace(c.S3Bucket) == "" {
				return errors.New("S3_BUCKET is required when STORAGE_PROVIDER=s3")
			}
			if strings.TrimSpace(c.S3Region) == "" {
				return errors.New("S3_REGION is required when STORAGE_PROVIDER=s3")
			}
		case "minio":
			if strings.TrimSpace(c.MinioEndpoint) == "" {
				return errors.New("MINIO_ENDPOINT is required when STORAGE_PROVIDER=minio")
			}
			if strings.TrimSpace(c.MinioBucket) == "" {
				return errors.New("MINIO_BUCKET is required when STORAGE_PROVIDER=minio")
			}
		default:
			return fmt.Errorf("unsupported STORAGE_PROVIDER: %s", c.StorageProvider)
		}
	}

	if c.SMTP.Host != "" {
		if c.SMTP.Port <= 0 {
			return errors.New("SMTP_PORT must be greater than 0")
		}
		if strings.TrimSpace(c.SMTP.FromEmail) == "" {
			return errors.New("SMTP_FROM_EMAIL is required when SMTP is enabled")
		}
		if strings.TrimSpace(c.SMTP.FromName) == "" {
			return errors.New("SMTP_FROM_NAME is required when SMTP is enabled")
		}
	}

	if c.Retry.MaxAttempts < 0 {
		return errors.New("RETRY_MAX_ATTEMPTS cannot be negative")
	}
	if c.Retry.BaseDelay < 0 {
		return errors.New("RETRY_BASE_DELAY cannot be negative")
	}

	if strings.TrimSpace(c.Notification.PlatformName) == "" {
		return errors.New("PLATFORM_NAME is required")
	}
	if strings.TrimSpace(c.Notification.SystemAdminName) == "" {
		return errors.New("SYSTEM_ADMIN_NAME is required")
	}
	if strings.TrimSpace(c.Notification.SystemAdminEmail) == "" {
		return errors.New("SYSTEM_ADMIN_EMAIL is required")
	}
	if strings.TrimSpace(c.Notification.AdminDashboardURL) == "" {
		return errors.New("ADMIN_DASHBOARD_URL is required")
	}

	return nil
}
