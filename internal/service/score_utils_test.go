package service

import (
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

func TestScoreFromYesNo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{name: "bool true", input: true, expected: 100, ok: true},
		{name: "bool false", input: false, expected: 0, ok: true},
		{name: "string ya", input: "ya", expected: 100, ok: true},
		{name: "string no", input: "no", expected: 0, ok: true},
		{name: "string true mixed", input: " TrUe ", expected: 100, ok: true},
		{name: "string unknown", input: "maybe", expected: 0, ok: false},
		{name: "unsupported type", input: 1, expected: 0, ok: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			score, ok := scoreFromYesNo(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if score != tc.expected {
				t.Fatalf("expected score=%v, got %v", tc.expected, score)
			}
		})
	}
}

func TestScoreFromScale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    interface{}
		max      int
		expected float64
		ok       bool
	}{
		{name: "float value", input: 3.0, max: 5, expected: 50, ok: true},
		{name: "int value", input: 4, max: 5, expected: 75, ok: true},
		{name: "string value", input: "2", max: 5, expected: 25, ok: true},
		{name: "string whitespace", input: " 5 ", max: 5, expected: 100, ok: true},
		{name: "out of range", input: 0, max: 5, expected: 0, ok: false},
		{name: "too large", input: 6, max: 5, expected: 0, ok: false},
		{name: "bad string", input: "abc", max: 5, expected: 0, ok: false},
		{name: "unsupported type", input: true, max: 5, expected: 0, ok: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			score, ok := scoreFromScale(tc.input, tc.max)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if score != tc.expected {
				t.Fatalf("expected score=%v, got %v", tc.expected, score)
			}
		})
	}
}

func TestScoreFromQuestionValue(t *testing.T) {
	t.Parallel()

	score, ok := scoreFromQuestionValue("yes", domain.QuestionYesNo)
	if !ok || score != 100 {
		t.Fatalf("expected yes/no score 100 ok=true, got %v ok=%v", score, ok)
	}

	score, ok = scoreFromQuestionValue(2, domain.QuestionLikert3)
	if !ok || score != 50 {
		t.Fatalf("expected likert3 score 50 ok=true, got %v ok=%v", score, ok)
	}

	score, ok = scoreFromQuestionValue(1, domain.QuestionMultipleChoice)
	if ok || score != 0 {
		t.Fatalf("expected unsupported type score 0 ok=false, got %v ok=%v", score, ok)
	}
}

func TestNormalizeLegacyScore(t *testing.T) {
	t.Parallel()

	score := normalizeLegacyScore(4)
	if score != 75 {
		t.Fatalf("expected legacy score 75, got %v", score)
	}

	score = normalizeLegacyScore(90)
	if score != 90 {
		t.Fatalf("expected score to remain 90, got %v", score)
	}
}
