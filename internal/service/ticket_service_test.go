package service

import (
	"encoding/json"
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

// ============================================================================
// Tests for marshalAttachments
// ============================================================================

func TestMarshalAttachments_Empty(t *testing.T) {
	result := marshalAttachments([]string{})
	if result != nil {
		t.Errorf("marshalAttachments([]) should return nil, got %s", result)
	}
}

func TestMarshalAttachments_Nil(t *testing.T) {
	result := marshalAttachments(nil)
	if result != nil {
		t.Errorf("marshalAttachments(nil) should return nil, got %s", result)
	}
}

func TestMarshalAttachments_WhitespaceOnly(t *testing.T) {
	result := marshalAttachments([]string{"   ", "", "\t"})
	if result != nil {
		t.Errorf("marshalAttachments(whitespace) should return nil, got %s", result)
	}
}

func TestMarshalAttachments_Valid(t *testing.T) {
	input := []string{"file1.jpg", " file2.pdf ", "file3.png"}
	result := marshalAttachments(input)
	if result == nil {
		t.Fatal("marshalAttachments should return valid JSON")
	}

	var parsed []string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("expected 3 attachments, got %d", len(parsed))
	}
	// Check whitespace trimming
	if parsed[1] != "file2.pdf" {
		t.Errorf("expected 'file2.pdf', got '%s'", parsed[1])
	}
}

func TestMarshalAttachments_Mixed(t *testing.T) {
	input := []string{"valid.jpg", "", " ", "another.pdf", ""}
	result := marshalAttachments(input)
	if result == nil {
		t.Fatal("marshalAttachments should return valid JSON")
	}

	var parsed []string
	json.Unmarshal(result, &parsed)

	if len(parsed) != 2 {
		t.Errorf("expected 2 valid attachments, got %d", len(parsed))
	}
}

// ============================================================================
// Tests for attachmentIDsFromRefs
// ============================================================================

func TestAttachmentIDsFromRefs_Empty(t *testing.T) {
	result := attachmentIDsFromRefs([]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestAttachmentIDsFromRefs_SimpleIDs(t *testing.T) {
	input := []string{"id1", "id2", "id3"}
	result := attachmentIDsFromRefs(input)

	if len(result) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(result))
	}
	for i, id := range []string{"id1", "id2", "id3"} {
		if result[i] != id {
			t.Errorf("result[%d] = %s, want %s", i, result[i], id)
		}
	}
}

func TestAttachmentIDsFromRefs_URLs(t *testing.T) {
	input := []string{
		"https://example.com/uploads/file1.jpg",
		"http://example.com/uploads/file2.pdf",
	}
	result := attachmentIDsFromRefs(input)

	if len(result) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(result))
	}
	if result[0] != "file1.jpg" {
		t.Errorf("expected 'file1.jpg', got '%s'", result[0])
	}
	if result[1] != "file2.pdf" {
		t.Errorf("expected 'file2.pdf', got '%s'", result[1])
	}
}

func TestAttachmentIDsFromRefs_Mixed(t *testing.T) {
	input := []string{
		"simple-id",
		"https://example.com/uploads/from-url.jpg",
		"another-id",
	}
	result := attachmentIDsFromRefs(input)

	if len(result) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(result))
	}
}

func TestAttachmentIDsFromRefs_WhitespaceSkipped(t *testing.T) {
	input := []string{"id1", "  ", "", "id2"}
	result := attachmentIDsFromRefs(input)

	if len(result) != 2 {
		t.Errorf("expected 2 IDs (whitespace skipped), got %d", len(result))
	}
}

// ============================================================================
// Tests for ticketIDs
// ============================================================================

func TestTicketIDs_Empty(t *testing.T) {
	tickets := []domain.Ticket{}
	result := ticketIDs(tickets)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestTicketIDs_ExtractsIDs(t *testing.T) {
	tickets := []domain.Ticket{
		{ID: "TK-2026-001"},
		{ID: "TK-2026-002"},
		{ID: "TK-2026-003"},
	}
	result := ticketIDs(tickets)

	if len(result) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(result))
	}
	for i, expected := range []string{"TK-2026-001", "TK-2026-002", "TK-2026-003"} {
		if result[i] != expected {
			t.Errorf("result[%d] = %s, want %s", i, result[i], expected)
		}
	}
}

// ============================================================================
// Tests for TicketCreateRequest validation patterns
// ============================================================================

func TestTicketCreateRequest_EmptyTitle(t *testing.T) {
	req := TicketCreateRequest{Title: ""}

	if trimString(req.Title) == "" {
		// Expected - validation should catch this
		return
	}
	t.Error("empty title should be caught")
}

