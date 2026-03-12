package service

import (
	"testing"
	"time"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/domain"
)

func TestAccessTokenExpiryForWebIsOneHour(t *testing.T) {
	service := &AuthService{
		cfg: config.Config{
			JWTExpiry:      24 * time.Hour,
			JWTExpiryUser:  24 * time.Hour,
			JWTExpiryAdmin: 24 * time.Hour,
		},
	}

	got := service.accessTokenExpiry(domain.User{Role: domain.RoleRegistered}, "web")
	if got != time.Hour {
		t.Fatalf("expected 1h for web client, got %s", got)
	}

	got = service.accessTokenExpiry(domain.User{Role: domain.RoleAdmin}, "web")
	if got != time.Hour {
		t.Fatalf("expected 1h for web admin client, got %s", got)
	}
}

func TestRefreshTokenPolicyDiffersByClientType(t *testing.T) {
	if shouldIssueRefreshToken("web") {
		t.Fatal("web client should not receive refresh token")
	}
	if !shouldIssueRefreshToken("mobile") {
		t.Fatal("mobile client should receive refresh token")
	}
}
