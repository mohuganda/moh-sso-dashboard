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
	FrontendBaseURL string `mapstructure:"FRONTEND_BASE_URL"`
	LoginUrl        string `mapstructure:"LOGIN_URL"`

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
	KeycloakBaseUrl     string `mapstructure:"KEYCLOAK_BASE_URL"`
	KeycloakRealm       string `mapstructure:"KEYCLOAK_REALM"`

	KeycloakAdminClientID     string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_ID"`
	KeycloakAdminClientSecret string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_SECRET"`

	KeycloakWebClientID     string `mapstructure:"KEYCLOAK_WEB_CLIENT_ID"`
	KeycloakWebClientSecret string `mapstructure:"KEYCLOAK_WEB_CLIENT_SECRET"`
	KeycloakRedirectUri     string `mapstructure:"KEYCLOAK_REDIRECT_URI"`

	// ==================================================
	// Backend
	// ==================================================
	ServerPort string `mapstructure:"SERVER_PORT"`

	// ==================================================
	// TLS
	// ==================================================
	EnableTls bool   `mapstructure:"ENABLE_TLS"`
	TlsCert   string `mapstructure:"TLS_CERT"`
	TlsKey    string `mapstructure:"TLS_KEY"`

	// ==================================================
	// Database
	// ==================================================
	DbDriver    string `mapstructure:"DB_DRIVER"`
	DbHost      string `mapstructure:"DB_HOST"`
	DbUser      string `mapstructure:"DB_USER"`
	DbPassword  string `mapstructure:"DB_PASSWORD"`
	DbName      string `mapstructure:"DB_NAME"`
	DbPort      string `mapstructure:"DB_PORT"`
	DbEnableSsl bool   `mapstructure:"DB_ENABLE_SSL"`

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

	// =================================================
	// storage
	// =================================================
	StorageProvider string `mapstructure:"STORAGE_PROVIDER"`
	LocalBasePath   string `mapstructure:"LOCAL_BASE_PATH"`
	NFSBasePath     string `mapstructure:"NFS_BASE_PATH"`

	// remote db
	RemoteDBHost     string `mapstructure:"REMOTE_DB_HOST"`
	RemoteDBPort     string `mapstructure:"REMOTE_DB_PORT"`
	RemoteDBUser     string `mapstructure:"REMOTE_DB_USER"`
	RemoteDBPassword string `mapstructure:"REMOTE_DB_PASSWORD"`
	RemoteDBName     string `mapstructure:"REMOTE_DB_NAME"`

	// s3 / minio
	S3Client          string `mapstructure:"S3_CLIENT"`
	S3Bucket          string `mapstructure:"S3_BUCKET"`
	S3Region          string `mapstructure:"S3_REGION"`
	S3AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	S3SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`

	MinioEndpoint        string `mapstructure:"MINIO_ENDPOINT"`
	MinioRegion          string `mapstructure:"MINIO_REGION"`
	MinioBucket          string `mapstructure:"MINIO_BUCKER"`
	MinioAccessKeyID     string `mapstructure:"MINIO_ACCESS_KEY_ID"`
	MinioSecretAccessKey string `mapstructure:"MINIO_SECRET_ACCESS_KEY"`
}

func LoadConfig(path string) (*Config, error) {

	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)

	// Optional config file
	if err := viper.ReadInConfig(); err != nil {
		// Only warn if file missing
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// ENV ALWAYS WINS
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)

	// Explicit binding (bulletproof for K8s)
	requiredKeys := []string{
		"DB_USER",
		"DB_PASSWORD",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"KEYCLOAK_WEB_CLIENT_SECRET",
		"KEYCLOAK_ADMIN_CLIENT_SECRET",
	}

	for _, key := range requiredKeys {
		_ = viper.BindEnv(key)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) DbSource() string {

	if c.DbUser == "" {
		panic("DB_USER is empty")
	}

	dbPassEscaped := url.QueryEscape(c.DbPassword)

	sslMode := "disable"
	if c.DbEnableSsl {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DbUser,
		dbPassEscaped,
		c.DbHost,
		c.DbPort,
		c.DbName,
		sslMode,
	)
}

func (c *Config) RemoteDbSource() string {

	if c.RemoteDBUser == "" {
		panic("REMOTE_DB_USER is empty")
	}

	dbPassEscaped := url.QueryEscape(c.RemoteDBPassword)

	sslMode := "disable"
	if c.DbEnableSsl {
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

func validateConfig(c *Config) error {

	if c.DbUser == "" {
		return errors.New("DB_USER is required")
	}
	if c.DbPassword == "" {
		return errors.New("DB_PASSWORD is required")
	}
	if c.DbHost == "" {
		return errors.New("DB_HOST is required")
	}
	return nil
}
