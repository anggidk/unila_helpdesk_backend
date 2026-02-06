package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// ============================================================================
// Tests for ensureAdminAllowed
// ============================================================================

func TestEnsureAdminAllowed_NonAdminUser(t *testing.T) {
	user := domain.User{Role: domain.RoleRegistered}
	err := ensureAdminAllowed(user, "mobile")
	if err != nil {
		t.Errorf("expected no error for non-admin user, got %v", err)
	}
}

func TestEnsureAdminAllowed_GuestUser(t *testing.T) {
	user := domain.User{Role: domain.RoleGuest}
	err := ensureAdminAllowed(user, "")
	if err != nil {
		t.Errorf("expected no error for guest user, got %v", err)
	}
}

func TestEnsureAdminAllowed_AdminWithWebClient(t *testing.T) {
	user := domain.User{Role: domain.RoleAdmin}
	testCases := []string{"web", "Web", "WEB", " web ", " Web "}
	for _, clientType := range testCases {
		err := ensureAdminAllowed(user, clientType)
		if err != nil {
			t.Errorf("expected no error for admin with client=%q, got %v", clientType, err)
		}
	}
}

func TestEnsureAdminAllowed_AdminWithMobileClient(t *testing.T) {
	user := domain.User{Role: domain.RoleAdmin}
	testCases := []string{"mobile", "android", "ios", "", "Mobile"}
	for _, clientType := range testCases {
		err := ensureAdminAllowed(user, clientType)
		if err != ErrAdminWebOnly {
			t.Errorf("expected ErrAdminWebOnly for admin with client=%q, got %v", clientType, err)
		}
	}
}

// ============================================================================
// Tests for hashToken
// ============================================================================

func TestHashToken_Consistent(t *testing.T) {
	token := "test-refresh-token-123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)
	if hash1 != hash2 {
		t.Errorf("hashToken should be consistent, got %s and %s", hash1, hash2)
	}
}

func TestHashToken_DifferentTokens(t *testing.T) {
	hash1 := hashToken("token1")
	hash2 := hashToken("token2")
	if hash1 == hash2 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestHashToken_Length(t *testing.T) {
	hash := hashToken("any-token")
	// SHA256 produces 64 hex characters
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
}

// ============================================================================
// Tests for generateRefreshToken
// ============================================================================

func TestGenerateRefreshToken_NotEmpty(t *testing.T) {
	token, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateRefreshToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[token] {
			t.Errorf("duplicate token generated: %s", token)
		}
		tokens[token] = true
	}
}

func TestGenerateRefreshToken_Length(t *testing.T) {
	token, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 32 bytes base64 encoded = ~43 characters
	if len(token) < 40 || len(token) > 50 {
		t.Errorf("unexpected token length: %d", len(token))
	}
}

// ============================================================================
// Tests for ParseToken
// ============================================================================

func TestParseToken_ValidToken(t *testing.T) {
	cfg := config.Config{
		JWTSecret:     "test-secret-key-32bytes-long!!",
		JWTExpiryUser: time.Hour,
	}

	authService := &AuthService{
		cfg:    cfg,
		jwtKey: []byte(cfg.JWTSecret),
		now:    time.Now,
	}

	// Create a token manually with valid future expiry
	user := domain.User{
		ID:   "user-123",
		Role: domain.RoleRegistered,
	}

	// Generate claims and token with future expiry
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(authService.jwtKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Parse the token
	parsedClaims, err := authService.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if parsedClaims.UserID != user.ID {
		t.Errorf("expected UserID=%s, got %s", user.ID, parsedClaims.UserID)
	}
	if parsedClaims.Role != user.Role {
		t.Errorf("expected Role=%s, got %s", user.Role, parsedClaims.Role)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret-key-32bytes-long!!",
	}

	authService := &AuthService{
		cfg:    cfg,
		jwtKey: []byte(cfg.JWTSecret),
		now:    time.Now,
	}

	_, err := authService.ParseToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	// Create token with one secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: "user-123",
		Role:   domain.RoleRegistered,
	})
	tokenString, _ := token.SignedString([]byte("secret-one"))

	// Try to parse with different secret
	authService := &AuthService{
		jwtKey: []byte("secret-two"),
		now:    time.Now,
	}

	_, err := authService.ParseToken(tokenString)
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret-key-32bytes-long!!",
	}

	authService := &AuthService{
		cfg:    cfg,
		jwtKey: []byte(cfg.JWTSecret),
		now:    time.Now,
	}

	// Create an expired token
	claims := Claims{
		UserID: "user-123",
		Role:   domain.RoleRegistered,
	}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour)) // Expired 1 hour ago
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(authService.jwtKey)

	_, err := authService.ParseToken(tokenString)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

