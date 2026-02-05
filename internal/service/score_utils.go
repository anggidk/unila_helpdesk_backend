package service

import (
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
	normalized := ((value - 1) * 100) / float64(max-1)
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
