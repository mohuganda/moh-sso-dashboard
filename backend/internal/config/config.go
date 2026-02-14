package config

import (
	"fmt"
	"net/url"
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

	// 🔐 Backend automation (client_credentials)
	KeycloakAdminClientID     string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_ID"`
	KeycloakAdminClientSecret string `mapstructure:"KEYCLOAK_ADMIN_CLIENT_SECRET"`

	// 🌍 Web login (authorization_code)
	KeycloakWebClientID     string `mapstructure:"KEYCLOAK_WEB_CLIENT_ID"`
	KeycloakWebClientSecret string `mapstructure:"KEYCLOAK_WEB_CLIENT_SECRET"`
	KeycloakRedirectUri     string `mapstructure:"KEYCLOAK_REDIRECT_URI"`

	// ==================================================
	// Backend (Go / Gin)
	// ==================================================
	ServerPort string `mapstructure:"SERVER_PORT"`

	// ==================================================
	// TLS
	// ==================================================
	EnableTls          bool   `mapstructure:"ENABLE_TLS"`
	TlsCert            string `mapstructure:"TLS_CERT"`
	TlsKey             string `mapstructure:"TLS_KEY"`
	KeycloakCACertPath string `mapstructure:"TLS_KEYCLOAK_CA_CERT_PATH"`

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
	// App-level JWT (NOT Keycloak)
	// ==================================================
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Config) DbSource() string {
	dbPassEscaped := url.QueryEscape(config.DbPassword)

	sslMode := "disable"
	if config.DbEnableSsl {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.DbUser,
		dbPassEscaped,
		config.DbHost,
		config.DbPort,
		config.DbName,
		sslMode,
	)
}
