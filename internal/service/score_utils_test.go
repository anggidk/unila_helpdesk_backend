package service

import (
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

// ============================================================================
// Tests for scoreFromYesNo
// ============================================================================

func TestScoreFromYesNo_BoolTrue(t *testing.T) {
	score, ok := scoreFromYesNo(true)
	if !ok {
		t.Error("expected ok=true for bool true")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromYesNo_BoolFalse(t *testing.T) {
	score, ok := scoreFromYesNo(false)
	if !ok {
		t.Error("expected ok=true for bool false")
	}
	if score != 0 {
		t.Errorf("expected score=0, got %v", score)
	}
}

func TestScoreFromYesNo_StringYa(t *testing.T) {
	testCases := []string{"ya", "Ya", "YA", " ya ", "yes", "Yes", "YES", "true", "True"}
	for _, tc := range testCases {
		score, ok := scoreFromYesNo(tc)
		if !ok {
			t.Errorf("expected ok=true for %q", tc)
		}
		if score != 100 {
			t.Errorf("expected score=100 for %q, got %v", tc, score)
		}
	}
}

func TestScoreFromYesNo_StringTidak(t *testing.T) {
	testCases := []string{"tidak", "Tidak", "TIDAK", " tidak ", "no", "No", "NO", "false", "False"}
	for _, tc := range testCases {
		score, ok := scoreFromYesNo(tc)
		if !ok {
			t.Errorf("expected ok=true for %q", tc)
		}
		if score != 0 {
			t.Errorf("expected score=0 for %q, got %v", tc, score)
		}
	}
}

func TestScoreFromYesNo_InvalidInput(t *testing.T) {
	testCases := []interface{}{"mungkin", "maybe", 123, 1.5, nil}
	for _, tc := range testCases {
		_, ok := scoreFromYesNo(tc)
		if ok {
			t.Errorf("expected ok=false for %v", tc)
		}
	}
}

// ============================================================================
// Tests for scoreFromScale
// ============================================================================

func TestScoreFromScale_Likert5_AllValues(t *testing.T) {
	// Likert 5: 1=0%, 2=25%, 3=50%, 4=75%, 5=100%
	expected := map[int]float64{
		1: 0,
		2: 25,
		3: 50,
		4: 75,
		5: 100,
	}
	for input, want := range expected {
		score, ok := scoreFromScale(input, 5)
		if !ok {
			t.Errorf("expected ok=true for input %d", input)
		}
		if score != want {
			t.Errorf("scoreFromScale(%d, 5) = %v, want %v", input, score, want)
		}
	}
}

func TestScoreFromScale_Likert3_AllValues(t *testing.T) {
	// Likert 3: 1=0%, 2=50%, 3=100%
	expected := map[int]float64{
		1: 0,
		2: 50,
		3: 100,
	}
	for input, want := range expected {
		score, ok := scoreFromScale(input, 3)
		if !ok {
			t.Errorf("expected ok=true for input %d", input)
		}
		if score != want {
			t.Errorf("scoreFromScale(%d, 3) = %v, want %v", input, score, want)
		}
	}
}

func TestScoreFromScale_Likert4_AllValues(t *testing.T) {
	// Likert 4: 1=0%, 2=33.33%, 3=66.67%, 4=100%
	testCases := []struct {
		input int
		min   float64
		max   float64
	}{
		{1, 0, 0},
		{2, 33, 34},
		{3, 66, 67},
		{4, 100, 100},
	}
	for _, tc := range testCases {
		score, ok := scoreFromScale(tc.input, 4)
		if !ok {
			t.Errorf("expected ok=true for input %d", tc.input)
		}
		if score < tc.min || score > tc.max {
			t.Errorf("scoreFromScale(%d, 4) = %v, want between %v and %v", tc.input, score, tc.min, tc.max)
		}
	}
}

func TestScoreFromScale_Float64Input(t *testing.T) {
	score, ok := scoreFromScale(3.0, 5)
	if !ok {
		t.Error("expected ok=true for float64 input")
	}
	if score != 50 {
		t.Errorf("expected score=50, got %v", score)
	}
}

func TestScoreFromScale_StringInput(t *testing.T) {
	score, ok := scoreFromScale("4", 5)
	if !ok {
		t.Error("expected ok=true for string input")
	}
	if score != 75 {
		t.Errorf("expected score=75, got %v", score)
	}
}

func TestScoreFromScale_StringWithSpaces(t *testing.T) {
	score, ok := scoreFromScale(" 5 ", 5)
	if !ok {
		t.Error("expected ok=true for string with spaces")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromScale_OutOfRange(t *testing.T) {
	testCases := []struct {
		value interface{}
		max   int
	}{
		{0, 5},  // below min
		{6, 5},  // above max
		{-1, 5}, // negative
		{0, 3},  // below min for likert3
		{4, 3},  // above max for likert3
	}
	for _, tc := range testCases {
		_, ok := scoreFromScale(tc.value, tc.max)
		if ok {
			t.Errorf("expected ok=false for value=%v, max=%d", tc.value, tc.max)
		}
	}
}

func TestScoreFromScale_InvalidString(t *testing.T) {
	_, ok := scoreFromScale("abc", 5)
	if ok {
		t.Error("expected ok=false for non-numeric string")
	}
}

func TestScoreFromScale_UnsupportedType(t *testing.T) {
	_, ok := scoreFromScale([]int{1, 2, 3}, 5)
	if ok {
		t.Error("expected ok=false for slice type")
	}
}

// ============================================================================
// Tests for normalizeToHundred
// ============================================================================

func TestNormalizeToHundred_Likert5(t *testing.T) {
	testCases := []struct {
		value float64
		want  float64
	}{
		{1, 0},
		{2, 25},
		{3, 50},
		{4, 75},
		{5, 100},
	}
	for _, tc := range testCases {
		got := normalizeToHundred(tc.value, 5)
		if got != tc.want {
			t.Errorf("normalizeToHundred(%v, 5) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestNormalizeToHundred_MaxOne(t *testing.T) {
	// Edge case: max <= 1 should return 100
	got := normalizeToHundred(1, 1)
	if got != 100 {
		t.Errorf("normalizeToHundred(1, 1) = %v, want 100", got)
	}
}

func TestNormalizeToHundred_NegativeClamp(t *testing.T) {
	// Value below 1 should clamp to 0
	got := normalizeToHundred(0, 5)
	if got != 0 {
		t.Errorf("normalizeToHundred(0, 5) = %v, want 0", got)
	}
}

func TestNormalizeToHundred_OverflowClamp(t *testing.T) {
	// Value above max should clamp to 100
	got := normalizeToHundred(10, 5)
	if got != 100 {
		t.Errorf("normalizeToHundred(10, 5) = %v, want 100", got)
	}
}

// ============================================================================
// Tests for normalizeLegacyScore
// ============================================================================

func TestNormalizeLegacyScore_ValidRange(t *testing.T) {
	testCases := []struct {
		input float64
		want  float64
	}{
		{1, 0},
		{2, 25},
		{3, 50},
		{4, 75},
		{5, 100},
	}
	for _, tc := range testCases {
		got := normalizeLegacyScore(tc.input)
		if got != tc.want {
			t.Errorf("normalizeLegacyScore(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeLegacyScore_OutOfRange(t *testing.T) {
	// Values outside 0-5 should be returned as-is
	testCases := []float64{0, -1, 6, 50, 100}
	for _, tc := range testCases {
		got := normalizeLegacyScore(tc)
		if got != tc {
			t.Errorf("normalizeLegacyScore(%v) = %v, want %v (unchanged)", tc, got, tc)
		}
	}
}

// ============================================================================
// Tests for scoreFromQuestionValue
// ============================================================================

func TestScoreFromQuestionValue_YesNo(t *testing.T) {
	score, ok := scoreFromQuestionValue("ya", domain.QuestionYesNo)
	if !ok {
		t.Error("expected ok=true for YesNo question")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromQuestionValue_Likert(t *testing.T) {
	score, ok := scoreFromQuestionValue(5, domain.QuestionLikert)
	if !ok {
		t.Error("expected ok=true for Likert question")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromQuestionValue_LikertQuality(t *testing.T) {
	score, ok := scoreFromQuestionValue(3, domain.QuestionLikertQuality)
	if !ok {
		t.Error("expected ok=true for LikertQuality question")
	}
	if score != 50 {
		t.Errorf("expected score=50, got %v", score)
	}
}

func TestScoreFromQuestionValue_Likert3(t *testing.T) {
	score, ok := scoreFromQuestionValue(2, domain.QuestionLikert3)
	if !ok {
		t.Error("expected ok=true for Likert3 question")
	}
	if score != 50 {
		t.Errorf("expected score=50, got %v", score)
	}
}

func TestScoreFromQuestionValue_Likert3Puas(t *testing.T) {
	score, ok := scoreFromQuestionValue(3, domain.QuestionLikert3Puas)
	if !ok {
		t.Error("expected ok=true for Likert3Puas question")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromQuestionValue_Likert4(t *testing.T) {
	score, ok := scoreFromQuestionValue(4, domain.QuestionLikert4)
	if !ok {
		t.Error("expected ok=true for Likert4 question")
	}
	if score != 100 {
		t.Errorf("expected score=100, got %v", score)
	}
}

func TestScoreFromQuestionValue_Likert4Puas(t *testing.T) {
	score, ok := scoreFromQuestionValue(1, domain.QuestionLikert4Puas)
	if !ok {
		t.Error("expected ok=true for Likert4Puas question")
	}
	if score != 0 {
		t.Errorf("expected score=0, got %v", score)
	}
}

func TestScoreFromQuestionValue_UnsupportedType(t *testing.T) {
	// Text and MultipleChoice should return ok=false
	_, ok := scoreFromQuestionValue("some text", domain.QuestionText)
	if ok {
		t.Error("expected ok=false for Text question type")
	}

	_, ok = scoreFromQuestionValue("option1", domain.QuestionMultipleChoice)
	if ok {
		t.Error("expected ok=false for MultipleChoice question type")
	}
}
