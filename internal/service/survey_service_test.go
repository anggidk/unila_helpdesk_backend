package service

import (
	"math"
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

func TestCalculateSurveyScore(t *testing.T) {
	template := &domain.SurveyTemplate{
		ID: "tmpl-1",
		Questions: []domain.SurveyQuestion{
			{ID: "q_yesno", Type: domain.QuestionYesNo},
			{ID: "q_scale", Type: domain.QuestionLikert4},
			{ID: "q_text", Type: domain.QuestionText},
		},
	}

	tests := []struct {
		name     string
		answers  map[string]interface{}
		template *domain.SurveyTemplate
		expect   float64
	}{
		{
			name:     "nil template returns zero",
			answers:  map[string]interface{}{"q_yesno": true},
			template: nil,
			expect:   0,
		},
		{
			name:     "empty answers returns zero",
			answers:  map[string]interface{}{},
			template: template,
			expect:   0,
		},
		{
			name:     "template without questions returns zero",
			answers:  map[string]interface{}{"q_yesno": true},
			template: &domain.SurveyTemplate{},
			expect:   0,
		},
		{
			name: "averages only scorable answers",
			answers: map[string]interface{}{
				"q_yesno": true,
				"q_scale": 2,
				"q_text":  "komentar bebas",
				"extra":   5,
			},
			template: template,
			expect:   (5 + 2.333333333333333) / 2,
		},
		{
			name: "partial answers average only answered scorable questions",
			answers: map[string]interface{}{
				"q_yesno": false,
			},
			template: template,
			expect:   1,
		},
		{
			name: "all invalid answers return zero",
			answers: map[string]interface{}{
				"q_yesno": "mungkin",
				"q_scale": "invalid",
				"q_text":  "komentar",
			},
			template: template,
			expect:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateSurveyScore(tc.answers, tc.template)
			if !almostEqualSurveyScore(got, tc.expect) {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func almostEqualSurveyScore(left float64, right float64) bool {
	return math.Abs(left-right) < 1e-9
}
