package service

import (
	"encoding/json"
	"testing"
	"time"

	"unila_helpdesk_backend/internal/domain"
)

// ============================================================================
// Tests for normalizePeriod
// ============================================================================

func TestNormalizePeriod_Daily(t *testing.T) {
	testCases := []string{"daily", "Daily", "DAILY", " daily "}
	for _, tc := range testCases {
		got := normalizePeriod(tc)
		if got != "daily" {
			t.Errorf("normalizePeriod(%q) = %q, want 'daily'", tc, got)
		}
	}
}

func TestNormalizePeriod_Weekly(t *testing.T) {
	testCases := []string{"weekly", "Weekly", "WEEKLY", " weekly "}
	for _, tc := range testCases {
		got := normalizePeriod(tc)
		if got != "weekly" {
			t.Errorf("normalizePeriod(%q) = %q, want 'weekly'", tc, got)
		}
	}
}

func TestNormalizePeriod_Yearly(t *testing.T) {
	testCases := []string{"yearly", "Yearly", "YEARLY", " yearly "}
	for _, tc := range testCases {
		got := normalizePeriod(tc)
		if got != "yearly" {
			t.Errorf("normalizePeriod(%q) = %q, want 'yearly'", tc, got)
		}
	}
}

func TestNormalizePeriod_Default(t *testing.T) {
	// Unknown values should default to "monthly"
	testCases := []string{"monthly", "Monthly", "MONTHLY", "invalid", "", "quarterly"}
	for _, tc := range testCases {
		got := normalizePeriod(tc)
		if got != "monthly" {
			t.Errorf("normalizePeriod(%q) = %q, want 'monthly'", tc, got)
		}
	}
}

// ============================================================================
// Tests for periodStart
// ============================================================================

func TestPeriodStart_Daily(t *testing.T) {
	input := time.Date(2026, 2, 6, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC)
	got := periodStart(input, "daily")
	if !got.Equal(expected) {
		t.Errorf("periodStart(daily) = %v, want %v", got, expected)
	}
}

func TestPeriodStart_Weekly_Monday(t *testing.T) {
	// Feb 3, 2026 is Monday
	input := time.Date(2026, 2, 3, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday of that week
	got := periodStart(input, "weekly")
	if !got.Equal(expected) {
		t.Errorf("periodStart(weekly, Monday) = %v, want %v", got, expected)
	}
}

func TestPeriodStart_Weekly_Friday(t *testing.T) {
	// Feb 6, 2026 is Friday
	input := time.Date(2026, 2, 6, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday of that week
	got := periodStart(input, "weekly")
	if !got.Equal(expected) {
		t.Errorf("periodStart(weekly, Friday) = %v, want %v", got, expected)
	}
}

func TestPeriodStart_Weekly_Sunday(t *testing.T) {
	// Feb 8, 2026 is Sunday
	input := time.Date(2026, 2, 8, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday of that week
	got := periodStart(input, "weekly")
	if !got.Equal(expected) {
		t.Errorf("periodStart(weekly, Sunday) = %v, want %v", got, expected)
	}
}

func TestPeriodStart_Monthly(t *testing.T) {
	input := time.Date(2026, 2, 15, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	got := periodStart(input, "monthly")
	if !got.Equal(expected) {
		t.Errorf("periodStart(monthly) = %v, want %v", got, expected)
	}
}

func TestPeriodStart_Yearly(t *testing.T) {
	input := time.Date(2026, 6, 15, 15, 30, 45, 0, time.UTC)
	expected := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := periodStart(input, "yearly")
	if !got.Equal(expected) {
		t.Errorf("periodStart(yearly) = %v, want %v", got, expected)
	}
}

// ============================================================================
// Tests for addPeriods
// ============================================================================

func TestAddPeriods_Daily(t *testing.T) {
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got := addPeriods(base, "daily", 5)
	expected := time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("addPeriods(daily, 5) = %v, want %v", got, expected)
	}

	got = addPeriods(base, "daily", -3)
	expected = time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("addPeriods(daily, -3) = %v, want %v", got, expected)
	}
}

func TestAddPeriods_Weekly(t *testing.T) {
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got := addPeriods(base, "weekly", 2)
	expected := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("addPeriods(weekly, 2) = %v, want %v", got, expected)
	}
}

func TestAddPeriods_Monthly(t *testing.T) {
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got := addPeriods(base, "monthly", 3)
	expected := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("addPeriods(monthly, 3) = %v, want %v", got, expected)
	}
}

