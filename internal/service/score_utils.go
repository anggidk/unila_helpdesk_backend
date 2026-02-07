package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"unila_helpdesk_backend/internal/domain"
)

func scoreFromQuestionValue(value interface{}, questionType domain.SurveyQuestionType) (float64, bool) {
	switch questionType {
	case domain.QuestionYesNo:
		return scoreFromYesNo(value)
	case domain.QuestionLikert:
		return scoreFromScale(value, 5)
	case domain.QuestionLikertQuality:
		return scoreFromScale(value, 5)
	case domain.QuestionLikert3Puas:
		return scoreFromScale(value, 3)
	case domain.QuestionLikert3:
		return scoreFromScale(value, 3)
	case domain.QuestionLikert4Puas:
		return scoreFromScale(value, 4)
	case domain.QuestionLikert4:
		return scoreFromScale(value, 4)
	default:
		return 0, false
	}
}

func scoreFromYesNo(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case bool:
		if v {
			return 100, true
		}
		return 0, true
	case string:
		cleaned := strings.ToLower(strings.TrimSpace(v))
		if cleaned == "ya" || cleaned == "yes" || cleaned == "true" {
			return 100, true
		}
		if cleaned == "tidak" || cleaned == "no" || cleaned == "false" {
			return 0, true
		}
	}
	return 0, false
}

func scoreFromScale(value interface{}, max int) (float64, bool) {
	var numeric float64
	switch v := value.(type) {
	case float64:
		numeric = v
	case int:
		numeric = float64(v)
	case string:
		cleaned := strings.TrimSpace(v)
		parsed, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return 0, false
		}
		numeric = parsed
	default:
		return 0, false
	}
	if numeric < 1 || numeric > float64(max) {
		return 0, false
	}
	return normalizeToHundred(numeric, max), true
}

func normalizeToHundred(value float64, max int) float64 {
	if max <= 1 {
		return 100
	}
	// Map the minimum Likert choice to 1 star-equivalent (20 on a 0-100 scale),
	// and the maximum choice to 5 stars-equivalent (100).
	normalized := 20 + ((value-1)*80)/float64(max-1)
	if normalized < 0 {
		return 0
	}
	if normalized > 100 {
		return 100
	}
	return normalized
}

func normalizeLegacyScore(score float64) float64 {
	if score > 0 && score <= 5 {
		return normalizeToHundred(score, 5)
	}
	return score
}

func scoreFromRawAnswers(raw json.RawMessage) float64 {
	var answers map[string]interface{}
	if err := json.Unmarshal(raw, &answers); err != nil {
		return 0
	}
	return calculateLegacyScore(answers)
}

func calculateLegacyScore(answers map[string]interface{}) float64 {
	if len(answers) == 0 {
		return 0
	}
	var total float64
	var count int
	for _, value := range answers {
		if score, ok := scoreFromLegacyValue(value); ok {
			total += score
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func scoreFromLegacyValue(value interface{}) (float64, bool) {
	if score, ok := scoreFromScale(value, 5); ok {
		return score, true
	}
	if score, ok := scoreFromYesNo(value); ok {
		return score, true
	}
	return 0, false
}
