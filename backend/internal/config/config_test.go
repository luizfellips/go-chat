package config

import (
	"testing"
)

func testGetenv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	t.Parallel()
	cfg, err := Load(testGetenv(map[string]string{
		"APP_ENV":           "development",
		"JWT_ACCESS_SECRET": "dev-access-secret-change-in-production-32",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppEnv != AppEnvDevelopment {
		t.Fatalf("AppEnv = %q", cfg.AppEnv)
	}
	if cfg.WSTicketTTL.Seconds() != 30 {
		t.Fatalf("WSTicketTTL = %v", cfg.WSTicketTTL)
	}
}

func TestLoadProductionRejectsDefaultSecret(t *testing.T) {
	t.Parallel()
	_, err := Load(testGetenv(map[string]string{
		"APP_ENV":           "production",
		"JWT_ACCESS_SECRET": defaultAccessSecret,
		"DATABASE_URL":      "postgres://user:pass@db:5432/gochat?sslmode=require",
	}))
	if err == nil {
		t.Fatal("expected error for default secret in production")
	}
}

func TestLoadRejectsShortSecret(t *testing.T) {
	t.Parallel()
	_, err := Load(testGetenv(map[string]string{
		"APP_ENV":           "development",
		"JWT_ACCESS_SECRET": "short",
	}))
	if err == nil {
		t.Fatal("expected error for short secret")
	}
}

func TestLoadProductionRequiresStrongSecret(t *testing.T) {
	t.Parallel()
	cfg, err := Load(testGetenv(map[string]string{
		"APP_ENV":           "production",
		"JWT_ACCESS_SECRET": "production-secret-that-is-long-enough-32",
		"DATABASE_URL":      "postgres://user:strongpass@db:5432/gochat?sslmode=require",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.IsProduction() {
		t.Fatal("expected production config")
	}
}
