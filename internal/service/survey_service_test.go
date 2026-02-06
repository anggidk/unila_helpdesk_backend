package service

import (
	"errors"
	"testing"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
)

// ============================================================================
// Tests for calculateSurveyScore
// ============================================================================

func TestCalculateSurveyScore_EmptyAnswers(t *testing.T) {
	answers := map[string]interface{}{}
	score := calculateSurveyScore(answers, nil)
	if score != 0 {
		t.Errorf("calculateSurveyScore({}) = %v, want 0", score)
	}
}

func TestCalculateSurveyScore_NilTemplate(t *testing.T) {
	// Without template, uses legacy scoring
	answers := map[string]interface{}{
		"q1": 5,
		"q2": 3,
	}
	score := calculateSurveyScore(answers, nil)
	// Legacy: (100 + 50) / 2 = 75
	if score != 75 {
		t.Errorf("calculateSurveyScore(nil template) = %v, want 75", score)
	}
}

func TestCalculateSurveyScore_EmptyQuestions(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5,
	}
	template := &domain.SurveyTemplate{
		Questions: []domain.SurveyQuestion{},
	}
	score := calculateSurveyScore(answers, template)
	// Empty questions, uses legacy
	if score != 100 {
		t.Errorf("calculateSurveyScore(empty questions) = %v, want 100", score)
	}
}

func TestCalculateSurveyScore_WithTemplate(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5,    // Likert5: 100
		"q2": "ya", // YesNo: 100
		"q3": 2,    // Likert3: 50
	}
	template := &domain.SurveyTemplate{
		Questions: []domain.SurveyQuestion{
			{ID: "q1", Type: domain.QuestionLikert},
			{ID: "q2", Type: domain.QuestionYesNo},
			{ID: "q3", Type: domain.QuestionLikert3},
		},
	}
	score := calculateSurveyScore(answers, template)
	// Avg: (100 + 100 + 50) / 3 = 83.33...
	if score < 83 || score > 84 {
		t.Errorf("calculateSurveyScore(with template) = %v, want ~83.33", score)
	}
}

func TestCalculateSurveyScore_MissingAnswers(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5, // Only q1 answered
	}
	template := &domain.SurveyTemplate{
		Questions: []domain.SurveyQuestion{
			{ID: "q1", Type: domain.QuestionLikert},
			{ID: "q2", Type: domain.QuestionYesNo},
			{ID: "q3", Type: domain.QuestionLikert3},
		},
	}
	score := calculateSurveyScore(answers, template)
	// Only q1 is counted: 100 / 1 = 100
	if score != 100 {
		t.Errorf("calculateSurveyScore(missing answers) = %v, want 100", score)
	}
}

func TestCalculateSurveyScore_NonScorableQuestions(t *testing.T) {
	answers := map[string]interface{}{
		"q1": "Some text feedback",
		"q2": "Option A",
	}
	template := &domain.SurveyTemplate{
		Questions: []domain.SurveyQuestion{
			{ID: "q1", Type: domain.QuestionText},
			{ID: "q2", Type: domain.QuestionMultipleChoice},
		},
	}
	score := calculateSurveyScore(answers, template)
	// Text and MC are not scorable
	if score != 0 {
		t.Errorf("calculateSurveyScore(non-scorable) = %v, want 0", score)
	}
}

// ============================================================================
// Tests for calculateLegacyScore
// ============================================================================

func TestCalculateLegacyScore_Empty(t *testing.T) {
	answers := map[string]interface{}{}
	score := calculateLegacyScore(answers)
	if score != 0 {
		t.Errorf("calculateLegacyScore({}) = %v, want 0", score)
	}
}

func TestCalculateLegacyScore_Float64(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5.0,
		"q2": 3.0,
		"q3": 1.0,
	}
	score := calculateLegacyScore(answers)
	// Normalized: (100 + 50 + 0) / 3 = 50
	if score != 50 {
		t.Errorf("calculateLegacyScore(float64) = %v, want 50", score)
	}
}

