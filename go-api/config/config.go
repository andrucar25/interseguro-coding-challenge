// Package config loads the Go API runtime configuration.
package config

import (
	"os"
	"time"
)

const defaultPort = "8080"

// Config contains the Go API runtime settings.
type Config struct {
	Port               string
	NodeAPIURL         string
	NodeAPIJWTSecret   string
	NodeAPIJWTIssuer   string
	NodeAPIJWTAudience string
	CORSAllowedOrigins string
	StatisticsTimeout  time.Duration
}

// Load reads the existing Go API environment configuration.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return Config{
		Port:               port,
		NodeAPIURL:         os.Getenv("NODE_API_URL"),
		NodeAPIJWTSecret:   os.Getenv("NODE_API_JWT_SECRET"),
		NodeAPIJWTIssuer:   os.Getenv("NODE_API_JWT_ISSUER"),
		NodeAPIJWTAudience: os.Getenv("NODE_API_JWT_AUDIENCE"),
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
		StatisticsTimeout:  5 * time.Second,
	}
}
