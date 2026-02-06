package mocks

import (
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
)

// Compile-time interface implementation checks
var _ repository.UserRepositoryInterface = (*MockUserRepository)(nil)
var _ repository.TicketRepositoryInterface = (*MockTicketRepository)(nil)
var _ repository.SurveyRepositoryInterface = (*MockSurveyRepository)(nil)
var _ repository.CategoryRepositoryInterface = (*MockCategoryRepository)(nil)
var _ repository.NotificationRepositoryInterface = (*MockNotificationRepository)(nil)
var _ repository.FCMTokenRepositoryInterface = (*MockFCMTokenRepository)(nil)
var _ repository.RefreshTokenRepositoryInterface = (*MockRefreshTokenRepository)(nil)
var _ repository.AttachmentRepositoryInterface = (*MockAttachmentRepository)(nil)

// ============================================================================
// Mock User Repository
// ============================================================================

type MockUserRepository struct {
	CreateFunc         func(user *domain.User) error
	UpsertByEmailFunc  func(user *domain.User) error
	FindByIDFunc       func(id string) (*domain.User, error)
	FindByEmailFunc    func(email string) (*domain.User, error)
	FindByUsernameFunc func(username string) (*domain.User, error)
}

func (m *MockUserRepository) Create(user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) UpsertByEmail(user *domain.User) error {
	if m.UpsertByEmailFunc != nil {
		return m.UpsertByEmailFunc(user)
	}
	return nil
}

func (m *MockUserRepository) FindByID(id string) (*domain.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*domain.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(email)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByUsername(username string) (*domain.User, error) {
	if m.FindByUsernameFunc != nil {
		return m.FindByUsernameFunc(username)
	}
	return nil, nil
}

// ============================================================================
// Mock Ticket Repository
// ============================================================================

type MockTicketRepository struct {
	CreateFunc          func(ticket *domain.Ticket) error
	UpdateFunc          func(ticket *domain.Ticket) error
	SoftDeleteFunc      func(ticketID string) error
	FindByIDFunc        func(ticketID string) (*domain.Ticket, error)
	ListByUserFunc      func(userID string) ([]domain.Ticket, error)
	ListAllFunc         func() ([]domain.Ticket, error)
	SearchFunc          func(query string, isGuest bool) ([]domain.Ticket, error)
	ListFilteredFunc    func(filter repository.TicketListFilter, page int, limit int) ([]domain.Ticket, int64, error)
	CountForYearFunc    func(year int) (int64, error)
	AddHistoryFunc      func(history *domain.TicketHistory) error
	AddCommentFunc      func(comment *domain.TicketComment) error
	UpdateStatusFunc    func(ticketID string, status domain.TicketStatus, surveyRequired bool) error
	GetSurveyScoresFunc func(ticketIDs []string) (map[string]float64, error)
}

func (m *MockTicketRepository) Create(ticket *domain.Ticket) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ticket)
	}
	return nil
}

func (m *MockTicketRepository) Update(ticket *domain.Ticket) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ticket)
	}
	return nil
}

func (m *MockTicketRepository) SoftDelete(ticketID string) error {
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ticketID)
	}
	return nil
}

func (m *MockTicketRepository) FindByID(ticketID string) (*domain.Ticket, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ticketID)
	}
	return nil, nil
}

func (m *MockTicketRepository) ListByUser(userID string) ([]domain.Ticket, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(userID)
	}
	return nil, nil
}

func (m *MockTicketRepository) ListAll() ([]domain.Ticket, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc()
	}
	return nil, nil
}

func (m *MockTicketRepository) Search(query string, isGuest bool) ([]domain.Ticket, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(query, isGuest)
	}
	return nil, nil
}

func (m *MockTicketRepository) ListFiltered(filter repository.TicketListFilter, page int, limit int) ([]domain.Ticket, int64, error) {
	if m.ListFilteredFunc != nil {
		return m.ListFilteredFunc(filter, page, limit)
	}
	return nil, 0, nil
}

func (m *MockTicketRepository) CountForYear(year int) (int64, error) {
	if m.CountForYearFunc != nil {
		return m.CountForYearFunc(year)
	}
	return 0, nil
}