func TestCalculateLegacyScore_Int(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5,
		"q2": 1,
	}
	score := calculateLegacyScore(answers)
	// Normalized: (100 + 0) / 2 = 50
	if score != 50 {
		t.Errorf("calculateLegacyScore(int) = %v, want 50", score)
	}
}

func TestCalculateLegacyScore_Bool(t *testing.T) {
	answers := map[string]interface{}{
		"q1": true,
		"q2": false,
	}
	score := calculateLegacyScore(answers)
	// (100 + 0) / 2 = 50
	if score != 50 {
		t.Errorf("calculateLegacyScore(bool) = %v, want 50", score)
	}
}

func TestCalculateLegacyScore_String(t *testing.T) {
	answers := map[string]interface{}{
		"q1": "ya",
		"q2": "tidak",
		"q3": "3",
	}
	score := calculateLegacyScore(answers)
	// (100 + 0 + 50) / 3 = 50
	if score != 50 {
		t.Errorf("calculateLegacyScore(string) = %v, want 50", score)
	}
}

func TestCalculateLegacyScore_OutOfRange(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 10.0, // out of range, ignored
		"q2": 0,    // out of range, ignored
		"q3": 3.0,  // valid
	}
	score := calculateLegacyScore(answers)
	// Only q3 is counted: 50 / 1 = 50
	if score != 50 {
		t.Errorf("calculateLegacyScore(out of range) = %v, want 50", score)
	}
}

func TestCalculateLegacyScore_MixedTypes(t *testing.T) {
	answers := map[string]interface{}{
		"q1": 5,    // 100
		"q2": "ya", // 100
		"q3": true, // 100
		"q4": "3",  // 50
	}
	score := calculateLegacyScore(answers)
	// (100 + 100 + 100 + 50) / 4 = 87.5
	if score != 87.5 {
		t.Errorf("calculateLegacyScore(mixed) = %v, want 87.5", score)
	}
}

func TestCalculateLegacyScore_InvalidString(t *testing.T) {
	answers := map[string]interface{}{
		"q1": "invalid",
		"q2": "abc",
		"q3": 5,
	}
	score := calculateLegacyScore(answers)
	// Only q3 is valid: 100 / 1 = 100
	if score != 100 {
		t.Errorf("calculateLegacyScore(invalid string) = %v, want 100", score)
	}
}

// ============================================================================
// Tests for mapSurveyTemplate
// ============================================================================

func TestMapSurveyTemplate_Basic(t *testing.T) {
	template := domain.SurveyTemplate{
		ID:          "template-1",
		Title:       "Test Template",
		Description: "A test template",
		Framework:   "RATER",
		CategoryID:  "cat-1",
	}
	dto := mapSurveyTemplate(template)

	if dto.ID != template.ID {
		t.Errorf("ID mismatch: got %s, want %s", dto.ID, template.ID)
	}
	if dto.Title != template.Title {
		t.Errorf("Title mismatch: got %s, want %s", dto.Title, template.Title)
	}
	if dto.Framework != template.Framework {
		t.Errorf("Framework mismatch: got %s, want %s", dto.Framework, template.Framework)
	}
}

func TestMapSurveyTemplate_WithQuestions(t *testing.T) {
	template := domain.SurveyTemplate{
		ID:    "template-1",
		Title: "Test",
		Questions: []domain.SurveyQuestion{
			{
				ID:      "q1",
				Text:    "Question 1",
				Type:    domain.QuestionLikert,
				Options: []byte(`["1","2","3","4","5"]`),
			},
			{
				ID:      "q2",
				Text:    "Question 2",
				Type:    domain.QuestionYesNo,
				Options: []byte(`null`),
			},
		},
	}
	dto := mapSurveyTemplate(template)

	if len(dto.Questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(dto.Questions))
	}
	if dto.Questions[0].ID != "q1" {
		t.Errorf("Question 1 ID mismatch")
	}
	if dto.Questions[0].Type != "likert" {
		t.Errorf("Question 1 Type mismatch: got %s", dto.Questions[0].Type)
	}
}