func TestAddPeriods_Yearly(t *testing.T) {
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got := addPeriods(base, "yearly", 2)
	expected := time.Date(2028, 2, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("addPeriods(yearly, 2) = %v, want %v", got, expected)
	}
}

// ============================================================================
// Tests for formatCohortLabel
// ============================================================================

func TestFormatCohortLabel_Daily(t *testing.T) {
	date := time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC)
	got := formatCohortLabel(date, "daily")
	expected := "06 Feb 2026"
	if got != expected {
		t.Errorf("formatCohortLabel(daily) = %q, want %q", got, expected)
	}
}

func TestFormatCohortLabel_Weekly(t *testing.T) {
	date := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	got := formatCohortLabel(date, "weekly")
	expected := "Week of 02 Feb 2026"
	if got != expected {
		t.Errorf("formatCohortLabel(weekly) = %q, want %q", got, expected)
	}
}

func TestFormatCohortLabel_Monthly(t *testing.T) {
	date := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	got := formatCohortLabel(date, "monthly")
	expected := "Feb 2026"
	if got != expected {
		t.Errorf("formatCohortLabel(monthly) = %q, want %q", got, expected)
	}
}

func TestFormatCohortLabel_Yearly(t *testing.T) {
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := formatCohortLabel(date, "yearly")
	expected := "2026"
	if got != expected {
		t.Errorf("formatCohortLabel(yearly) = %q, want %q", got, expected)
	}
}

// ============================================================================
// Tests for scoreFromValue
// ============================================================================

func TestScoreFromValue_Float64(t *testing.T) {
	testCases := []struct {
		input    float64
		expected float64
		valid    bool
	}{
		{1, 0, true},
		{3, 50, true},
		{5, 100, true},
		{0, 0, false}, // out of range
		{6, 0, false}, // out of range
	}
	for _, tc := range testCases {
		score, ok := scoreFromValue(tc.input)
		if ok != tc.valid {
			t.Errorf("scoreFromValue(%v) ok=%v, want %v", tc.input, ok, tc.valid)
		}
		if ok && score != tc.expected {
			t.Errorf("scoreFromValue(%v) = %v, want %v", tc.input, score, tc.expected)
		}
	}
}

func TestScoreFromValue_Int(t *testing.T) {
	testCases := []struct {
		input    int
		expected float64
		valid    bool
	}{
		{1, 0, true},
		{3, 50, true},
		{5, 100, true},
		{0, 0, false},
		{6, 0, false},
	}
	for _, tc := range testCases {
		score, ok := scoreFromValue(tc.input)
		if ok != tc.valid {
			t.Errorf("scoreFromValue(%d) ok=%v, want %v", tc.input, ok, tc.valid)
		}
		if ok && score != tc.expected {
			t.Errorf("scoreFromValue(%d) = %v, want %v", tc.input, score, tc.expected)
		}
	}
}

func TestScoreFromValue_Bool(t *testing.T) {
	score, ok := scoreFromValue(true)
	if !ok || score != 100 {
		t.Errorf("scoreFromValue(true) = %v, %v, want 100, true", score, ok)
	}

	score, ok = scoreFromValue(false)
	if !ok || score != 0 {
		t.Errorf("scoreFromValue(false) = %v, %v, want 0, true", score, ok)
	}
}

func TestScoreFromValue_String(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
		valid    bool
	}{
		{"ya", 100, true},
		{"Yes", 100, true},
		{"true", 100, true},
		{"tidak", 0, true},
		{"No", 0, true},
		{"false", 0, true},
		{"3", 50, true},
		{"5", 100, true},
		{"invalid", 0, false},
		{"", 0, false},
	}
	for _, tc := range testCases {
		score, ok := scoreFromValue(tc.input)
		if ok != tc.valid {
			t.Errorf("scoreFromValue(%q) ok=%v, want %v", tc.input, ok, tc.valid)
		}
		if ok && score != tc.expected {
			t.Errorf("scoreFromValue(%q) = %v, want %v", tc.input, score, tc.expected)
		}
	}
}

// ============================================================================
// Tests for scoreFromAnswers
// ============================================================================

