package config

import (
    "fmt"
    "os"
    "strings"
    "time"
)

type Config struct {
    AppName           string
    Environment       string
    HTTPPort          string
    BaseURL           string
    JWTSecret         string
    JWTExpiry         time.Duration
    DatabaseURL       string
    DatabaseMaxConns  int
    DatabaseIdleConns int
    CORSOrigins       []string
    FCMEnabled        bool
    FCMProjectID      string
    FCMCredentials    string
}

func Load() Config {
    jwtExpiry := envDuration("JWT_EXPIRY", 24*time.Hour)
    return Config{
        AppName:           envString("APP_NAME", "Unila Helpdesk API"),
        Environment:       envString("APP_ENV", "development"),
        HTTPPort:          envString("HTTP_PORT", "8080"),
        BaseURL:           envString("BASE_URL", "http://localhost:8080"),
        JWTSecret:         envString("JWT_SECRET", "change-me"),
        JWTExpiry:         jwtExpiry,
        DatabaseURL:       envString("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/unila_helpdesk?sslmode=disable"),
        DatabaseMaxConns:  envInt("DB_MAX_CONNS", 10),
        DatabaseIdleConns: envInt("DB_IDLE_CONNS", 5),
        CORSOrigins:       envCSV("CORS_ORIGINS", "*"),
        FCMEnabled:        envBool("FCM_ENABLED", false),
        FCMProjectID:      envString("FCM_PROJECT_ID", ""),
        FCMCredentials:    envString("FCM_CREDENTIALS", ""),
    }
}

func envString(key, fallback string) string {
    if value := strings.TrimSpace(os.Getenv(key)); value != "" {
        return value
    }
    return fallback
}

func envCSV(key, fallback string) []string {
    raw := envString(key, fallback)
    parts := strings.Split(raw, ",")
    values := make([]string, 0, len(parts))
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            values = append(values, trimmed)
        }
    }
    if len(values) == 0 {
        return []string{"*"}
    }
    return values
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
    if err != nil || value <= 0 {
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
