package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	OAuth    OAuthConfig
	Redis    RedisConfig
	Logging  LoggingConfig
	Provider ProviderConfig
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	FrontendURL     string
	Environment     string
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	// For SQLite
	Path string
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	JWTSecret           string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	BCryptCost          int
	SessionSecret       string
}

// OAuthConfig contains OAuth provider configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig
	GitHub GitHubOAuthConfig
}

// GoogleOAuthConfig contains Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GitHubOAuthConfig contains GitHub OAuth configuration
type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Password string
	DB       int
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string
	Format     string // json or console
	OutputPath string
}

// ProviderConfig contains cloud provider and integration configuration
type ProviderConfig struct {
	OpenAIAPIKey    string
	SlackWebhookURL string
	SlackChannel    string
	StripeAPIKey    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore errors as it's optional)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "infraudit"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			Path:            getEnv("DB_PATH", "./data.db"),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("JWT_SECRET", "supersecretkey"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			BCryptCost:         getEnvAsInt("BCRYPT_COST", 12),
			SessionSecret:      getEnv("SESSION_SECRET", "session-secret-key"),
		},
		OAuth: OAuthConfig{
			Google: GoogleOAuthConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
			},
			GitHub: GitHubOAuthConfig{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/auth/github/callback"),
			},
		},
		Redis: RedisConfig{
			Enabled:  getEnvAsBool("REDIS_ENABLED", false),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT", "stdout"),
		},
		Provider: ProviderConfig{
			OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
			SlackWebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
			SlackChannel:    getEnv("SLACK_CHANNEL", "#alerts"),
			StripeAPIKey:    getEnv("STRIPE_API_KEY", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "supersecretkey" {
		return fmt.Errorf("JWT_SECRET must be set and should not use default value in production")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}

	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
