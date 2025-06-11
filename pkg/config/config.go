package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from environment.
type Config struct {
	DBAddr   string
	HTTPAddr string
}

// Load reads .env (if present) and required variables from the environment.
// It logs.Fatal if any mandatory variable is missing.
func Load() *Config {
	// Load .env silently; it's ok if the file is absent in production.
	_ = godotenv.Load()

	cfg := &Config{
		DBAddr:   os.Getenv("SUPABASE_DB_URL"),
		HTTPAddr: os.Getenv("HTTP_ADDR"),
	}

	if cfg.DBAddr == "" {
		log.Fatal("missing required environment variables: SUPABASE_DB_URL")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	return cfg
}