func TestScoreFromAnswers_Valid(t *testing.T) {
	answers := json.RawMessage(`{"q1": 5, "q2": 3, "q3": 1}`)
	score := scoreFromAnswers(answers)
	// Avg of normalized: (100 + 50 + 0) / 3 = 50
	if score != 50 {
		t.Errorf("scoreFromAnswers = %v, want 50", score)
	}
}

func TestScoreFromAnswers_Empty(t *testing.T) {
	answers := json.RawMessage(`{}`)
	score := scoreFromAnswers(answers)
	if score != 0 {
		t.Errorf("scoreFromAnswers({}) = %v, want 0", score)
	}
}

func TestScoreFromAnswers_Invalid(t *testing.T) {
	answers := json.RawMessage(`invalid json`)
	score := scoreFromAnswers(answers)
	if score != 0 {
		t.Errorf("scoreFromAnswers(invalid) = %v, want 0", score)
	}
}

func TestScoreFromAnswers_MixedTypes(t *testing.T) {
	answers := json.RawMessage(`{"q1": "ya", "q2": 3}`)
	score := scoreFromAnswers(answers)
	// Avg of: (100 + 50) / 2 = 75
	if score != 75 {
		t.Errorf("scoreFromAnswers(mixed) = %v, want 75", score)
	}
}

// ============================================================================
// Tests for calculateCohortScores
// ============================================================================

func TestCalculateCohortScores_Empty(t *testing.T) {
	responses := []domain.SurveyResponse{}
	avg, rate := calculateCohortScores(responses)
	if avg != 0 || rate != 0 {
		t.Errorf("calculateCohortScores([]) = %v, %v, want 0, 0", avg, rate)
	}
}

func TestCalculateCohortScores_WithScores(t *testing.T) {
	responses := []domain.SurveyResponse{
		{Score: 80}, // Already normalized (>5), no change
		{Score: 60},
	}
	avg, rate := calculateCohortScores(responses)
	// Avg: (80 + 60) / 2 = 70
	if avg != 70 {
		t.Errorf("calculateCohortScores avg = %v, want 70", avg)
	}
	// Rate: 2/2 * 100 = 100
	if rate != 100 {
		t.Errorf("calculateCohortScores rate = %v, want 100", rate)
	}
}

func TestCalculateCohortScores_LegacyScores(t *testing.T) {
	responses := []domain.SurveyResponse{
		{Score: 5}, // Legacy 1-5 scale, should normalize to 100
		{Score: 3}, // Should normalize to 50
	}
	avg, rate := calculateCohortScores(responses)
	// Avg: (100 + 50) / 2 = 75
	if avg != 75 {
		t.Errorf("calculateCohortScores(legacy) avg = %v, want 75", avg)
	}
	if rate != 100 {
		t.Errorf("calculateCohortScores(legacy) rate = %v, want 100", rate)
	}
}

// ============================================================================
// Tests for scaleMaxFor
// ============================================================================

func TestScaleMaxFor_AllTypes(t *testing.T) {
	testCases := []struct {
		qType    domain.SurveyQuestionType
		expected int
	}{
		{domain.QuestionLikert3, 3},
		{domain.QuestionLikert3Puas, 3},
		{domain.QuestionLikert4, 4},
		{domain.QuestionLikert4Puas, 4},
		{domain.QuestionLikert, 5},
		{domain.QuestionLikertQuality, 5},
		{domain.QuestionYesNo, 0},
		{domain.QuestionText, 0},
		{domain.QuestionMultipleChoice, 0},
	}
	for _, tc := range testCases {
		got := scaleMaxFor(tc.qType)
		if got != tc.expected {
			t.Errorf("scaleMaxFor(%s) = %d, want %d", tc.qType, got, tc.expected)
		}
	}
}

// ============================================================================
// Tests for parseNumericValue
// ============================================================================

func TestParseNumericValue_Float64(t *testing.T) {
	val, ok := parseNumericValue(3.5)
	if !ok || val != 3.5 {
		t.Errorf("parseNumericValue(3.5) = %v, %v, want 3.5, true", val, ok)
	}
}

func TestParseNumericValue_Int(t *testing.T) {
	val, ok := parseNumericValue(5)
	if !ok || val != 5 {
		t.Errorf("parseNumericValue(5) = %v, %v, want 5, true", val, ok)
	}
}

