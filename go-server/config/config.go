package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config representa as configurações da aplicação
type Config struct {
	PostgresHost     string
	PostgresPort     int
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	PostgresSSLMode  string
	RedisAddr        string
	// PostgresURL is built internally from the individual parameters
	PostgresURL string
}

// LoadConfig carrega as variáveis de ambiente e retorna uma estrutura Config
func LoadConfig() (*Config, error) {
	// Carrega variáveis de ambiente do arquivo .env, se existir
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresSSLMode:  getEnvWithDefault("POSTGRES_SSLMODE", "prefer"),
		RedisAddr:        os.Getenv("REDIS_ADDR"),
	}

	// Debug: Print all environment variables
	fmt.Printf("🔍 DEBUG CONFIG:\n")
	fmt.Printf("  POSTGRES_HOST: '%s'\n", config.PostgresHost)
	fmt.Printf("  POSTGRES_DB: '%s'\n", config.PostgresDB)
	fmt.Printf("  POSTGRES_USER: '%s'\n", config.PostgresUser)
	fmt.Printf("  POSTGRES_PASSWORD: '%s'\n", maskPassword(config.PostgresPassword))
	fmt.Printf("  POSTGRES_SSLMODE: '%s'\n", config.PostgresSSLMode)
	fmt.Printf("  REDIS_ADDR: '%s'\n", config.RedisAddr)
	// Parse PostgreSQL port
	if portStr := os.Getenv("POSTGRES_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
		}
		config.PostgresPort = port
	} else {
		config.PostgresPort = 5432 // default PostgreSQL port
	}

	// Validate database configuration
	if config.PostgresHost == "" || config.PostgresUser == "" || config.PostgresDB == "" {
		return nil, fmt.Errorf("POSTGRES_HOST, POSTGRES_USER, and POSTGRES_DB must be set")
	}

	// Build PostgresURL from individual parameters
	config.PostgresURL = buildPostgresURL(config)

	if config.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR not set")
	}

	return config, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// maskPassword masks password for logging
func maskPassword(password string) string {
	if password == "" {
		return "<empty>"
	}
	return "***"
}

// maskURLPassword masks password in connection URL
func maskURLPassword(connURL string) string {
	if connURL == "" {
		return "<empty>"
	}
	// Simple masking: replace anything between :// and @
	if idx := strings.Index(connURL, "://"); idx != -1 {
		if idx2 := strings.Index(connURL[idx:], "@"); idx2 != -1 {
			user := connURL[idx+3 : idx+idx2]
			if pidx := strings.Index(user, ":"); pidx != -1 {
				return connURL[:idx+3+pidx+1] + "***" + connURL[idx+idx2:]
			}
		}
	}
	return connURL
}

// buildPostgresURL constructs PostgreSQL connection URL from individual parameters
func buildPostgresURL(config *Config) string {
	password := ""
	if config.PostgresPassword != "" {
		password = ":" + config.PostgresPassword
	}

	url := fmt.Sprintf("postgres://%s%s@%s:%d/%s?sslmode=%s",
		config.PostgresUser,
		password,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDB,
		config.PostgresSSLMode,
	)

	fmt.Printf("🔧 Built PostgresURL: %s\n", maskURLPassword(url))
	return url
}
