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
    case domain.QuestionLikert3:
        return scoreFromScale(value, 3)
    case domain.QuestionLikert4:
        return scoreFromScale(value, 4)
    case domain.QuestionLikert6:
        return scoreFromScale(value, 6)
    case domain.QuestionLikert7:
        return scoreFromScale(value, 7)
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
    return numeric, true
}