func TestParseNumericValue_String(t *testing.T) {
	val, ok := parseNumericValue(" 3.5 ")
	if !ok || val != 3.5 {
		t.Errorf("parseNumericValue(' 3.5 ') = %v, %v, want 3.5, true", val, ok)
	}
}

func TestParseNumericValue_InvalidString(t *testing.T) {
	_, ok := parseNumericValue("abc")
	if ok {
		t.Error("parseNumericValue('abc') should return false")
	}
}

func TestParseNumericValue_UnsupportedType(t *testing.T) {
	_, ok := parseNumericValue([]int{1, 2})
	if ok {
		t.Error("parseNumericValue(slice) should return false")
	}
}

// ============================================================================
// Tests for answerKey
// ============================================================================

func TestAnswerKey_YesNo_Bool(t *testing.T) {
	key, ok := answerKey(true, domain.QuestionYesNo)
	if !ok || key != "Ya" {
		t.Errorf("answerKey(true, YesNo) = %q, %v, want 'Ya', true", key, ok)
	}

	key, ok = answerKey(false, domain.QuestionYesNo)
	if !ok || key != "Tidak" {
		t.Errorf("answerKey(false, YesNo) = %q, %v, want 'Tidak', true", key, ok)
	}
}

func TestAnswerKey_YesNo_String(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"ya", "Ya"},
		{"Yes", "Ya"},
		{"true", "Ya"},
		{"tidak", "Tidak"},
		{"No", "Tidak"},
		{"false", "Tidak"},
	}
	for _, tc := range testCases {
		key, ok := answerKey(tc.input, domain.QuestionYesNo)
		if !ok || key != tc.expected {
			t.Errorf("answerKey(%q, YesNo) = %q, %v, want %q, true", tc.input, key, ok, tc.expected)
		}
	}
}

func TestAnswerKey_MultipleChoice(t *testing.T) {
	key, ok := answerKey("Option A", domain.QuestionMultipleChoice)
	if !ok || key != "Option A" {
		t.Errorf("answerKey('Option A', MC) = %q, %v, want 'Option A', true", key, ok)
	}

	_, ok = answerKey("", domain.QuestionMultipleChoice)
	if ok {
		t.Error("answerKey('', MC) should return false")
	}
}

func TestAnswerKey_Text(t *testing.T) {
	_, ok := answerKey("any text", domain.QuestionText)
	if ok {
		t.Error("answerKey(Text) should always return false")
	}
}

func TestAnswerKey_Likert(t *testing.T) {
	key, ok := answerKey(3, domain.QuestionLikert)
	if !ok || key != "3" {
		t.Errorf("answerKey(3, Likert) = %q, %v, want '3', true", key, ok)
	}

	// Out of range
	_, ok = answerKey(6, domain.QuestionLikert)
	if ok {
		t.Error("answerKey(6, Likert) should return false (out of range)")
	}
}

// ============================================================================
// Tests for periodRange
// ============================================================================

func TestPeriodRange_Monthly5Periods(t *testing.T) {
	fixedNow := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedNow }

	start, end := periodRange("monthly", 5, nowFn)

	// End should be start of next month from current
	expectedEnd := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	// Start should be 4 months back (5 periods including current)
	expectedStart := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	if !end.Equal(expectedEnd) {
		t.Errorf("periodRange end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("periodRange start = %v, want %v", start, expectedStart)
	}
}

func TestPeriodRange_Weekly3Periods(t *testing.T) {
	// Feb 6, 2026 is Friday
	fixedNow := time.Date(2026, 2, 6, 10, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedNow }

	start, end := periodRange("weekly", 3, nowFn)

	// Current week starts Monday Feb 2
	// End = Feb 2 + 1 week = Feb 9
	expectedEnd := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	// Start = Feb 2 - 2 weeks = Jan 19
	expectedStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)

	if !end.Equal(expectedEnd) {
		t.Errorf("periodRange(weekly) end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("periodRange(weekly) start = %v, want %v", start, expectedStart)
	}
}

func TestPeriodRange_Daily7Periods(t *testing.T) {
	fixedNow := time.Date(2026, 2, 6, 10, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedNow }

	start, end := periodRange("daily", 7, nowFn)

	expectedEnd := time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC)
	expectedStart := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	if !end.Equal(expectedEnd) {
		t.Errorf("periodRange(daily) end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("periodRange(daily) start = %v, want %v", start, expectedStart)
	}
}

