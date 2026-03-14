package service

import (
	"errors"
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

func TestEnsureAdminAllowed(t *testing.T) {
	tests := []struct {
		name       string
		user       domain.User
		clientType string
		wantErr    bool
	}{
		{
			name:       "non admin is always allowed",
			user:       domain.User{Role: domain.RoleRegistered},
			clientType: "mobile",
			wantErr:    false,
		},
		{
			name:       "non admin empty client type is allowed",
			user:       domain.User{Role: domain.RoleRegistered},
			clientType: "",
			wantErr:    false,
		},
		{
			name:       "admin web is allowed",
			user:       domain.User{Role: domain.RoleAdmin},
			clientType: "web",
			wantErr:    false,
		},
		{
			name:       "admin web with mixed case and whitespace is allowed",
			user:       domain.User{Role: domain.RoleAdmin},
			clientType: " Web ",
			wantErr:    false,
		},
		{
			name:       "admin mobile is rejected",
			user:       domain.User{Role: domain.RoleAdmin},
			clientType: "mobile",
			wantErr:    true,
		},
		{
			name:       "admin empty client type is rejected",
			user:       domain.User{Role: domain.RoleAdmin},
			clientType: "",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureAdminAllowed(tc.user, tc.clientType)
			if tc.wantErr {
				if !errors.Is(err, ErrAdminWebOnly) {
					t.Fatalf("expected ErrAdminWebOnly, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
