package service

import (
	"strconv"
	"strings"

	"unila_helpdesk_backend/internal/domain"
)

func scoreFromQuestionValue(
	value interface{},
	questionType domain.SurveyQuestionType,
) (float64, bool) {
	switch questionType {
	case domain.QuestionYesNo:
		return scoreFromYesNo(value)
	case domain.QuestionLikert3:
		return scoreFromScale(value, 3)
	case domain.QuestionLikert4:
		return scoreFromScale(value, 4)
	case domain.QuestionLikert5:
		return scoreFromScale(value, 5)
	default:
		return 0, false
	}
}

func scoreFromYesNo(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case bool:
		if v {
			return 5, true
		}
		return 1, true
	case string:
		cleaned := strings.ToLower(strings.TrimSpace(v))
		if cleaned == "ya" || cleaned == "yes" || cleaned == "true" {
			return 5, true
		}
		if cleaned == "tidak" || cleaned == "no" || cleaned == "false" {
			return 1, true
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
	return normalizeToFive(numeric, max), true
}

func normalizeToFive(value float64, max int) float64 {
	if max <= 1 {
		return 5
	}
	// Map every ordinal scale (2/3/4/5) into the same 1..5 basis.
	normalized := 1 + ((value-1)*4)/float64(max-1)
	if normalized < 1 {
		return 1
	}
	if normalized > 5 {
		return 5
	}
	return normalized
}

func scoreToFivePoint(score float64) float64 {
	if score <= 0 {
		return 0
	}
	if score <= 5 {
		return score
	}
	normalized := score / 20
	if normalized < 0 {
		return 0
	}
	if normalized > 5 {
		return 5
	}
	return normalized
}