func (m *MockTicketRepository) AddHistory(history *domain.TicketHistory) error {
	if m.AddHistoryFunc != nil {
		return m.AddHistoryFunc(history)
	}
	return nil
}

func (m *MockTicketRepository) AddComment(comment *domain.TicketComment) error {
	if m.AddCommentFunc != nil {
		return m.AddCommentFunc(comment)
	}
	return nil
}

func (m *MockTicketRepository) UpdateStatus(ticketID string, status domain.TicketStatus, surveyRequired bool) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ticketID, status, surveyRequired)
	}
	return nil
}

func (m *MockTicketRepository) GetSurveyScores(ticketIDs []string) (map[string]float64, error) {
	if m.GetSurveyScoresFunc != nil {
		return m.GetSurveyScoresFunc(ticketIDs)
	}
	return make(map[string]float64), nil
}

// ============================================================================
// Mock Category Repository
// ============================================================================

type MockCategoryRepository struct {
	ListFunc           func() ([]domain.ServiceCategory, error)
	FindByIDFunc       func(id string) (*domain.ServiceCategory, error)
	FindByNameFunc     func(name string) (*domain.ServiceCategory, error)
	UpsertFunc         func(category domain.ServiceCategory) error
	UpdateTemplateFunc func(categoryID string, templateID string) error
}

func (m *MockCategoryRepository) List() ([]domain.ServiceCategory, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

func (m *MockCategoryRepository) FindByID(id string) (*domain.ServiceCategory, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, nil
}

func (m *MockCategoryRepository) FindByName(name string) (*domain.ServiceCategory, error) {
	if m.FindByNameFunc != nil {
		return m.FindByNameFunc(name)
	}
	return nil, nil
}

func (m *MockCategoryRepository) Upsert(category domain.ServiceCategory) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(category)
	}
	return nil
}

func (m *MockCategoryRepository) UpdateTemplate(categoryID string, templateID string) error {
	if m.UpdateTemplateFunc != nil {
		return m.UpdateTemplateFunc(categoryID, templateID)
	}
	return nil
}

// ============================================================================
// Mock Survey Repository
// ============================================================================

type MockSurveyRepository struct {
	ListTemplatesFunc   func() ([]domain.SurveyTemplate, error)
	FindByCategoryFunc  func(categoryID string) (*domain.SurveyTemplate, error)
	FindByIDFunc        func(templateID string) (*domain.SurveyTemplate, error)
	CreateTemplateFunc  func(template *domain.SurveyTemplate) error
	UpdateTemplateFunc  func(template *domain.SurveyTemplate) error
	ReplaceTemplateFunc func(template *domain.SurveyTemplate) error
	DeleteTemplateFunc  func(templateID string) error
	SaveResponseFunc    func(response *domain.SurveyResponse) error
	HasResponseFunc     func(ticketID string, userID string) (bool, error)
	ListResponsesFunc   func(filter repository.SurveyResponseFilter, page int, limit int) ([]repository.SurveyResponseRow, int64, error)
}

func (m *MockSurveyRepository) ListTemplates() ([]domain.SurveyTemplate, error) {
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc()
	}
	return nil, nil
}

func (m *MockSurveyRepository) FindByCategory(categoryID string) (*domain.SurveyTemplate, error) {
	if m.FindByCategoryFunc != nil {
		return m.FindByCategoryFunc(categoryID)
	}
	return nil, nil
}

func (m *MockSurveyRepository) FindByID(templateID string) (*domain.SurveyTemplate, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(templateID)
	}
	return nil, nil
}

func (m *MockSurveyRepository) CreateTemplate(template *domain.SurveyTemplate) error {
	if m.CreateTemplateFunc != nil {
		return m.CreateTemplateFunc(template)
	}
	return nil
}

func (m *MockSurveyRepository) UpdateTemplate(template *domain.SurveyTemplate) error {
	if m.UpdateTemplateFunc != nil {
		return m.UpdateTemplateFunc(template)
	}
	return nil
}

func (m *MockSurveyRepository) ReplaceTemplate(template *domain.SurveyTemplate) error {
	if m.ReplaceTemplateFunc != nil {
		return m.ReplaceTemplateFunc(template)
	}
	return nil
}

func (m *MockSurveyRepository) DeleteTemplate(templateID string) error {
	if m.DeleteTemplateFunc != nil {
		return m.DeleteTemplateFunc(templateID)
	}
	return nil
}

