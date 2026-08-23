package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address         string
	ParserBuild     string
	SigningKeyID    string
	SigningSecret   string
	AuthSecret      string
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

func Load() (Config, error) {
	timeout, err := time.ParseDuration(value("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse shutdown timeout: %w", err)
	}
	cfg := Config{Address: value("HTTP_ADDRESS", ":8080"), ParserBuild: value("PARSER_BUILD", "static-v1"), SigningKeyID: value("SIGNING_KEY_ID", "local-demo-v1"), SigningSecret: value("SIGNING_SECRET", "local-demo-signing-secret"), AuthSecret: value("AUTH_SECRET", "local-demo-auth-secret-change-me"), ShutdownTimeout: timeout, DatabaseURL: os.Getenv("DATABASE_URL")}
	if len(cfg.SigningSecret) < 16 {
		return Config{}, fmt.Errorf("SIGNING_SECRET must have at least 16 bytes")
	}
	if len(cfg.AuthSecret) < 24 {
		return Config{}, fmt.Errorf("AUTH_SECRET must have at least 24 bytes")
	}
	return cfg, nil
}
func value(name, fallback string) string {
	if current := os.Getenv(name); current != "" {
		return current
	}
	return fallback
}
func Int(name string, fallback int) int {
	current := os.Getenv(name)
	if current == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(current)
	if err != nil {
		return fallback
	}
	return parsed
}
