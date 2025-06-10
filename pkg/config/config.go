package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from environment.
type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
	HTTPAddr           string
}

// Load reads .env (if present) and required variables from the environment.
// It logs.Fatal if any mandatory variable is missing.
func Load() *Config {
	// Load .env silently; it's ok if the file is absent in production.
	_ = godotenv.Load()

	cfg := &Config{
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_KEY"),
		HTTPAddr:           os.Getenv("HTTP_ADDR"),
	}

	if cfg.SupabaseURL == "" ||
		cfg.SupabaseServiceKey == "" {
		log.Fatal("missing required environment variables: SUPABASE_URL and/or SUPABASE_SERVICE_KEY")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	return cfg
}
