package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	AppEnvDevelopment = "development"
	AppEnvProduction  = "production"

	defaultAccessSecret = "dev-access-secret-change-in-production-32"
)

type Config struct {
	DatabaseURL          string
	JWTAccessSecret      string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	CORSOrigin           string
	Port                 string
	AppEnv               string
	WSTicketTTL          time.Duration
	LoginRateLimitRPM    int
	RegisterRateLimitRPM int
	RefreshRateLimitRPM  int
	SeedDemoUsers        bool
	MetricsToken         string
	TrustedProxyCIDRs    []string
	LogFormat            string
	LogLevel             string
}

func Load(getenv func(string) string) (Config, error) {
	env := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}

	cfg := Config{
		DatabaseURL:          env("DATABASE_URL", "postgres://gochat:gochat@localhost:5432/gochat?sslmode=disable"),
		JWTAccessSecret:      env("JWT_ACCESS_SECRET", defaultAccessSecret),
		AccessTokenTTL:       parseDuration(env("ACCESS_TOKEN_TTL", "15m"), 15*time.Minute),
		RefreshTokenTTL:      parseDuration(env("REFRESH_TOKEN_TTL", "168h"), 168*time.Hour),
		CORSOrigin:           env("CORS_ORIGIN", "http://localhost:5173"),
		Port:                 env("PORT", "8080"),
		AppEnv:               env("APP_ENV", AppEnvDevelopment),
		WSTicketTTL:          parseDuration(env("WS_TICKET_TTL", "30s"), 30*time.Second),
		LoginRateLimitRPM:    parseInt(env("LOGIN_RATE_LIMIT_RPM", "5"), 5),
		RegisterRateLimitRPM: parseInt(env("REGISTER_RATE_LIMIT_RPM", "10"), 10),
		RefreshRateLimitRPM:  parseInt(env("REFRESH_RATE_LIMIT_RPM", "30"), 30),
		SeedDemoUsers:        parseBool(env("SEED_DEMO_USERS", "true")),
		MetricsToken:         env("METRICS_TOKEN", ""),
		LogFormat:            env("LOG_FORMAT", "console"),
		LogLevel:             env("LOG_LEVEL", "info"),
	}

	if v := env("TRUSTED_PROXY_CIDRS", ""); v != "" {
		cfg.TrustedProxyCIDRs = strings.Split(v, ",")
	}

	if err := cfg.validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c Config) IsProduction() bool {
	return c.AppEnv == AppEnvProduction
}

func (c Config) validate() error {
	if c.AppEnv != AppEnvDevelopment && c.AppEnv != AppEnvProduction {
		return fmt.Errorf("APP_ENV must be %q or %q", AppEnvDevelopment, AppEnvProduction)
	}

	if len(c.JWTAccessSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters")
	}

	if c.IsProduction() {
		if c.JWTAccessSecret == defaultAccessSecret {
			return fmt.Errorf("JWT_ACCESS_SECRET must not use the default dev value in production")
		}
		if strings.Contains(c.DatabaseURL, "gochat:gochat@") {
			return fmt.Errorf("DATABASE_URL must not use default credentials in production")
		}
	}

	return nil
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}

func parseInt(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

func parseBool(s string) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return s == "1" || strings.EqualFold(s, "yes")
	}
	return v
}