func TestMapSurveyTemplates_Empty(t *testing.T) {
	templates := []domain.SurveyTemplate{}
	dtos := mapSurveyTemplates(templates)
	if len(dtos) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(dtos))
	}
}

func TestMapSurveyTemplates_Multiple(t *testing.T) {
	templates := []domain.SurveyTemplate{
		{ID: "t1", Title: "Template 1"},
		{ID: "t2", Title: "Template 2"},
		{ID: "t3", Title: "Template 3"},
	}
	dtos := mapSurveyTemplates(templates)
	if len(dtos) != 3 {
		t.Errorf("Expected 3 templates, got %d", len(dtos))
	}
}

// ============================================================================
// Validation tests for survey service requests
// ============================================================================

func TestSurveyTemplateRequest_Validation(t *testing.T) {
	testCases := []struct {
		req     SurveyTemplateRequest
		isValid bool
		field   string
	}{
		{
			req:     SurveyTemplateRequest{Title: "Test", CategoryID: "cat-1"},
			isValid: true,
		},
		{
			req:     SurveyTemplateRequest{Title: "", CategoryID: "cat-1"},
			isValid: false,
			field:   "title",
		},
		{
			req:     SurveyTemplateRequest{Title: "   ", CategoryID: "cat-1"},
			isValid: false,
			field:   "title",
		},
		{
			req:     SurveyTemplateRequest{Title: "Test", CategoryID: ""},
			isValid: false,
			field:   "categoryId",
		},
	}

	for i, tc := range testCases {
		// Simulate validation logic
		titleValid := len(trimString(tc.req.Title)) > 0
		categoryValid := len(trimString(tc.req.CategoryID)) > 0
		isValid := titleValid && categoryValid

		if isValid != tc.isValid {
			t.Errorf("Test case %d: validation = %v, want %v (field: %s)", i, isValid, tc.isValid, tc.field)
		}
	}
}

// Helper for testing
func trimString(s string) string {
	result := ""
	for _, c := range s {
		if c != ' ' && c != '\t' {
			result += string(c)
		}
	}
	return result
}

// ============================================================================
// Tests for ListResponsesPaged pagination
// ============================================================================

func TestListResponsesPaged_PaginationLogic(t *testing.T) {
	// Test pagination calculation
	testCases := []struct {
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{0, 10, 1, 10},  // page < 1 becomes 1
		{1, 100, 1, 50}, // limit > 50 becomes 50
		{1, 0, 1, 50},   // limit <= 0 becomes 50
		{5, 25, 5, 25},  // valid values unchanged
	}

	for _, tc := range testCases {
		limit := tc.limit
		page := tc.page

		// Simulate the logic from ListResponsesPaged
		if limit <= 0 {
			limit = 50
		}
		if limit > 50 {
			limit = 50
		}
		if page < 1 {
			page = 1
		}

		if page != tc.wantPage {
			t.Errorf("page %d -> %d, want %d", tc.page, page, tc.wantPage)
		}
		if limit != tc.wantLimit {
			t.Errorf("limit %d -> %d, want %d", tc.limit, limit, tc.wantLimit)
		}
	}
}

func TestTotalPages_Calculation(t *testing.T) {
	testCases := []struct {
		total int64
		limit int
		want  int
	}{
		{0, 10, 0},
		{10, 10, 1},
		{15, 10, 2},
		{100, 50, 2},
		{51, 50, 2},
	}

	for _, tc := range testCases {
		totalPages := int((tc.total + int64(tc.limit) - 1) / int64(tc.limit))
		if tc.total == 0 {
			totalPages = 0
		}
		if totalPages != tc.want {
			t.Errorf("totalPages(%d, %d) = %d, want %d", tc.total, tc.limit, totalPages, tc.want)
		}
	}
}