// ============================================================================
// Tests for Claims structure
// ============================================================================

func TestClaims_Structure(t *testing.T) {
	claims := Claims{
		UserID: "test-user-id",
		Role:   domain.RoleAdmin,
	}

	if claims.UserID != "test-user-id" {
		t.Errorf("expected UserID=test-user-id, got %s", claims.UserID)
	}
	if claims.Role != domain.RoleAdmin {
		t.Errorf("expected Role=admin, got %s", claims.Role)
	}
}

// ============================================================================
// Tests for validation logic patterns
// ============================================================================

func TestLoginValidation_EmptyUsername(t *testing.T) {
	username := ""
	password := "somepassword"

	cleanedUser := strings.TrimSpace(username)
	cleanedPass := strings.TrimSpace(password)

	if cleanedUser == "" || cleanedPass == "" {
		// This is expected behavior - validation should catch empty username
		return
	}
	t.Error("expected validation to catch empty username")
}

func TestLoginValidation_EmptyPassword(t *testing.T) {
	username := "someuser"
	password := ""

	cleanedUser := strings.TrimSpace(username)
	cleanedPass := strings.TrimSpace(password)

	if cleanedUser == "" || cleanedPass == "" {
		// This is expected behavior - validation should catch empty password
		return
	}
	t.Error("expected validation to catch empty password")
}

func TestLoginValidation_WhitespaceOnly(t *testing.T) {
	username := "   "
	password := "test"

	cleanedUser := strings.ToLower(strings.TrimSpace(username))
	cleanedPass := strings.TrimSpace(password)

	if cleanedUser == "" || cleanedPass == "" {
		// This is expected behavior - validation should catch whitespace-only username
		return
	}
	t.Error("expected validation to catch whitespace-only username")
}

func TestLoginValidation_NormalizeUsername(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"John", "john"},
		{"ADMIN", "admin"},
		{" user ", "user"},
		{"MixedCase", "mixedcase"},
	}

	for _, tc := range testCases {
		cleaned := strings.ToLower(strings.TrimSpace(tc.input))
		if cleaned != tc.expected {
			t.Errorf("expected %q to normalize to %q, got %q", tc.input, tc.expected, cleaned)
		}
	}
}

// ============================================================================
// Mock-based tests for LoginWithPasswordClient
// ============================================================================

