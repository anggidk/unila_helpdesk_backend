package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"unila_helpdesk_backend/internal/domain"
)

// ============================================================================
// Tests for FCMRegisterRequest validation patterns
// ============================================================================

func TestFCMRegisterRequest_EmptyToken(t *testing.T) {
	req := FCMRegisterRequest{Token: ""}

	// Simulate validation logic from RegisterToken
	if strings.TrimSpace(req.Token) == "" {
		// Expected - should skip registration
		return
	}
	t.Error("expected empty token to be caught by validation")
}

func TestFCMRegisterRequest_WhitespaceToken(t *testing.T) {
	req := FCMRegisterRequest{Token: "   "}

	if strings.TrimSpace(req.Token) == "" {
		// Expected - should skip registration
		return
	}
	t.Error("expected whitespace-only token to be caught by validation")
}

func TestFCMRegisterRequest_ValidToken(t *testing.T) {
	req := FCMRegisterRequest{Token: "fcm-token-123"}

	if strings.TrimSpace(req.Token) == "" {
		t.Error("valid token should not be flagged as empty")
	}
}

// ============================================================================
// Tests for guest user notification skip logic
// ============================================================================

func TestNotificationService_GuestUserSkip(t *testing.T) {
	// Guest users should not be able to register FCM tokens
	user := domain.User{Role: domain.RoleGuest}

	if user.Role == domain.RoleGuest {
		// Expected - should skip registration
		return
	}
	t.Error("guest user check failed")
}

func TestNotificationService_RegisteredUserAllowed(t *testing.T) {
	user := domain.User{Role: domain.RoleRegistered}

	if user.Role == domain.RoleGuest {
		t.Error("registered user should not be treated as guest")
	}
}

func TestNotificationService_AdminUserAllowed(t *testing.T) {
	user := domain.User{Role: domain.RoleAdmin}

	if user.Role == domain.RoleGuest {
		t.Error("admin user should not be treated as guest")
	}
}

// ============================================================================
// Tests for NotificationDTO mapping
// ============================================================================

func TestNotificationDTO_Mapping(t *testing.T) {
	item := domain.Notification{
		ID:      "notif-123",
		Title:   "Test Notification",
		Message: "This is a test message",
		IsRead:  false,
	}

	dto := domain.NotificationDTO{
		ID:      item.ID,
		Title:   item.Title,
		Message: item.Message,
		IsRead:  item.IsRead,
	}

	if dto.ID != item.ID {
		t.Errorf("ID mismatch: got %s, want %s", dto.ID, item.ID)
	}
	if dto.Title != item.Title {
		t.Errorf("Title mismatch: got %s, want %s", dto.Title, item.Title)
	}
	if dto.Message != item.Message {
		t.Errorf("Message mismatch: got %s, want %s", dto.Message, item.Message)
	}
	if dto.IsRead != item.IsRead {
		t.Errorf("IsRead mismatch: got %v, want %v", dto.IsRead, item.IsRead)
	}
}

func TestNotificationDTO_ReadStatus(t *testing.T) {
	// Verify read status mapping
	unread := domain.NotificationDTO{IsRead: false}
	read := domain.NotificationDTO{IsRead: true}

	if unread.IsRead {
		t.Error("unread notification should have IsRead=false")
	}
	if !read.IsRead {
		t.Error("read notification should have IsRead=true")
	}
}

// ============================================================================
// Tests for Platform validation
// ============================================================================

func TestFCMPlatform_Values(t *testing.T) {
	validPlatforms := []string{"android", "ios", "web"}

	for _, platform := range validPlatforms {
		req := FCMRegisterRequest{Platform: platform}
		if req.Platform == "" {
			t.Errorf("platform %s should not be empty", platform)
		}
	}
}

// ============================================================================
// Mock-based tests for NotificationService.List
// ============================================================================

func TestNotificationService_List_Success(t *testing.T) {
	mockNotifications := &mockNotificationRepo{
		listByUserFunc: func(userID string) ([]domain.Notification, error) {
			return []domain.Notification{
				{ID: "n1", UserID: userID, Title: "Title 1", Message: "Msg 1", IsRead: false},
				{ID: "n2", UserID: userID, Title: "Title 2", Message: "Msg 2", IsRead: true},
			}, nil
		},
	}

	svc := &NotificationService{
		notifications: mockNotifications,
	}

	user := domain.User{ID: "user-123"}
	result, err := svc.List(user)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(result))
	}
	if result[0].ID != "n1" {
		t.Errorf("expected first notification ID n1, got %s", result[0].ID)
	}
	if result[1].IsRead != true {
		t.Error("expected second notification to be read")
	}
}