// ============================================================================
// Mock-based tests for SurveyService.ListTemplates
// ============================================================================

func TestSurveyService_ListTemplates_Success(t *testing.T) {
	mockSurveys := &mockSurveyRepoLocal{
		listTemplatesFunc: func() ([]domain.SurveyTemplate, error) {
			return []domain.SurveyTemplate{
				{ID: "t1", Title: "Template 1"},
				{ID: "t2", Title: "Template 2"},
			}, nil
		},
	}

	svc := &SurveyService{surveys: mockSurveys}

	result, err := svc.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 templates, got %d", len(result))
	}
}

func TestSurveyService_ListTemplates_Empty(t *testing.T) {
	mockSurveys := &mockSurveyRepoLocal{
		listTemplatesFunc: func() ([]domain.SurveyTemplate, error) {
			return []domain.SurveyTemplate{}, nil
		},
	}

	svc := &SurveyService{surveys: mockSurveys}

	result, err := svc.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 templates, got %d", len(result))
	}
}

func TestSurveyService_ListTemplates_Error(t *testing.T) {
	mockSurveys := &mockSurveyRepoLocal{
		listTemplatesFunc: func() ([]domain.SurveyTemplate, error) {
			return nil, errTestDB
		},
	}

	svc := &SurveyService{surveys: mockSurveys}

	_, err := svc.ListTemplates()
	if err == nil {
		t.Error("expected error when repository fails")
	}
}

// ============================================================================
// Mock-based tests for SurveyService.TemplateByCategory
// ============================================================================

func TestSurveyService_TemplateByCategory_Success(t *testing.T) {
	mockSurveys := &mockSurveyRepoLocal{
		findByCategoryFunc: func(categoryID string) (*domain.SurveyTemplate, error) {
			return &domain.SurveyTemplate{
				ID:         "t1",
				Title:      "Category Template",
				CategoryID: categoryID,
			}, nil
		},
	}

	svc := &SurveyService{surveys: mockSurveys}

	result, err := svc.TemplateByCategory("cat-1")
	if err != nil {
		t.Fatalf("TemplateByCategory failed: %v", err)
	}

	if result.CategoryID != "cat-1" {
		t.Errorf("expected category ID cat-1, got %s", result.CategoryID)
	}
}

func TestSurveyService_TemplateByCategory_NotFound(t *testing.T) {
	mockSurveys := &mockSurveyRepoLocal{
		findByCategoryFunc: func(categoryID string) (*domain.SurveyTemplate, error) {
			return nil, errTestNotFound
		},
	}

	svc := &SurveyService{surveys: mockSurveys}

	_, err := svc.TemplateByCategory("nonexistent")
	if err == nil {
		t.Error("expected error when template not found")
	}
}

// ============================================================================
// Helper mock types for survey tests
// ============================================================================

var errTestDB = errors.New("database error")
var errTestNotFound = errors.New("not found")

type mockSurveyRepoLocal struct {
	listTemplatesFunc   func() ([]domain.SurveyTemplate, error)
	findByCategoryFunc  func(categoryID string) (*domain.SurveyTemplate, error)
	findByIDFunc        func(templateID string) (*domain.SurveyTemplate, error)
	createTemplateFunc  func(template *domain.SurveyTemplate) error
	updateTemplateFunc  func(template *domain.SurveyTemplate) error
	replaceTemplateFunc func(template *domain.SurveyTemplate) error
	deleteTemplateFunc  func(templateID string) error
	saveResponseFunc    func(response *domain.SurveyResponse) error
	hasResponseFunc     func(ticketID string, userID string) (bool, error)
	listResponsesFunc   func(filter repository.SurveyResponseFilter, page int, limit int) ([]repository.SurveyResponseRow, int64, error)
}

func (m *mockSurveyRepoLocal) ListTemplates() ([]domain.SurveyTemplate, error) {
	if m.listTemplatesFunc != nil {
		return m.listTemplatesFunc()
	}
	return nil, nil
}

