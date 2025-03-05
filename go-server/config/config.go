package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config representa as configurações da aplicação
type Config struct {
	PostgresURL string
	RedisAddr   string
}

// LoadConfig carrega as variáveis de ambiente e retorna uma estrutura Config
func LoadConfig() (*Config, error) {
	// Carrega variáveis de ambiente do arquivo .env, se existir
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	config := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
	}

	if config.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL not set")
	}
	if config.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR not set")
	}

	return config, nil
}
