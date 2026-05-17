package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	ServerPort  string
	JWTSecret   string
	CORSOrigins string
	AdminUser   string
	AdminPass   string
}

// Load reads configuration from environment variables.
// It attempts to load a .env file if present (for local development).
func Load() (*Config, error) {
	// Best-effort .env load — not an error if missing (e.g. in Docker)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPass:      getEnv("DB_PASS", "postgres"),
		DBName:      getEnv("DB_NAME", "mdm"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),
		AdminUser:   getEnv("ADMIN_USERNAME", "admin"),
		AdminPass:   getEnv("ADMIN_PASSWORD", ""),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