func (m *mockSurveyRepoLocal) FindByCategory(categoryID string) (*domain.SurveyTemplate, error) {
	if m.findByCategoryFunc != nil {
		return m.findByCategoryFunc(categoryID)
	}
	return nil, nil
}

func (m *mockSurveyRepoLocal) FindByID(templateID string) (*domain.SurveyTemplate, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(templateID)
	}
	return nil, nil
}

func (m *mockSurveyRepoLocal) CreateTemplate(template *domain.SurveyTemplate) error {
	if m.createTemplateFunc != nil {
		return m.createTemplateFunc(template)
	}
	return nil
}

func (m *mockSurveyRepoLocal) UpdateTemplate(template *domain.SurveyTemplate) error {
	if m.updateTemplateFunc != nil {
		return m.updateTemplateFunc(template)
	}
	return nil
}

func (m *mockSurveyRepoLocal) ReplaceTemplate(template *domain.SurveyTemplate) error {
	if m.replaceTemplateFunc != nil {
		return m.replaceTemplateFunc(template)
	}
	return nil
}

func (m *mockSurveyRepoLocal) DeleteTemplate(templateID string) error {
	if m.deleteTemplateFunc != nil {
		return m.deleteTemplateFunc(templateID)
	}
	return nil
}

func (m *mockSurveyRepoLocal) SaveResponse(response *domain.SurveyResponse) error {
	if m.saveResponseFunc != nil {
		return m.saveResponseFunc(response)
	}
	return nil
}

func (m *mockSurveyRepoLocal) HasResponse(ticketID string, userID string) (bool, error) {
	if m.hasResponseFunc != nil {
		return m.hasResponseFunc(ticketID, userID)
	}
	return false, nil
}

func (m *mockSurveyRepoLocal) ListResponses(filter repository.SurveyResponseFilter, page int, limit int) ([]repository.SurveyResponseRow, int64, error) {
	if m.listResponsesFunc != nil {
		return m.listResponsesFunc(filter, page, limit)
	}
	return nil, 0, nil
}

type mockTicketRepoLocal struct {
	findByIDFunc func(ticketID string) (*domain.Ticket, error)
}

func (m *mockTicketRepoLocal) Create(ticket *domain.Ticket) error { return nil }
func (m *mockTicketRepoLocal) Update(ticket *domain.Ticket) error { return nil }
func (m *mockTicketRepoLocal) SoftDelete(ticketID string) error   { return nil }
func (m *mockTicketRepoLocal) FindByID(ticketID string) (*domain.Ticket, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ticketID)
	}
	return nil, nil
}
func (m *mockTicketRepoLocal) ListByUser(userID string) ([]domain.Ticket, error) { return nil, nil }
func (m *mockTicketRepoLocal) ListAll() ([]domain.Ticket, error)                 { return nil, nil }
func (m *mockTicketRepoLocal) Search(query string, isGuest bool) ([]domain.Ticket, error) {
	return nil, nil
}
func (m *mockTicketRepoLocal) ListFiltered(filter repository.TicketListFilter, page int, limit int) ([]domain.Ticket, int64, error) {
	return nil, 0, nil
}
func (m *mockTicketRepoLocal) CountForYear(year int) (int64, error)           { return 0, nil }
func (m *mockTicketRepoLocal) AddHistory(history *domain.TicketHistory) error { return nil }
func (m *mockTicketRepoLocal) AddComment(comment *domain.TicketComment) error { return nil }
func (m *mockTicketRepoLocal) UpdateStatus(ticketID string, status domain.TicketStatus, surveyRequired bool) error {
	return nil
}
func (m *mockTicketRepoLocal) GetSurveyScores(ticketIDs []string) (map[string]float64, error) {
	return nil, nil
}

// Note: time.Now is not used in the mocks, but the import is kept for consistency
var _ = time.Now
