package service

import (
    "encoding/json"
    "errors"
    "sort"
    "strconv"
    "strings"
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

func (service *ReportService) CohortReport(period string, periods int) ([]domain.CohortRowDTO, error) {
    if periods <= 0 {
        periods = 5
    }
    unit := normalizePeriod(period)
    now := service.now().UTC()
    end := periodStart(now, unit)
    start := addPeriods(end, unit, -(periods-1))

    rows := make([]domain.CohortRowDTO, 0, periods)
    for i := 0; i < periods; i++ {
        cohortStart := addPeriods(start, unit, i)
        cohortEnd := addPeriods(cohortStart, unit, 1)

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
        retention := make([]int, 0, periods)
        if cohortSize == 0 {
            retention = make([]int, periods)
            rows = append(rows, domain.CohortRowDTO{
                Label:        formatCohortLabel(cohortStart, unit),
                Users:        0,
                Retention:    retention,
                AvgScore:     0,
                ResponseRate: 0,
            })
            continue
        }

        avgScore, responseRate := calculateCohortScores(responses)

        retention = append(retention, 100)
        for step := 1; step < periods; step++ {
            periodStart := addPeriods(cohortStart, unit, step)
            periodEnd := addPeriods(periodStart, unit, 1)
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
            Label:        formatCohortLabel(cohortStart, unit),
            Users:        cohortSize,
            Retention:    retention,
            AvgScore:     avgScore,
            ResponseRate: responseRate,
        })
    }

    return rows, nil
}

func normalizePeriod(value string) string {
    switch strings.ToLower(strings.TrimSpace(value)) {
    case "daily":
        return "daily"
    case "weekly":
        return "weekly"
    case "yearly":
        return "yearly"
    default:
        return "monthly"
    }
}