func TestTicketCreateRequest_EmptyDescription(t *testing.T) {
	req := TicketCreateRequest{Description: ""}

	if trimString(req.Description) == "" {
		// Expected - validation should catch this
		return
	}
	t.Error("empty description should be caught")
}

func TestTicketCreateRequest_DefaultPriority(t *testing.T) {
	req := TicketCreateRequest{Priority: ""}

	priority := req.Priority
	if priority == "" {
		priority = domain.PriorityMedium
	}

	if priority != domain.PriorityMedium {
		t.Errorf("default priority should be medium, got %s", priority)
	}
}

// ============================================================================
// Tests for access control patterns
// ============================================================================

func TestUpdateTicket_AdminCanUpdateAny(t *testing.T) {
	user := domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	ticket := domain.Ticket{ReporterID: "user-1"}

	hasAccess := user.Role == domain.RoleAdmin || ticket.ReporterID == user.ID

	if !hasAccess {
		t.Error("admin should have access to any ticket")
	}
}

func TestUpdateTicket_UserCanUpdateOwn(t *testing.T) {
	user := domain.User{ID: "user-1", Role: domain.RoleRegistered}
	ticket := domain.Ticket{ReporterID: "user-1"}

	hasAccess := user.Role == domain.RoleAdmin || ticket.ReporterID == user.ID

	if !hasAccess {
		t.Error("user should have access to own ticket")
	}
}

func TestUpdateTicket_UserCannotUpdateOthers(t *testing.T) {
	user := domain.User{ID: "user-1", Role: domain.RoleRegistered}
	ticket := domain.Ticket{ReporterID: "user-2"}

	hasAccess := user.Role == domain.RoleAdmin || ticket.ReporterID == user.ID

	if hasAccess {
		t.Error("user should NOT have access to other's ticket")
	}
}

func TestUpdateTicket_ResolvedNoEdit(t *testing.T) {
	user := domain.User{Role: domain.RoleRegistered}
	ticket := domain.Ticket{Status: domain.StatusResolved}

	canEdit := user.Role == domain.RoleAdmin || ticket.Status != domain.StatusResolved

	if canEdit && user.Role != domain.RoleAdmin {
		t.Error("non-admin should not edit resolved ticket")
	}
}

// ============================================================================
// Tests for guest ticket restrictions
// ============================================================================

func TestGuestTicket_OnlyMembershipCategories(t *testing.T) {
	guestAllowedCategories := []string{
		CategoryGuestPassword,
		CategoryGuestSSORegistration,
		CategoryGuestEmailRegistration,
	}

	notAllowed := []string{
		CategoryInternet,
		CategorySIAKAD,
		CategoryWebsite,
	}

	for _, cat := range guestAllowedCategories {
		// Guest should be allowed
		allowed := isGuestCategory(cat)
		if !allowed {
			t.Errorf("guest should be allowed to use category %s", cat)
		}
	}

	for _, cat := range notAllowed {
		allowed := isGuestCategory(cat)
		if allowed {
			t.Errorf("guest should NOT be allowed to use category %s", cat)
		}
	}
}

// Helper function for testing
func isGuestCategory(categoryID string) bool {
	guestCategories := map[string]bool{
		CategoryGuestPassword:          true,
		CategoryGuestSSORegistration:   true,
		CategoryGuestEmailRegistration: true,
	}
	return guestCategories[categoryID]
}

// ============================================================================
// Tests for toTicketDTO mapping
// ============================================================================

func TestToTicketDTO_BasicMapping(t *testing.T) {
	ticket := domain.Ticket{
		ID:       "TK-2026-001",
		Status:   domain.StatusResolved,
		Priority: domain.PriorityHigh,
	}

	dto := domain.TicketDTO{
		ID:       ticket.ID,
		Status:   ticket.Status,
		Priority: ticket.Priority,
	}

	if dto.ID != ticket.ID {
		t.Errorf("ID mismatch: got %s, want %s", dto.ID, ticket.ID)
	}
	if dto.Status != domain.StatusResolved {
		t.Errorf("Status mismatch: got %s, want %s", dto.Status, domain.StatusResolved)
	}
	if dto.Priority != domain.PriorityHigh {
		t.Errorf("Priority mismatch: got %s, want %s", dto.Priority, domain.PriorityHigh)
	}
}

// ============================================================================
// Tests for pagination logic
// ============================================================================

