package service

import (
    "encoding/json"
    "strconv"
    "time"

    "unila_helpdesk_backend/internal/domain"

    "gorm.io/gorm"
)

type ReportService struct {
    db  *gorm.DB
    now func() time.Time
}

func NewReportService(db *gorm.DB) *ReportService {
    return &ReportService{db: db, now: time.Now}
}

func (service *ReportService) CohortReport(months int) ([]domain.CohortRowDTO, error) {
    if months <= 0 {
        months = 5
    }
    end := time.Date(service.now().Year(), service.now().Month(), 1, 0, 0, 0, 0, time.UTC)
    start := end.AddDate(0, -(months-1), 0)

    rows := make([]domain.CohortRowDTO, 0, months)
    for m := 0; m < months; m++ {
        cohortStart := start.AddDate(0, m, 0)
        cohortEnd := cohortStart.AddDate(0, 1, 0)

        var responses []domain.SurveyResponse
        if err := service.db.Model(&domain.SurveyResponse{}).
            Where("created_at >= ? AND created_at < ?", cohortStart, cohortEnd).
            Find(&responses).Error; err != nil {
            return nil, err
        }

        userSet := make(map[string]struct{})
        for _, response := range responses {
            userSet[response.UserID] = struct{}{}
        }

        users := make([]string, 0, len(userSet))
        for id := range userSet {
            users = append(users, id)
        }

        cohortSize := len(users)
        retention := make([]int, 0, 5)
        if cohortSize == 0 {
            retention = []int{0, 0, 0, 0, 0}
            rows = append(rows, domain.CohortRowDTO{
                Label:        cohortStart.Format("Jan 2006"),
                Users:        0,
                Retention:    retention,
                AvgScore:     0,
                ResponseRate: 0,
            })
            continue
        }

        avgScore, responseRate := calculateCohortScores(responses)

        retention = append(retention, 100)
        for i := 1; i < 5; i++ {
            periodStart := cohortStart.AddDate(0, i, 0)
            periodEnd := periodStart.AddDate(0, 1, 0)
            var activeUsers []string
            if err := service.db.Model(&domain.SurveyResponse{}).
                Where("user_id IN ?", users).
                Where("created_at >= ? AND created_at < ?", periodStart, periodEnd).
                Distinct().
                Pluck("user_id", &activeUsers).Error; err != nil {
                return nil, err
            }
            percent := int(float64(len(activeUsers)) / float64(cohortSize) * 100)
            retention = append(retention, percent)
        }

        rows = append(rows, domain.CohortRowDTO{
            Label:        cohortStart.Format("Jan 2006"),
            Users:        cohortSize,
            Retention:    retention,
            AvgScore:     avgScore,
            ResponseRate: responseRate,
        })
    }

    return rows, nil
}

func calculateCohortScores(responses []domain.SurveyResponse) (float64, float64) {
    if len(responses) == 0 {
        return 0, 0
    }

    var total float64
    var count int
    var responsesWithLikert int

    for _, response := range responses {
        var answers map[string]interface{}
        if err := json.Unmarshal(response.Answers, &answers); err != nil {
            continue
        }
        hasLikert := false
        for _, value := range answers {
            switch v := value.(type) {
            case float64:
                total += v
                count++
                hasLikert = true
            case int:
                total += float64(v)
                count++
                hasLikert = true
            case string:
                if parsed, err := strconv.ParseFloat(v, 64); err == nil {
                    total += parsed
                    count++
                    hasLikert = true
                }
            }
        }
        if hasLikert {
            responsesWithLikert++
        }
    }

    avg := 0.0
    if count > 0 {
        avg = total / float64(count)
    }
    responseRate := float64(responsesWithLikert) / float64(len(responses)) * 100
    return avg, responseRate
}

func (service *ReportService) ServiceTrends(start time.Time, end time.Time) ([]domain.ServiceTrendDTO, error) {
    type row struct {
        CategoryID string
        Total      int64
    }
    var rows []row
    if err := service.db.Model(&domain.Ticket{}).
        Select("category_id as category_id, count(*) as total").
        Where("created_at >= ? AND created_at < ?", start, end).
        Group("category_id").
        Order("total desc").
        Find(&rows).Error; err != nil {
        return nil, err
    }

    var overall int64
    for _, item := range rows {
        overall += item.Total
    }
    if overall == 0 {
        return []domain.ServiceTrendDTO{}, nil
    }

    categories := make(map[string]string)
    var categoryRows []domain.ServiceCategory
    if err := service.db.Find(&categoryRows).Error; err == nil {
        for _, cat := range categoryRows {
            categories[cat.ID] = cat.Name
        }
    }

    trends := make([]domain.ServiceTrendDTO, 0, len(rows))
    for _, item := range rows {
        label := categories[item.CategoryID]
        if label == "" {
            label = item.CategoryID
        }
        trends = append(trends, domain.ServiceTrendDTO{
            Label:      label,
            Percentage: float64(item.Total) / float64(overall) * 100,
            Note:       "",
        })
    }

    return trends, nil
}
