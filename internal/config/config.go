package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppName  string
	HTTPAddr string
	GinMode  string
	Database Database
	Session  Session
}

type Session struct {
	Secret       string
	CookieSecure bool
}

type Database struct {
	Enabled  bool
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func Load() (Config, error) {
	enabled, err := strconv.ParseBool(value("DB_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("DB_ENABLED must be true or false: %w", err)
	}
	cookieSecure, err := strconv.ParseBool(value("SESSION_COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("SESSION_COOKIE_SECURE must be true or false: %w", err)
	}
	sessionSecret := os.Getenv("SESSION_SECRET")
	if len(sessionSecret) < 32 {
		return Config{}, fmt.Errorf("SESSION_SECRET must contain at least 32 characters")
	}
	mode := value("GIN_MODE", "debug")
	if mode != "debug" && mode != "release" && mode != "test" {
		return Config{}, fmt.Errorf("invalid GIN_MODE: %s", mode)
	}
	cfg := Config{
		AppName:  value("APP_NAME", "MK Siaga"),
		HTTPAddr: value("HTTP_ADDR", "127.0.0.1:8080"),
		GinMode:  mode,
		Database: Database{
			Enabled:  enabled,
			Host:     value("DB_HOST", "127.0.0.1"),
			Port:     value("DB_PORT", "3306"),
			Name:     value("DB_NAME", "mksiaga_dev"),
			User:     value("DB_USER", "root"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		Session: Session{
			Secret:       sessionSecret,
			CookieSecure: cookieSecure,
		},
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