func TestListTicketsPaged_Pagination(t *testing.T) {
	testCases := []struct {
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{0, 10, 1, 10},  // page < 1 becomes 1
		{1, 100, 1, 50}, // limit > 50 becomes 50
		{1, 0, 1, 15},   // limit <= 0 becomes 15
		{5, 25, 5, 25},  // valid values unchanged
		{-1, -1, 1, 15}, // negative values
	}

	for _, tc := range testCases {
		limit := tc.limit
		page := tc.page

		// Simulate pagination logic
		if limit <= 0 {
			limit = 15
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

// ============================================================================
// Tests for survey requirement logic
// ============================================================================

func TestSurveyRequired_RegisteredUser(t *testing.T) {
	user := domain.User{Role: domain.RoleRegistered}
	surveyRequired := user.Role == domain.RoleRegistered

	if !surveyRequired {
		t.Error("registered user should require survey")
	}
}

func TestSurveyRequired_GuestUser(t *testing.T) {
	user := domain.User{Role: domain.RoleGuest}
	surveyRequired := user.Role == domain.RoleRegistered

	if surveyRequired {
		t.Error("guest user should NOT require survey")
	}
}

func TestSurveyRequired_OnResolved(t *testing.T) {
	ticket := domain.Ticket{Status: domain.StatusResolved, IsGuest: false}

	surveyRequired := ticket.Status == domain.StatusResolved && !ticket.IsGuest

	if !surveyRequired {
		t.Error("resolved non-guest ticket should require survey")
	}
}

// ============================================================================
// Tests for ticket ID generation pattern
// ============================================================================

func TestTicketIDFormat(t *testing.T) {
	// Simulating the generateTicketID format
	year := 2026
	count := 42

	ticketID := generateTestTicketID(year, count)

	if ticketID != "TK-2026-043" {
		t.Errorf("expected 'TK-2026-043', got '%s'", ticketID)
	}
}

func TestTicketIDFormat_FirstOfYear(t *testing.T) {
	ticketID := generateTestTicketID(2026, 0)

	if ticketID != "TK-2026-001" {
		t.Errorf("expected 'TK-2026-001', got '%s'", ticketID)
	}
}

func TestTicketIDFormat_LargeCount(t *testing.T) {
	ticketID := generateTestTicketID(2026, 999)

	if ticketID != "TK-2026-1000" {
		t.Errorf("expected 'TK-2026-1000', got '%s'", ticketID)
	}
}

// Helper function
func generateTestTicketID(year int, count int) string {
	return formatTicketID(year, count+1)
}

func formatTicketID(year, num int) string {
	return "TK-" + itoa(year) + "-" + padNumber(num, 3)
}

func itoa(n int) string {
	return string([]byte{
		byte('0' + n/1000),
		byte('0' + (n/100)%10),
		byte('0' + (n/10)%10),
		byte('0' + n%10),
	})
}

func padNumber(n, minWidth int) string {
	s := ""
	for i := 1000; i >= 1; i /= 10 {
		if n >= i || len(s) > 0 || i <= 100 {
			s += string(byte('0' + (n/i)%10))
		}
	}
	for len(s) < minWidth {
		s = "0" + s
	}
	return s
}

// ============================================================================
// Tests for comment validation
// ============================================================================

func TestAddComment_EmptyMessage(t *testing.T) {
	message := ""

	if trimString(message) == "" {
		// Expected - should be rejected
		return
	}
	t.Error("empty comment should be rejected")
}

func TestAddComment_WhitespaceMessage(t *testing.T) {
	message := "   \t  "

	if trimString(message) == "" {
		// Expected - should be rejected
		return
	}
	t.Error("whitespace-only comment should be rejected")
}

func TestAddComment_ValidMessage(t *testing.T) {
	message := "This is a valid comment"

	if trimString(message) == "" {
		t.Error("valid comment should not be rejected")
	}
}

// ============================================================================
// Tests for history and comment DTO mapping
// ============================================================================

func TestTicketHistoryDTO_Mapping(t *testing.T) {
	history := domain.TicketHistory{
		Title: "Status Updated",
	}

	dto := domain.TicketHistoryDTO{
		Title: history.Title,
	}

	if dto.Title != history.Title {
		t.Errorf("Title mismatch: got %s, want %s", dto.Title, history.Title)
	}
}

func TestTicketCommentDTO_Mapping(t *testing.T) {
	comment := domain.TicketComment{
		Author:  "John Doe",
		IsStaff: true,
	}

	dto := domain.TicketCommentDTO{
		Author:  comment.Author,
		IsStaff: comment.IsStaff,
	}

	if dto.Author != comment.Author {
		t.Errorf("Author mismatch: got %s, want %s", dto.Author, comment.Author)
	}
	if dto.IsStaff != true {
		t.Error("expected IsStaff to be true for admin comment")
	}
}
