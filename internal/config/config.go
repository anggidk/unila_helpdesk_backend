package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppName               string
	Environment           string
	HTTPPort              string
	BaseURL               string
	JWTSecret             string
	JWTExpiry             time.Duration
	JWTExpiryUser         time.Duration
	JWTExpiryAdmin        time.Duration
	JWTRefreshExpiry      time.Duration
	JWTRefreshExpiryUser  time.Duration
	JWTRefreshExpiryAdmin time.Duration
	DatabaseURL           string
	DatabaseMaxConns      int
	DatabaseIdleConns     int
	CORSOrigins           string
	FCMEnabled            bool
	FCMCredentials        string
}

func Load() Config {
	jwtExpiry := envDuration("JWT_EXPIRY", 0)
	jwtExpiryUser := envDuration("JWT_EXPIRY_USER", jwtExpiry)
	jwtExpiryAdmin := envDuration("JWT_EXPIRY_ADMIN", jwtExpiry)
	jwtRefreshExpiry := envDuration("JWT_REFRESH_EXPIRY", 0)
	jwtRefreshExpiryUser := envDuration("JWT_REFRESH_EXPIRY_USER", jwtRefreshExpiry)
	jwtRefreshExpiryAdmin := envDuration("JWT_REFRESH_EXPIRY_ADMIN", jwtRefreshExpiry)
	return Config{
		AppName:               envString("APP_NAME", ""),
		Environment:           envString("APP_ENV", ""),
		HTTPPort:              envString("HTTP_PORT", ""),
		BaseURL:               envString("BASE_URL", ""),
		JWTSecret:             envString("JWT_SECRET", ""),
		JWTExpiry:             jwtExpiry,
		JWTExpiryUser:         jwtExpiryUser,
		JWTExpiryAdmin:        jwtExpiryAdmin,
		JWTRefreshExpiry:      jwtRefreshExpiry,
		JWTRefreshExpiryUser:  jwtRefreshExpiryUser,
		JWTRefreshExpiryAdmin: jwtRefreshExpiryAdmin,
		DatabaseURL:           envString("DATABASE_URL", ""),
		DatabaseMaxConns:      envInt("DB_MAX_CONNS", 0),
		DatabaseIdleConns:     envInt("DB_IDLE_CONNS", 0),
		CORSOrigins:           envString("CORS_ORIGINS", ""),
		FCMEnabled:            envBool("FCM_ENABLED", false),
		FCMCredentials:        envString("FCM_CREDENTIALS", ""),
	}
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(envString(key, ""))
	if raw == "" {
		return fallback
	}
	return raw == "true" || raw == "1" || raw == "yes"
}

func envInt(key string, fallback int) int {
	raw := envString(key, "")
	if raw == "" {
		return fallback
	}
	var value int
	_, err := fmt.Sscanf(raw, "%d", &value)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := envString(key, "")
	if raw == "" {
		return fallback
	}
	if parsed, err := time.ParseDuration(raw); err == nil {
		return parsed
	}
	return fallback
}