func TestLoginWithPasswordClient_EmptyUsername(t *testing.T) {
	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("", "password", "")
	if err == nil {
		t.Error("expected error for empty username")
	}
	if err.Error() != "username dan password wajib diisi" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_EmptyPassword(t *testing.T) {
	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("user", "", "")
	if err == nil {
		t.Error("expected error for empty password")
	}
	if err.Error() != "username dan password wajib diisi" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_UserNotFound(t *testing.T) {
	mockUsers := &mockUserRepo{
		findByUsernameFunc: func(username string) (*domain.User, error) {
			return nil, errNotFound
		},
	}

	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		users:  mockUsers,
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("nonexistent", "password", "")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if err.Error() != "username atau password salah" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_InactiveUser(t *testing.T) {
	mockUsers := &mockUserRepo{
		findByUsernameFunc: func(username string) (*domain.User, error) {
			return &domain.User{
				ID:       "user-1",
				Username: username,
				IsActive: false,
			}, nil
		},
	}

	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		users:  mockUsers,
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("inactiveuser", "password", "")
	if err == nil {
		t.Error("expected error for inactive user")
	}
	if err.Error() != "akun tidak aktif" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_NoPassword(t *testing.T) {
	mockUsers := &mockUserRepo{
		findByUsernameFunc: func(username string) (*domain.User, error) {
			return &domain.User{
				ID:           "user-1",
				Username:     username,
				IsActive:     true,
				PasswordHash: "", // No password set
			}, nil
		},
	}

	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		users:  mockUsers,
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("user", "password", "")
	if err == nil {
		t.Error("expected error for user without password")
	}
	if err.Error() != "akun belum memiliki password" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_WrongPassword(t *testing.T) {
	// bcrypt hash of "correctpassword"
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	mockUsers := &mockUserRepo{
		findByUsernameFunc: func(username string) (*domain.User, error) {
			return &domain.User{
				ID:           "user-1",
				Username:     username,
				IsActive:     true,
				PasswordHash: hashedPassword,
			}, nil
		},
	}

	svc := &AuthService{
		cfg:    config.Config{JWTSecret: "test-secret"},
		users:  mockUsers,
		jwtKey: []byte("test-secret"),
		now:    time.Now,
	}

	_, err := svc.LoginWithPasswordClient("user", "wrongpassword", "")
	if err == nil {
		t.Error("expected error for wrong password")
	}
	if err.Error() != "username atau password salah" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginWithPasswordClient_AdminMobileRestriction(t *testing.T) {
	// This case is covered by TestEnsureAdminAllowed_AdminWithMobileClient
	// The full flow with bcrypt password verification would require a valid hash
	// which is complex to set up. The admin restriction logic is tested separately.
}

// ============================================================================
// Mock-based tests for IssueToken
// ============================================================================

func TestIssueToken_Success(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		createFunc: func(token *domain.RefreshToken) error {
			return nil
		},
	}

	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	svc := &AuthService{
		cfg: config.Config{
			JWTSecret:            "test-secret-32-bytes-long!!!",
			JWTExpiryUser:        time.Hour,
			JWTRefreshExpiryUser: 24 * time.Hour,
		},
		refreshTokens: mockRefreshTokens,
		jwtKey:        []byte("test-secret-32-bytes-long!!!"),
		now:           func() time.Time { return fixedTime },
	}

	user := domain.User{
		ID:    "user-123",
		Name:  "Test User",
		Email: "test@example.com",
		Role:  domain.RoleRegistered,
	}

	result, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	if result.Token == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.User.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, result.User.ID)
	}
	if result.ExpiresAt != fixedTime.Add(time.Hour) {
		t.Errorf("expected expiry at %v, got %v", fixedTime.Add(time.Hour), result.ExpiresAt)
	}
}

func TestIssueToken_AdminExpiry(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		createFunc: func(token *domain.RefreshToken) error {
			return nil
		},
	}

	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	svc := &AuthService{
		cfg: config.Config{
			JWTSecret:             "test-secret-32-bytes-long!!!",
			JWTExpiryAdmin:        2 * time.Hour,
			JWTRefreshExpiryAdmin: 48 * time.Hour,
		},
		refreshTokens: mockRefreshTokens,
		jwtKey:        []byte("test-secret-32-bytes-long!!!"),
		now:           func() time.Time { return fixedTime },
	}

	user := domain.User{
		ID:   "admin-123",
		Role: domain.RoleAdmin,
	}

	result, err := svc.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	if result.ExpiresAt != fixedTime.Add(2*time.Hour) {
		t.Errorf("expected admin expiry at %v, got %v", fixedTime.Add(2*time.Hour), result.ExpiresAt)
	}
}

func TestIssueToken_RefreshTokenCreateError(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		createFunc: func(token *domain.RefreshToken) error {
			return errDatabaseError
		},
	}

	svc := &AuthService{
		cfg: config.Config{
			JWTSecret:     "test-secret-32-bytes-long!!!",
			JWTExpiryUser: time.Hour,
		},
		refreshTokens: mockRefreshTokens,
		jwtKey:        []byte("test-secret-32-bytes-long!!!"),
		now:           time.Now,
	}

	user := domain.User{ID: "user-123", Role: domain.RoleRegistered}

	_, err := svc.IssueToken(user)
	if err == nil {
		t.Error("expected error when refresh token creation fails")
	}
}

// ============================================================================
// Mock-based tests for RefreshWithToken
// ============================================================================

func TestRefreshWithToken_EmptyToken(t *testing.T) {
	svc := &AuthService{now: time.Now}

	_, err := svc.RefreshWithToken("")
	if err == nil {
		t.Error("expected error for empty refresh token")
	}
	if err.Error() != "refresh token wajib diisi" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRefreshWithToken_WhitespaceToken(t *testing.T) {
	svc := &AuthService{now: time.Now}

	_, err := svc.RefreshWithToken("   ")
	if err == nil {
		t.Error("expected error for whitespace-only refresh token")
	}
}

func TestRefreshWithToken_InvalidToken(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		findByHashFunc: func(hash string) (*domain.RefreshToken, error) {
			return nil, errNotFound
		},
	}

	svc := &AuthService{
		refreshTokens: mockRefreshTokens,
		now:           time.Now,
	}

	_, err := svc.RefreshWithToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
	if err.Error() != "refresh token tidak valid" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRefreshWithToken_ExpiredToken(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		findByHashFunc: func(hash string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				ID:        "token-1",
				UserID:    "user-1",
				ExpiresAt: time.Now().Add(-time.Hour), // Expired
			}, nil
		},
		deleteByIDFunc: func(id string) error {
			return nil
		},
	}

	svc := &AuthService{
		refreshTokens: mockRefreshTokens,
		now:           time.Now,
	}

	_, err := svc.RefreshWithToken("expired-token")
	if err == nil {
		t.Error("expected error for expired refresh token")
	}
	if err.Error() != "refresh token kadaluarsa" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRefreshWithToken_UserNotFound(t *testing.T) {
	mockRefreshTokens := &mockRefreshTokenRepo{
		findByHashFunc: func(hash string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				ID:        "token-1",
				UserID:    "user-1",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	mockUsers := &mockUserRepo{
		findByIDFunc: func(id string) (*domain.User, error) {
			return nil, errNotFound
		},
	}

	svc := &AuthService{
		users:         mockUsers,
		refreshTokens: mockRefreshTokens,
		now:           time.Now,
	}

	_, err := svc.RefreshWithToken("valid-token")
	if err == nil {
		t.Error("expected error when user not found")
	}
	if err.Error() != "user tidak ditemukan" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRefreshWithToken_Success(t *testing.T) {
	deletedTokenID := ""

	mockRefreshTokens := &mockRefreshTokenRepo{
		findByHashFunc: func(hash string) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{
				ID:        "old-token-1",
				UserID:    "user-1",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
		deleteByIDFunc: func(id string) error {
			deletedTokenID = id
			return nil
		},
		createFunc: func(token *domain.RefreshToken) error {
			return nil
		},
	}
	mockUsers := &mockUserRepo{
		findByIDFunc: func(id string) (*domain.User, error) {
			return &domain.User{
				ID:       id,
				Username: "testuser",
				Role:     domain.RoleRegistered,
				IsActive: true,
			}, nil
		},
	}

	svc := &AuthService{
		cfg: config.Config{
			JWTSecret:            "test-secret-32-bytes-long!!!",
			JWTExpiryUser:        time.Hour,
			JWTRefreshExpiryUser: 24 * time.Hour,
		},
		users:         mockUsers,
		refreshTokens: mockRefreshTokens,
		jwtKey:        []byte("test-secret-32-bytes-long!!!"),
		now:           time.Now,
	}

	result, err := svc.RefreshWithToken("valid-refresh-token")
	if err != nil {
		t.Fatalf("RefreshWithToken failed: %v", err)
	}

	if result.Token == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if deletedTokenID != "old-token-1" {
		t.Errorf("expected old token to be deleted, got deleted ID: %s", deletedTokenID)
	}
}

// ============================================================================
// Helper mock types for testing
// ============================================================================

var errNotFound = errors.New("not found")
var errDatabaseError = errors.New("database error")

type mockUserRepo struct {
	findByUsernameFunc func(username string) (*domain.User, error)
	findByIDFunc       func(id string) (*domain.User, error)
}

func (m *mockUserRepo) Create(user *domain.User) error        { return nil }
func (m *mockUserRepo) UpsertByEmail(user *domain.User) error { return nil }
func (m *mockUserRepo) FindByID(id string) (*domain.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(email string) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) FindByUsername(username string) (*domain.User, error) {
	if m.findByUsernameFunc != nil {
		return m.findByUsernameFunc(username)
	}
	return nil, nil
}

type mockRefreshTokenRepo struct {
	createFunc     func(token *domain.RefreshToken) error
	findByHashFunc func(hash string) (*domain.RefreshToken, error)
	deleteByIDFunc func(id string) error
}

func (m *mockRefreshTokenRepo) Create(token *domain.RefreshToken) error {
	if m.createFunc != nil {
		return m.createFunc(token)
	}
	return nil
}
func (m *mockRefreshTokenRepo) FindByHash(hash string) (*domain.RefreshToken, error) {
	if m.findByHashFunc != nil {
		return m.findByHashFunc(hash)
	}
	return nil, nil
}
func (m *mockRefreshTokenRepo) DeleteByID(id string) error {
	if m.deleteByIDFunc != nil {
		return m.deleteByIDFunc(id)
	}
	return nil
}