func TestPeriodRange_DefaultPeriods(t *testing.T) {
	fixedNow := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedNow }

	// periods <= 0 should default to 5
	start1, end1 := periodRange("monthly", 0, nowFn)
	start2, end2 := periodRange("monthly", 5, nowFn)

	if !start1.Equal(start2) || !end1.Equal(end2) {
		t.Error("periodRange with 0 periods should default to 5")
	}
}

func TestPeriodRange_Yearly(t *testing.T) {
	fixedNow := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	nowFn := func() time.Time { return fixedNow }

	start, end := periodRange("yearly", 3, nowFn)

	expectedEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if !end.Equal(expectedEnd) {
		t.Errorf("periodRange(yearly) end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("periodRange(yearly) start = %v, want %v", start, expectedStart)
	}
}

// ============================================================================
// Tests for mapSurveyTemplateDTO
// ============================================================================

func TestMapSurveyTemplateDTO_Basic(t *testing.T) {
	template := domain.SurveyTemplate{
		ID:          "template-1",
		Title:       "Test Template",
		Description: "A test template",
		Framework:   "RATER",
		CategoryID:  "cat-1",
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	dto := mapSurveyTemplateDTO(template)

	if dto.ID != template.ID {
		t.Errorf("ID mismatch: got %s, want %s", dto.ID, template.ID)
	}
	if dto.Title != template.Title {
		t.Errorf("Title mismatch: got %s, want %s", dto.Title, template.Title)
	}
	if dto.Description != template.Description {
		t.Errorf("Description mismatch")
	}
	if dto.Framework != template.Framework {
		t.Errorf("Framework mismatch")
	}
	if dto.CategoryID != template.CategoryID {
		t.Errorf("CategoryID mismatch")
	}
	if len(dto.Questions) != 0 {
		t.Errorf("Expected 0 questions, got %d", len(dto.Questions))
	}
}

func TestMapSurveyTemplateDTO_WithQuestions(t *testing.T) {
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

	dto := mapSurveyTemplateDTO(template)

	if len(dto.Questions) != 2 {
		t.Fatalf("Expected 2 questions, got %d", len(dto.Questions))
	}

	// Check first question
	if dto.Questions[0].ID != "q1" {
		t.Errorf("Question 1 ID mismatch")
	}
	if dto.Questions[0].Type != "likert" {
		t.Errorf("Question 1 Type = %s, want 'likert'", dto.Questions[0].Type)
	}
	if len(dto.Questions[0].Options) != 5 {
		t.Errorf("Question 1 Options = %d, want 5", len(dto.Questions[0].Options))
	}

	// Check second question
	if dto.Questions[1].ID != "q2" {
		t.Errorf("Question 2 ID mismatch")
	}
	if dto.Questions[1].Type != "yesNo" {
		t.Errorf("Question 2 Type = %s, want 'yesNo'", dto.Questions[1].Type)
	}
}

func TestMapSurveyTemplateDTO_InvalidOptions(t *testing.T) {
	template := domain.SurveyTemplate{
		ID: "template-1",
		Questions: []domain.SurveyQuestion{
			{
				ID:      "q1",
				Text:    "Question 1",
				Type:    domain.QuestionLikert,
				Options: []byte(`invalid json`),
			},
		},
	}

	dto := mapSurveyTemplateDTO(template)

	// Should still map, but options will be nil/empty
	if len(dto.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(dto.Questions))
	}
	if len(dto.Questions[0].Options) > 0 {
		t.Errorf("Expected nil/empty options for invalid JSON")
	}
}

// ============================================================================
// Tests for mapSurveyTemplateDTOs
// ============================================================================

func TestMapSurveyTemplateDTOs_Empty(t *testing.T) {
	templates := []domain.SurveyTemplate{}
	result := mapSurveyTemplateDTOs(templates)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d", len(result))
	}
}

func TestMapSurveyTemplateDTOs_Multiple(t *testing.T) {
	templates := []domain.SurveyTemplate{
		{ID: "t1", Title: "Template 1"},
		{ID: "t2", Title: "Template 2"},
		{ID: "t3", Title: "Template 3"},
	}

	result := mapSurveyTemplateDTOs(templates)

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}
	for i, dto := range result {
		if dto.ID != templates[i].ID {
			t.Errorf("Result[%d].ID = %s, want %s", i, dto.ID, templates[i].ID)
		}
	}
}
