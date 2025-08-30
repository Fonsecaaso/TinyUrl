package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config representa as configurações da aplicação
type Config struct {
	PostgresURL      string
	PostgresHost     string
	PostgresPort     int
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	PostgresSSLMode  string
	RedisAddr        string
}

// LoadConfig carrega as variáveis de ambiente e retorna uma estrutura Config
func LoadConfig() (*Config, error) {
	// Carrega variáveis de ambiente do arquivo .env, se existir
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		PostgresURL:      os.Getenv("POSTGRES_URL"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresSSLMode:  getEnvWithDefault("POSTGRES_SSLMODE", "prefer"),
		RedisAddr:        os.Getenv("REDIS_ADDR"),
	}
	print(" CONFIG:\n\n", config)
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
	if config.PostgresURL == "" {
		// If PostgresURL is not set, validate individual parameters
		if config.PostgresHost == "" || config.PostgresUser == "" || config.PostgresDB == "" {
			return nil, fmt.Errorf("either POSTGRES_URL or POSTGRES_HOST, POSTGRES_USER, and POSTGRES_DB must be set")
		}
		// Build PostgresURL from individual parameters
		config.PostgresURL = buildPostgresURL(config)
	}
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

// buildPostgresURL constructs PostgreSQL connection URL from individual parameters
func buildPostgresURL(config *Config) string {
	password := ""
	if config.PostgresPassword != "" {
		password = ":" + config.PostgresPassword
	}

	return fmt.Sprintf("postgres://%s%s@%s:%d/%s?sslmode=%s",
		config.PostgresUser,
		password,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDB,
		config.PostgresSSLMode,
	)
}