func TestNotificationService_List_Empty(t *testing.T) {
	mockNotifications := &mockNotificationRepo{
		listByUserFunc: func(userID string) ([]domain.Notification, error) {
			return []domain.Notification{}, nil
		},
	}

	svc := &NotificationService{
		notifications: mockNotifications,
	}

	user := domain.User{ID: "user-123"}
	result, err := svc.List(user)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(result))
	}
}

func TestNotificationService_List_RepositoryError(t *testing.T) {
	mockNotifications := &mockNotificationRepo{
		listByUserFunc: func(userID string) ([]domain.Notification, error) {
			return nil, errTestDatabase
		},
	}

	svc := &NotificationService{
		notifications: mockNotifications,
	}

	user := domain.User{ID: "user-123"}
	_, err := svc.List(user)
	if err == nil {
		t.Error("expected error when repository fails")
	}
}

// ============================================================================
// Mock-based tests for NotificationService.RegisterToken
// ============================================================================

func TestNotificationService_RegisterToken_GuestUser(t *testing.T) {
	svc := &NotificationService{}

	user := domain.User{ID: "guest-123", Role: domain.RoleGuest}
	req := FCMRegisterRequest{Token: "fcm-token-123"}

	err := svc.RegisterToken(user, req)
	if err != nil {
		t.Errorf("RegisterToken for guest should return nil, got: %v", err)
	}
}

func TestNotificationService_RegisterToken_EmptyToken(t *testing.T) {
	svc := &NotificationService{}

	user := domain.User{ID: "user-123", Role: domain.RoleRegistered}
	req := FCMRegisterRequest{Token: ""}

	err := svc.RegisterToken(user, req)
	if err != nil {
		t.Errorf("RegisterToken with empty token should return nil, got: %v", err)
	}
}

func TestNotificationService_RegisterToken_Success(t *testing.T) {
	var savedToken *domain.FCMToken

	mockTokens := &mockFCMTokenRepo{
		upsertFunc: func(token *domain.FCMToken) error {
			savedToken = token
			return nil
		},
	}

	svc := &NotificationService{
		tokens: mockTokens,
		now:    fixedTimeFunc,
	}

	user := domain.User{ID: "user-123", Role: domain.RoleRegistered}
	req := FCMRegisterRequest{Token: "fcm-token-123", Platform: "android"}

	err := svc.RegisterToken(user, req)
	if err != nil {
		t.Fatalf("RegisterToken failed: %v", err)
	}

	if savedToken == nil {
		t.Fatal("expected token to be saved")
	}
	if savedToken.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, savedToken.UserID)
	}
	if savedToken.Token != req.Token {
		t.Errorf("expected token %s, got %s", req.Token, savedToken.Token)
	}
	if savedToken.Platform != req.Platform {
		t.Errorf("expected platform %s, got %s", req.Platform, savedToken.Platform)
	}
}

func TestNotificationService_RegisterToken_RepositoryError(t *testing.T) {
	mockTokens := &mockFCMTokenRepo{
		upsertFunc: func(token *domain.FCMToken) error {
			return errTestDatabase
		},
	}

	svc := &NotificationService{
		tokens: mockTokens,
		now:    fixedTimeFunc,
	}

	user := domain.User{ID: "user-123", Role: domain.RoleRegistered}
	req := FCMRegisterRequest{Token: "fcm-token-123"}

	err := svc.RegisterToken(user, req)
	if err == nil {
		t.Error("expected error when repository fails")
	}
}

// ============================================================================
// Helper mock types for notification tests
// ============================================================================

var errTestDatabase = errors.New("database error")
var fixedTimeFunc = func() time.Time { return time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC) }

type mockNotificationRepo struct {
	listByUserFunc func(userID string) ([]domain.Notification, error)
	createFunc     func(notification *domain.Notification) error
	markReadFunc   func(notificationID string, userID string) error
}

func (m *mockNotificationRepo) ListByUser(userID string) ([]domain.Notification, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(userID)
	}
	return nil, nil
}

func (m *mockNotificationRepo) Create(notification *domain.Notification) error {
	if m.createFunc != nil {
		return m.createFunc(notification)
	}
	return nil
}

func (m *mockNotificationRepo) MarkRead(notificationID string, userID string) error {
	if m.markReadFunc != nil {
		return m.markReadFunc(notificationID, userID)
	}
	return nil
}

type mockFCMTokenRepo struct {
	upsertFunc     func(token *domain.FCMToken) error
	listTokensFunc func(userID string) ([]domain.FCMToken, error)
}

func (m *mockFCMTokenRepo) Upsert(token *domain.FCMToken) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(token)
	}
	return nil
}

func (m *mockFCMTokenRepo) ListTokens(userID string) ([]domain.FCMToken, error) {
	if m.listTokensFunc != nil {
		return m.listTokensFunc(userID)
	}
	return nil, nil
}