func (m *MockSurveyRepository) SaveResponse(response *domain.SurveyResponse) error {
	if m.SaveResponseFunc != nil {
		return m.SaveResponseFunc(response)
	}
	return nil
}

func (m *MockSurveyRepository) HasResponse(ticketID string, userID string) (bool, error) {
	if m.HasResponseFunc != nil {
		return m.HasResponseFunc(ticketID, userID)
	}
	return false, nil
}

func (m *MockSurveyRepository) ListResponses(filter repository.SurveyResponseFilter, page int, limit int) ([]repository.SurveyResponseRow, int64, error) {
	if m.ListResponsesFunc != nil {
		return m.ListResponsesFunc(filter, page, limit)
	}
	return nil, 0, nil
}

// ============================================================================
// Mock Notification Repository
// ============================================================================

type MockNotificationRepository struct {
	ListByUserFunc func(userID string) ([]domain.Notification, error)
	CreateFunc     func(notification *domain.Notification) error
	MarkReadFunc   func(notificationID string, userID string) error
}

func (m *MockNotificationRepository) ListByUser(userID string) ([]domain.Notification, error) {
	if m.ListByUserFunc != nil {
		return m.ListByUserFunc(userID)
	}
	return nil, nil
}

func (m *MockNotificationRepository) Create(notification *domain.Notification) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(notification)
	}
	return nil
}

func (m *MockNotificationRepository) MarkRead(notificationID string, userID string) error {
	if m.MarkReadFunc != nil {
		return m.MarkReadFunc(notificationID, userID)
	}
	return nil
}

// ============================================================================
// Mock FCM Token Repository
// ============================================================================

type MockFCMTokenRepository struct {
	UpsertFunc     func(token *domain.FCMToken) error
	ListTokensFunc func(userID string) ([]domain.FCMToken, error)
}

func (m *MockFCMTokenRepository) Upsert(token *domain.FCMToken) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(token)
	}
	return nil
}

func (m *MockFCMTokenRepository) ListTokens(userID string) ([]domain.FCMToken, error) {
	if m.ListTokensFunc != nil {
		return m.ListTokensFunc(userID)
	}
	return nil, nil
}

// ============================================================================
// Mock Refresh Token Repository
// ============================================================================

type MockRefreshTokenRepository struct {
	CreateFunc     func(token *domain.RefreshToken) error
	FindByHashFunc func(hash string) (*domain.RefreshToken, error)
	DeleteByIDFunc func(id string) error
}

func (m *MockRefreshTokenRepository) Create(token *domain.RefreshToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(token)
	}
	return nil
}

func (m *MockRefreshTokenRepository) FindByHash(hash string) (*domain.RefreshToken, error) {
	if m.FindByHashFunc != nil {
		return m.FindByHashFunc(hash)
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) DeleteByID(id string) error {
	if m.DeleteByIDFunc != nil {
		return m.DeleteByIDFunc(id)
	}
	return nil
}

// ============================================================================
// Mock Attachment Repository
// ============================================================================

type MockAttachmentRepository struct {
	CreateFunc         func(attachment *domain.Attachment) error
	FindByIDFunc       func(id string) (*domain.Attachment, error)
	AttachToTicketFunc func(ids []string, ticketID string) error
}

func (m *MockAttachmentRepository) Create(attachment *domain.Attachment) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(attachment)
	}
	return nil
}

func (m *MockAttachmentRepository) FindByID(id string) (*domain.Attachment, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, nil
}

func (m *MockAttachmentRepository) AttachToTicket(ids []string, ticketID string) error {
	if m.AttachToTicketFunc != nil {
		return m.AttachToTicketFunc(ids, ticketID)
	}
	return nil
}

// ============================================================================
// Mock FCM Client
// ============================================================================

type MockFCMClient struct {
	SendToTokensFunc func(tokens []string, title string, body string, data map[string]string) error
}

func (m *MockFCMClient) SendToTokens(tokens []string, title string, body string, data map[string]string) error {
	if m.SendToTokensFunc != nil {
		return m.SendToTokensFunc(tokens, title, body, data)
	}
	return nil
}

// ============================================================================
// Helper: Fixed time function for testing
// ============================================================================

func FixedTime(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}
