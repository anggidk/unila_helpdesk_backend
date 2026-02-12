package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName               string
	Environment           string
	HTTPPort              string
	BaseURL               string
	TicketInitialStatus   string
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

func Load() (Config, error) {
	appName, err := envRequiredString("APP_NAME")
	if err != nil {
		return Config{}, err
	}
	environment, err := envRequiredString("APP_ENV")
	if err != nil {
		return Config{}, err
	}
	httpPort := strings.TrimSpace(os.Getenv("PORT"))
	if httpPort == "" {
		httpPort = strings.TrimSpace(os.Getenv("HTTP_PORT"))
	}
	if httpPort == "" {
		httpPort = "8080"
	}
	baseURL, err := envRequiredString("BASE_URL")
	if err != nil {
		return Config{}, err
	}
	ticketInitialStatus, err := envRequiredString("TICKET_INITIAL_STATUS")
	if err != nil {
		return Config{}, err
	}
	jwtSecret, err := envRequiredString("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	jwtExpiry, err := envRequiredDuration("JWT_EXPIRY")
	if err != nil {
		return Config{}, err
	}
	jwtExpiryUser, err := envRequiredDuration("JWT_EXPIRY_USER")
	if err != nil {
		return Config{}, err
	}
	jwtExpiryAdmin, err := envRequiredDuration("JWT_EXPIRY_ADMIN")
	if err != nil {
		return Config{}, err
	}
	jwtRefreshExpiry, err := envRequiredDuration("JWT_REFRESH_EXPIRY")
	if err != nil {
		return Config{}, err
	}
	jwtRefreshExpiryUser, err := envRequiredDuration("JWT_REFRESH_EXPIRY_USER")
	if err != nil {
		return Config{}, err
	}
	jwtRefreshExpiryAdmin, err := envRequiredDuration("JWT_REFRESH_EXPIRY_ADMIN")
	if err != nil {
		return Config{}, err
	}
	databaseURL, err := envRequiredString("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	databaseMaxConns, err := envRequiredInt("DB_MAX_CONNS")
	if err != nil {
		return Config{}, err
	}
	databaseIdleConns, err := envRequiredInt("DB_IDLE_CONNS")
	if err != nil {
		return Config{}, err
	}
	corsOrigins, err := envRequiredString("CORS_ORIGINS")
	if err != nil {
		return Config{}, err
	}
	fcmEnabled, err := envRequiredBool("FCM_ENABLED")
	if err != nil {
		return Config{}, err
	}

	fcmCredentials := ""
	if fcmEnabled {
		fcmCredentials, err = envRequiredString("FCM_CREDENTIALS")
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		AppName:               appName,
		Environment:           environment,
		HTTPPort:              httpPort,
		BaseURL:               baseURL,
		TicketInitialStatus:   ticketInitialStatus,
		JWTSecret:             jwtSecret,
		JWTExpiry:             jwtExpiry,
		JWTExpiryUser:         jwtExpiryUser,
		JWTExpiryAdmin:        jwtExpiryAdmin,
		JWTRefreshExpiry:      jwtRefreshExpiry,
		JWTRefreshExpiryUser:  jwtRefreshExpiryUser,
		JWTRefreshExpiryAdmin: jwtRefreshExpiryAdmin,
		DatabaseURL:           databaseURL,
		DatabaseMaxConns:      databaseMaxConns,
		DatabaseIdleConns:     databaseIdleConns,
		CORSOrigins:           corsOrigins,
		FCMEnabled:            fcmEnabled,
		FCMCredentials:        fcmCredentials,
	}, nil
}

func envRequiredString(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func envRequiredBool(key string) (bool, error) {
	raw, err := envRequiredString(key)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(raw) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a boolean value", key)
	}
}

func envRequiredInt(key string) (int, error) {
	raw, err := envRequiredString(key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.Atoi(raw)
	if parseErr != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be > 0", key)
	}
	return value, nil
}

func envRequiredDuration(key string) (time.Duration, error) {
	raw, err := envRequiredString(key)
	if err != nil {
		return 0, err
	}
	parsed, parseErr := time.ParseDuration(raw)
	if parseErr != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}
	return parsed, nil
}