func periodStart(value time.Time, unit string) time.Time {
    value = value.UTC()
    switch unit {
    case "daily":
        return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
    case "weekly":
        weekday := int(value.Weekday())
        // Monday as start of week.
        offset := (weekday + 6) % 7
        start := value.AddDate(0, 0, -offset)
        return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
    case "yearly":
        return time.Date(value.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
    default:
        return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
    }
}

func addPeriods(value time.Time, unit string, count int) time.Time {
    switch unit {
    case "daily":
        return value.AddDate(0, 0, count)
    case "weekly":
        return value.AddDate(0, 0, 7*count)
    case "yearly":
        return value.AddDate(count, 0, 0)
    default:
        return value.AddDate(0, count, 0)
    }
}

func formatCohortLabel(value time.Time, unit string) string {
    switch unit {
    case "daily":
        return value.Format("02 Jan 2006")
    case "weekly":
        return "Week of " + value.Format("02 Jan 2006")
    case "yearly":
        return value.Format("2006")
    default:
        return value.Format("Jan 2006")
    }
}

func calculateCohortScores(responses []domain.SurveyResponse) (float64, float64) {
    if len(responses) == 0 {
        return 0, 0
    }

    var total float64
    var count int
    var responsesWithScore int

    for _, response := range responses {
        score := response.Score
        if score <= 0 {
            parsed := scoreFromAnswers(json.RawMessage(response.Answers))
            score = parsed
        }
        if score > 0 {
            total += score
            count++
            responsesWithScore++
        }
    }

    avg := 0.0
    if count > 0 {
        avg = total / float64(count)
    }
    responseRate := float64(responsesWithScore) / float64(len(responses)) * 100
    return avg, responseRate
}

func scoreFromAnswers(raw json.RawMessage) float64 {
    var answers map[string]interface{}
    if err := json.Unmarshal(raw, &answers); err != nil {
        return 0
    }
    var total float64
    var count int
    for _, value := range answers {
        if score, ok := scoreFromValue(value); ok {
            total += score
            count++
        }
    }
    if count == 0 {
        return 0
    }
    return total / float64(count)
}

func scoreFromValue(value interface{}) (float64, bool) {
    switch v := value.(type) {
    case float64:
        if v >= 1 && v <= 5 {
            return v, true
        }
    case int:
        if v >= 1 && v <= 5 {
            return float64(v), true
        }
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
        if parsed, err := strconv.ParseFloat(cleaned, 64); err == nil {
            if parsed >= 1 && parsed <= 5 {
                return parsed, true
            }
        }
    }
    return 0, false
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

func (service *ReportService) DashboardSummary() (domain.DashboardSummaryDTO, error) {
    var totalTickets int64
    if err := service.db.Model(&domain.Ticket{}).Count(&totalTickets).Error; err != nil {
        return domain.DashboardSummaryDTO{}, err
    }

    var openTickets int64
    if err := service.db.Model(&domain.Ticket{}).
        Where("status != ?", domain.StatusResolved).
        Count(&openTickets).Error; err != nil {
        return domain.DashboardSummaryDTO{}, err
    }

    now := service.now().UTC()
    monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    var resolvedThisMonth int64
    if err := service.db.Model(&domain.Ticket{}).
        Where("status = ?", domain.StatusResolved).
        Where("updated_at >= ? AND updated_at < ?", monthStart, monthStart.AddDate(0, 1, 0)).
        Count(&resolvedThisMonth).Error; err != nil {
        return domain.DashboardSummaryDTO{}, err
    }

    var avgScore float64
    if err := service.db.Model(&domain.SurveyResponse{}).
        Where("score > 0").
        Select("COALESCE(AVG(score), 0)").
        Scan(&avgScore).Error; err != nil {
        return domain.DashboardSummaryDTO{}, err
    }

    return domain.DashboardSummaryDTO{
        TotalTickets:       int(totalTickets),
        OpenTickets:        int(openTickets),
        ResolvedThisPeriod: int(resolvedThisMonth),
        AvgRating:          avgScore,
    }, nil
}

func (service *ReportService) ServiceSatisfactionSummary(period string, periods int) ([]domain.ServiceSatisfactionDTO, error) {
    start, end := periodRange(period, periods, service.now)

    type row struct {
        CategoryID string
        AvgScore   float64
        Responses  int
    }
    var rows []row
    if err := service.db.Raw(`
        SELECT t.category_id AS category_id,
               COALESCE(AVG(sr.score), 0) AS avg_score,
               COUNT(*) AS responses
        FROM survey_responses sr
        JOIN tickets t ON t.id = sr.ticket_id
        WHERE sr.created_at >= ? AND sr.created_at < ?
          AND sr.score > 0
        GROUP BY t.category_id
    `, start, end).Scan(&rows).Error; err != nil {
        return nil, err
    }

    categories := make(map[string]string)
    var categoryRows []domain.ServiceCategory
    if err := service.db.Find(&categoryRows).Error; err == nil {
        for _, cat := range categoryRows {
            categories[cat.ID] = cat.Name
        }
    }

    totalWeighted := 0.0
    for _, row := range rows {
        totalWeighted += row.AvgScore * float64(row.Responses)
    }

    result := make([]domain.ServiceSatisfactionDTO, 0, len(rows))
    for _, row := range rows {
        label := categories[row.CategoryID]
        if label == "" {
            label = row.CategoryID
        }
        percentage := 0.0
        if totalWeighted > 0 {
            percentage = (row.AvgScore * float64(row.Responses)) / totalWeighted * 100
        }
        result = append(result, domain.ServiceSatisfactionDTO{
            CategoryID: row.CategoryID,
            Label:      label,
            AvgScore:   row.AvgScore,
            Responses:  row.Responses,
            Percentage: percentage,
        })
    }

    return result, nil
}

func (service *ReportService) SurveySatisfaction(
    categoryID string,
    templateID string,
    period string,
    periods int,
) (*domain.SurveySatisfactionDTO, error) {
    if categoryID == "" && templateID == "" {
        return nil, gorm.ErrRecordNotFound
    }

    var template domain.SurveyTemplate
    if templateID != "" {
        if err := service.db.Preload("Questions").First(&template, "id = ?", templateID).Error; err != nil {
            return nil, err
        }
    } else {
        var category domain.ServiceCategory
        if err := service.db.First(&category, "id = ?", categoryID).Error; err != nil {
            return nil, err
        }
        if category.SurveyTemplateID == "" {
            return nil, gorm.ErrRecordNotFound
        }
        if err := service.db.Preload("Questions").First(&template, "id = ?", category.SurveyTemplateID).Error; err != nil {
            return nil, err
        }
    }

    start, end := periodRange(period, periods, service.now)

    var responses []domain.SurveyResponse
    query := service.db.Model(&domain.SurveyResponse{}).
        Joins("JOIN tickets t ON t.id = survey_responses.ticket_id").
        Where("survey_responses.created_at >= ? AND survey_responses.created_at < ?", start, end)
    if categoryID != "" {
        query = query.Where("t.category_id = ?", categoryID)
    }
    if template.ID != "" {
        query = query.Where("survey_responses.template_id = ?", template.ID)
    }
    if err := query.Find(&responses).Error; err != nil {
        return nil, err
    }

    sums := make(map[string]float64)
    scoreCounts := make(map[string]int)
    answerCounts := make(map[string]int)
    for _, response := range responses {
        var answers map[string]interface{}
        if err := json.Unmarshal(response.Answers, &answers); err != nil {
            continue
        }
        for _, question := range template.Questions {
            value, ok := answers[question.ID]
            if !ok {
                continue
            }
            answerCounts[question.ID]++
            if score, ok := scoreFromQuestionValue(value, question.Type); ok {
                sums[question.ID] += score
                scoreCounts[question.ID]++
            }
        }
    }

    rows := make([]domain.SurveySatisfactionRowDTO, 0, len(template.Questions))
    for _, question := range template.Questions {
        responsesCount := answerCounts[question.ID]
        avgScore := 0.0
        if scoreCounts[question.ID] > 0 {
            avgScore = sums[question.ID] / float64(scoreCounts[question.ID])
        }
        rows = append(rows, domain.SurveySatisfactionRowDTO{
            QuestionID: question.ID,
            Question:   question.Text,
            Type:       string(question.Type),
            AvgScore:   avgScore,
            Responses:  responsesCount,
        })
    }

    categoryName := categoryID
    if categoryID == "" {
        categoryName = "Semua Kategori"
    } else {
        var category domain.ServiceCategory
        if err := service.db.First(&category, "id = ?", categoryID).Error; err == nil {
            if category.Name != "" {
                categoryName = category.Name
            }
        }
    }

    report := &domain.SurveySatisfactionDTO{
        TemplateID: template.ID,
        Template:   template.Title,
        CategoryID: categoryID,
        Category:   categoryName,
        Period:     normalizePeriod(period),
        Start:      start,
        End:        end,
        Rows:       rows,
    }
    return report, nil
}

func (service *ReportService) TemplatesByCategory(categoryID string) ([]domain.SurveyTemplateDTO, error) {
    if strings.TrimSpace(categoryID) == "" {
        return nil, errors.New("categoryId wajib diisi")
    }

    var category domain.ServiceCategory
    if err := service.db.First(&category, "id = ?", categoryID).Error; err != nil {
        return nil, err
    }

    templateIDs := map[string]struct{}{}
    if category.SurveyTemplateID != "" {
        templateIDs[category.SurveyTemplateID] = struct{}{}
    }

    var usedIDs []string
    if err := service.db.Model(&domain.SurveyResponse{}).
        Joins("JOIN tickets t ON t.id = survey_responses.ticket_id").
        Where("t.category_id = ? AND survey_responses.template_id <> ''", categoryID).
        Distinct().
        Pluck("survey_responses.template_id", &usedIDs).Error; err != nil {
        return nil, err
    }
    for _, id := range usedIDs {
        if id != "" {
            templateIDs[id] = struct{}{}
        }
    }

    if len(templateIDs) == 0 {
        return []domain.SurveyTemplateDTO{}, nil
    }

    ids := make([]string, 0, len(templateIDs))
    for id := range templateIDs {
        ids = append(ids, id)
    }

    var templates []domain.SurveyTemplate
    if err := service.db.Preload("Questions").
        Where("id IN ?", ids).
        Find(&templates).Error; err != nil {
        return nil, err
    }

    sort.Slice(templates, func(i, j int) bool {
        if templates[i].ID == category.SurveyTemplateID {
            return true
        }
        if templates[j].ID == category.SurveyTemplateID {
            return false
        }
        return templates[i].UpdatedAt.After(templates[j].UpdatedAt)
    })

    return mapSurveyTemplateDTOs(templates), nil
}

func (service *ReportService) UsageCohort(period string, periods int) ([]domain.UsageCohortRowDTO, error) {
    if periods <= 0 {
        periods = 5
    }
    unit := normalizePeriod(period)
    now := service.now().UTC()
    end := periodStart(now, unit)
    start := addPeriods(end, unit, -(periods-1))

    rows := make([]domain.UsageCohortRowDTO, 0, periods)
    for i := 0; i < periods; i++ {
        windowStart := addPeriods(start, unit, i)
        windowEnd := addPeriods(windowStart, unit, 1)

        var ticketCount int64
        if err := service.db.Model(&domain.Ticket{}).
            Where("created_at >= ? AND created_at < ?", windowStart, windowEnd).
            Count(&ticketCount).Error; err != nil {
            return nil, err
        }

        var surveyCount int64
        if err := service.db.Model(&domain.SurveyResponse{}).
            Where("created_at >= ? AND created_at < ?", windowStart, windowEnd).
            Count(&surveyCount).Error; err != nil {
            return nil, err
        }

        rows = append(rows, domain.UsageCohortRowDTO{
            Label:   formatCohortLabel(windowStart, unit),
            Tickets: int(ticketCount),
            Surveys: int(surveyCount),
        })
    }
    return rows, nil
}

func (service *ReportService) EntityServiceMatrix(period string, periods int) ([]domain.EntityServiceDTO, error) {
    type row struct {
        Entity     string
        CategoryID string
        Total      int
    }

    ticketCounts := make(map[string]map[string]int)
    surveyCounts := make(map[string]map[string]int)

    start, end := periodRange(period, periods, service.now)

    var ticketRows []row
    if err := service.db.Raw(`
        SELECT u.entity AS entity, t.category_id AS category_id, COUNT(*) AS total
        FROM tickets t
        JOIN users u ON u.id = t.reporter_id
        WHERE u.role = 'registered'
          AND t.created_at >= ? AND t.created_at < ?
        GROUP BY u.entity, t.category_id
    `, start, end).Scan(&ticketRows).Error; err != nil {
        return nil, err
    }
    for _, item := range ticketRows {
        if ticketCounts[item.Entity] == nil {
            ticketCounts[item.Entity] = make(map[string]int)
        }
        ticketCounts[item.Entity][item.CategoryID] = item.Total
    }

    var surveyRows []row
    if err := service.db.Raw(`
        SELECT u.entity AS entity, t.category_id AS category_id, COUNT(*) AS total
        FROM survey_responses sr
        JOIN users u ON u.id = sr.user_id
        JOIN tickets t ON t.id = sr.ticket_id
        WHERE u.role = 'registered'
          AND sr.created_at >= ? AND sr.created_at < ?
        GROUP BY u.entity, t.category_id
    `, start, end).Scan(&surveyRows).Error; err != nil {
        return nil, err
    }
    for _, item := range surveyRows {
        if surveyCounts[item.Entity] == nil {
            surveyCounts[item.Entity] = make(map[string]int)
        }
        surveyCounts[item.Entity][item.CategoryID] = item.Total
    }

	categories, err := service.listRegisteredCategories()
	if err != nil {
		return nil, err
	}

	entities := make(map[string]struct{})
	for entity := range ticketCounts {
		entities[entity] = struct{}{}
	}
	for entity := range surveyCounts {
		entities[entity] = struct{}{}
	}
	var entityRows []string
	if err := service.db.Model(&domain.User{}).
		Distinct("entity").
		Where("role = ?", domain.RoleRegistered).
		Where("entity <> ''").
		Order("entity asc").
		Pluck("entity", &entityRows).Error; err != nil {
		return nil, err
	}
	for _, entity := range entityRows {
		entities[entity] = struct{}{}
	}

    rows := make([]domain.EntityServiceDTO, 0)
    for entity := range entities {
        for _, cat := range categories {
            rows = append(rows, domain.EntityServiceDTO{
                Entity:     entity,
                CategoryID: cat.ID,
                Category:   cat.Name,
                Tickets:    ticketCounts[entity][cat.ID],
                Surveys:    surveyCounts[entity][cat.ID],
            })
        }
    }
    return rows, nil
}

func (service *ReportService) listRegisteredCategories() ([]domain.ServiceCategory, error) {
    var categories []domain.ServiceCategory
    if err := service.db.Model(&domain.ServiceCategory{}).
        Where("guest_allowed = ?", false).
        Order("name asc").
        Find(&categories).Error; err != nil {
        return nil, err
    }
    hidden := map[string]struct{}{
        "membership":    {},
        "guest-account": {},
        "guest-email":   {},
        "email":         {},
        "vclass":        {},
    }
    filtered := make([]domain.ServiceCategory, 0, len(categories))
    for _, cat := range categories {
        if _, skip := hidden[cat.ID]; skip {
            continue
        }
        filtered = append(filtered, cat)
    }
    return filtered, nil
}

func periodRange(period string, periods int, nowFn func() time.Time) (time.Time, time.Time) {
    if periods <= 0 {
        periods = 5
    }
    unit := normalizePeriod(period)
    now := nowFn().UTC()
    end := addPeriods(periodStart(now, unit), unit, 1)
    start := addPeriods(periodStart(now, unit), unit, -(periods - 1))
    return start, end
}

func mapSurveyTemplateDTOs(templates []domain.SurveyTemplate) []domain.SurveyTemplateDTO {
    result := make([]domain.SurveyTemplateDTO, 0, len(templates))
    for _, template := range templates {
        result = append(result, mapSurveyTemplateDTO(template))
    }
    return result
}

func mapSurveyTemplateDTO(template domain.SurveyTemplate) domain.SurveyTemplateDTO {
    questions := make([]domain.SurveyQuestionDTO, 0, len(template.Questions))
    for _, question := range template.Questions {
        var options []string
        _ = json.Unmarshal(question.Options, &options)
        questions = append(questions, domain.SurveyQuestionDTO{
            ID:      question.ID,
            Text:    question.Text,
            Type:    string(question.Type),
            Options: options,
        })
    }
    return domain.SurveyTemplateDTO{
        ID:          template.ID,
        Title:       template.Title,
        Description: template.Description,
        CategoryID:  template.CategoryID,
        Questions:   questions,
        CreatedAt:   template.CreatedAt,
        UpdatedAt:   template.UpdatedAt,
    }
}
