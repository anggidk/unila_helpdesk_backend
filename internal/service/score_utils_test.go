package service

import (
	"math"
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

func TestScoreFromYesNo(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantScore float64
		wantOK    bool
	}{
		{name: "bool true", input: true, wantScore: 5, wantOK: true},
		{name: "bool false", input: false, wantScore: 1, wantOK: true},
		{name: "string ya", input: "Ya", wantScore: 5, wantOK: true},
		{name: "string yes with spaces", input: " yes ", wantScore: 5, wantOK: true},
		{name: "string tidak", input: "tidak", wantScore: 1, wantOK: true},
		{name: "string false uppercase", input: "FALSE", wantScore: 1, wantOK: true},
		{name: "unknown string", input: "maybe", wantScore: 0, wantOK: false},
		{name: "unsupported type", input: 1, wantScore: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := scoreFromYesNo(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if got != tc.wantScore {
				t.Fatalf("expected score=%v, got %v", tc.wantScore, got)
			}
		})
	}
}

func TestScoreFromScale(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		max       int
		wantScore float64
		wantOK    bool
	}{
		{name: "float64 value in range", input: float64(1), max: 3, wantScore: 1, wantOK: true},
		{name: "int value in range", input: 2, max: 3, wantScore: 3, wantOK: true},
		{name: "string value in range", input: "4", max: 5, wantScore: 4, wantOK: true},
		{name: "single point scale", input: 1, max: 1, wantScore: 5, wantOK: true},
		{name: "below lower bound", input: "0", max: 5, wantScore: 0, wantOK: false},
		{name: "above upper bound", input: "6", max: 5, wantScore: 0, wantOK: false},
		{name: "non numeric string", input: "abc", max: 5, wantScore: 0, wantOK: false},
		{name: "unsupported type", input: true, max: 5, wantScore: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := scoreFromScale(tc.input, tc.max)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if !almostEqualScore(got, tc.wantScore) {
				t.Fatalf("expected score=%v, got %v", tc.wantScore, got)
			}
		})
	}
}

func TestNormalizeToFive(t *testing.T) {
	tests := []struct {
		name   string
		value  float64
		max    int
		expect float64
	}{
		{name: "five point scale stays same", value: 3, max: 5, expect: 3},
		{name: "three point midpoint becomes three", value: 2, max: 3, expect: 3},
		{name: "four point max becomes five", value: 4, max: 4, expect: 5},
		{name: "lower bound is clamped to one", value: 0, max: 5, expect: 1},
		{name: "upper bound is clamped to five", value: 6, max: 5, expect: 5},
		{name: "degenerate scale returns five", value: 1, max: 1, expect: 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeToFive(tc.value, tc.max); !almostEqualScore(got, tc.expect) {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestScoreFromQuestionValue(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		questionType domain.SurveyQuestionType
		wantScore    float64
		wantOK       bool
	}{
		{name: "dispatches yes no", input: "ya", questionType: domain.QuestionYesNo, wantScore: 5, wantOK: true},
		{name: "dispatches likert4", input: 2, questionType: domain.QuestionLikert4, wantScore: 2.333333333333333, wantOK: true},
		{name: "dispatches likert5 string", input: "4", questionType: domain.QuestionLikert5, wantScore: 4, wantOK: true},
		{name: "text is unsupported", input: "free form", questionType: domain.QuestionText, wantScore: 0, wantOK: false},
		{name: "multiple choice is unsupported", input: "A", questionType: domain.QuestionMultipleChoice, wantScore: 0, wantOK: false},
		{name: "unknown type is unsupported", input: 1, questionType: domain.SurveyQuestionType("custom"), wantScore: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := scoreFromQuestionValue(tc.input, tc.questionType)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if !almostEqualScore(got, tc.wantScore) {
				t.Fatalf("expected score=%v, got %v", tc.wantScore, got)
			}
		})
	}
}

func almostEqualScore(left float64, right float64) bool {
	return math.Abs(left-right) < 1e-9
}
