package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("NODE_API_URL", "")
	t.Setenv("NODE_API_JWT_SECRET", "")
	t.Setenv("NODE_API_JWT_ISSUER", "")
	t.Setenv("NODE_API_JWT_AUDIENCE", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	got := Load()
	if got.Port != defaultPort {
		t.Errorf("Port = %q, want %q", got.Port, defaultPort)
	}
	if got.NodeAPIURL != "" {
		t.Errorf("NodeAPIURL = %q, want empty", got.NodeAPIURL)
	}
	if got.NodeAPIJWTSecret != "" {
		t.Errorf("NodeAPIJWTSecret = %q, want empty", got.NodeAPIJWTSecret)
	}
	if got.NodeAPIJWTIssuer != "" {
		t.Errorf("NodeAPIJWTIssuer = %q, want empty", got.NodeAPIJWTIssuer)
	}
	if got.NodeAPIJWTAudience != "" {
		t.Errorf("NodeAPIJWTAudience = %q, want empty", got.NodeAPIJWTAudience)
	}
	if got.CORSAllowedOrigins != "" {
		t.Errorf("CORSAllowedOrigins = %q, want empty", got.CORSAllowedOrigins)
	}
	if got.StatisticsTimeout != 5*time.Second {
		t.Errorf("StatisticsTimeout = %s, want %s", got.StatisticsTimeout, 5*time.Second)
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("NODE_API_URL", "https://statistics.example.test")
	t.Setenv("NODE_API_JWT_SECRET", "test-secret")
	t.Setenv("NODE_API_JWT_ISSUER", "test-issuer")
	t.Setenv("NODE_API_JWT_AUDIENCE", "test-audience")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.test,https://admin.example.test")

	got := Load()
	if got.Port != "9090" {
		t.Errorf("Port = %q, want %q", got.Port, "9090")
	}
	if got.NodeAPIURL != "https://statistics.example.test" {
		t.Errorf("NodeAPIURL = %q, want configured URL", got.NodeAPIURL)
	}
	if got.NodeAPIJWTSecret != "test-secret" {
		t.Errorf("NodeAPIJWTSecret = %q, want configured secret", got.NodeAPIJWTSecret)
	}
	if got.NodeAPIJWTIssuer != "test-issuer" {
		t.Errorf("NodeAPIJWTIssuer = %q, want configured issuer", got.NodeAPIJWTIssuer)
	}
	if got.NodeAPIJWTAudience != "test-audience" {
		t.Errorf("NodeAPIJWTAudience = %q, want configured audience", got.NodeAPIJWTAudience)
	}
	if got.CORSAllowedOrigins != "https://app.example.test,https://admin.example.test" {
		t.Errorf("CORSAllowedOrigins = %q, want configured origins", got.CORSAllowedOrigins)
	}
}
