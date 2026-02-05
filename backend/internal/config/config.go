package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string
	Port       string
	DBPath     string
	CORSOrigin string
}

func Load() (Config, error) {
	// Dev convenience only | In prod, env vars come from systemd/docker.
	_ = godotenv.Load()

	cfg := Config{
		Env:        get("ENV", "dev"),
		Port:       get("PORT", "4206"),
		DBPath:     must("DB_PATH"),
		CORSOrigin: get("CORS_ORIGIN", "http://localhost:5173"),
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		// return error via validate path; but we need a value here
		return ""
	}
	return v
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func validate(cfg Config) error {
	if cfg.DBPath == "" {
		return fmt.Errorf("missing required env var: DB_PATH")
	}
	if cfg.Port == "" {
		return fmt.Errorf("missing required env var: PORT")
	}
	return nil
}
