package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"

	"gorm.io/gorm"
)

var reportLocationWIB = time.FixedZone("WIB", 7*60*60)

type ReportService struct {
	reports    *repository.ReportRepository
	categories *repository.CategoryRepository
	surveys    *repository.SurveyRepository
	now        func() time.Time
}

func NewReportService(
	reports *repository.ReportRepository,
	categories *repository.CategoryRepository,
	surveys *repository.SurveyRepository,
) *ReportService {
	return &ReportService{
		reports:    reports,
		categories: categories,
		surveys:    surveys,
		now:        time.Now,
	}
}

func (service *ReportService) CohortReport(period string, periods int) ([]domain.CohortRowDTO, error) {
	if periods <= 0 {
		periods = 5
	}
	unit := normalizePeriod(period)
	now := nowInWIB(service.now)
	end := periodStart(now, unit)
	start := addPeriods(end, unit, -(periods - 1))

	rows := make([]domain.CohortRowDTO, 0, periods)
	for i := 0; i < periods; i++ {
		cohortStart := addPeriods(start, unit, i)
		cohortEnd := addPeriods(cohortStart, unit, 1)

		responses, err := service.reports.ListSurveyResponsesByCreatedRange(cohortStart, cohortEnd)
		if err != nil {
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
			activeUsers, err := service.reports.ListActiveUsersInRange(users, periodStart, periodEnd)
			if err != nil {
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
	value = value.In(reportLocationWIB)
	switch unit {
	case "daily":
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, reportLocationWIB)
	case "weekly":
		weekday := int(value.Weekday())
		// Monday as start of week.
		offset := (weekday + 6) % 7
		start := value.AddDate(0, 0, -offset)
		return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, reportLocationWIB)
	case "yearly":
		return time.Date(value.Year(), 1, 1, 0, 0, 0, 0, reportLocationWIB)
	default:
		return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, reportLocationWIB)
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
			parsed := scoreFromRawAnswers(json.RawMessage(response.Answers))
			score = parsed
		}
		score = normalizeLegacyScore(score)
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

func (service *ReportService) ServiceTrends(start time.Time, end time.Time) ([]domain.ServiceTrendDTO, error) {
	rows, err := service.reports.ListTicketTotalsByCategory(start, end)
	if err != nil {
		return nil, err
	}

	var overall int64
	for _, item := range rows {
		overall += item.Total
	}
	if overall == 0 {
		return []domain.ServiceTrendDTO{}, nil
	}

	categories := service.categoryNameMap()

	trends := make([]domain.ServiceTrendDTO, 0, len(rows))
	for _, item := range rows {
		label := categories[item.CategoryID]
		if label == "" {
			label = item.CategoryID
		}
		trends = append(trends, domain.ServiceTrendDTO{
			Label:      label,
			Percentage: float64(item.Total) / float64(overall) * 100,
		})
	}

	return trends, nil
}

func (service *ReportService) DashboardSummary() (domain.DashboardSummaryDTO, error) {
	totalTickets, err := service.reports.CountTickets()
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	openTickets, err := service.reports.CountOpenTickets([]domain.TicketStatus{
		domain.StatusWaiting,
		domain.StatusInProgress,
	})
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	now := nowInWIB(service.now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, reportLocationWIB)
	resolvedThisMonth, err := service.reports.CountResolvedTicketsInRange(
		monthStart,
		monthStart.AddDate(0, 1, 0),
		domain.StatusResolved,
	)
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	avgScore, err := service.reports.AveragePositiveSurveyScore()
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}
	avgScore = normalizeLegacyScore(avgScore)

	return domain.DashboardSummaryDTO{
		TotalTickets:       int(totalTickets),
		OpenTickets:        int(openTickets),
		ResolvedThisPeriod: int(resolvedThisMonth),
		AvgRating:          avgScore,
	}, nil
}

func (service *ReportService) ServiceSatisfactionSummary(period string, periods int) ([]domain.ServiceSatisfactionDTO, error) {
	start, end := periodRange(period, periods, service.now)

	rows, err := service.reports.ListServiceSatisfactionRows(start, end)
	if err != nil {
		return nil, err
	}

	categories := service.categoryNameMap()

	totalWeighted := 0.0
	for _, row := range rows {
		row.AvgScore = normalizeLegacyScore(row.AvgScore)
		totalWeighted += row.AvgScore * float64(row.Responses)
	}

	result := make([]domain.ServiceSatisfactionDTO, 0, len(rows))
	for _, row := range rows {
		row.AvgScore = normalizeLegacyScore(row.AvgScore)
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

	template, err := service.resolveTemplate(categoryID, templateID, false)
	if err != nil {
		return nil, err
	}

	start, end := periodRange(period, periods, service.now)

	responses, err := service.reports.ListSurveyResponsesByTicketCategoryAndTemplate(
		start,
		end,
		categoryID,
		template.ID,
		false,
	)
	if err != nil {
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

	categoryName := "Semua Kategori"
	if categoryID != "" {
		categoryName = service.resolveCategoryName(categoryID)
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

func (service *ReportService) SurveySatisfactionExport(
	categoryID string,
	templateID string,
	period string,
	periods int,
) (*domain.SurveySatisfactionExportDTO, error) {
	if strings.TrimSpace(categoryID) == "" {
		return nil, errors.New("categoryId wajib diisi")
	}

	template, err := service.resolveTemplate(categoryID, templateID, true)
	if err != nil {
		return nil, err
	}

	start, end := periodRange(period, periods, service.now)

	responses, err := service.reports.ListSurveyResponsesByTicketCategoryAndTemplate(
		start,
		end,
		categoryID,
		template.ID,
		true,
	)
	if err != nil {
		return nil, err
	}

	questions := make([]domain.SurveySatisfactionExportQuestionDTO, 0, len(template.Questions))
	for _, question := range template.Questions {
		questions = append(questions, domain.SurveySatisfactionExportQuestionDTO{
			ID:   question.ID,
			Text: question.Text,
			Type: string(question.Type),
		})
	}

	responseDTOs := make([]domain.SurveySatisfactionExportResponseDTO, 0, len(responses))
	for _, response := range responses {
		responseDTOs = append(responseDTOs, domain.SurveySatisfactionExportResponseDTO{
			ID:        response.ID,
			TicketID:  response.TicketID,
			UserID:    response.UserID,
			Score:     response.Score,
			CreatedAt: response.CreatedAt,
			Answers:   response.Answers,
		})
	}

	categoryName := service.resolveCategoryName(categoryID)

	report := &domain.SurveySatisfactionExportDTO{
		TemplateID: template.ID,
		Template:   template.Title,
		CategoryID: categoryID,
		Category:   categoryName,
		Period:     normalizePeriod(period),
		Start:      start,
		End:        end,
		Questions:  questions,
		Responses:  responseDTOs,
	}
	return report, nil
}

func (service *ReportService) TemplatesByCategory(categoryID string) ([]domain.SurveyTemplateDTO, error) {
	if strings.TrimSpace(categoryID) == "" {
		return nil, errors.New("categoryId wajib diisi")
	}

	category, err := service.categories.FindByID(categoryID)
	if err != nil {
		return nil, err
	}

	templateIDs := map[string]struct{}{}
	if category.SurveyTemplateID != "" {
		templateIDs[category.SurveyTemplateID] = struct{}{}
	}

	usedIDs, err := service.reports.ListUsedTemplateIDsByCategory(categoryID)
	if err != nil {
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

	templates, err := service.reports.ListTemplatesByIDsWithQuestions(ids)
	if err != nil {
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

	return mapSurveyTemplates(templates), nil
}

func (service *ReportService) UsageCohort(period string, periods int) ([]domain.UsageCohortRowDTO, error) {
	if periods <= 0 {
		periods = 5
	}
	unit := normalizePeriod(period)
	now := nowInWIB(service.now)
	end := periodStart(now, unit)
	start := addPeriods(end, unit, -(periods - 1))

	rows := make([]domain.UsageCohortRowDTO, 0, periods)
	for i := 0; i < periods; i++ {
		windowStart := addPeriods(start, unit, i)
		windowEnd := addPeriods(windowStart, unit, 1)

		ticketCount, err := service.reports.CountTicketsInRange(windowStart, windowEnd)
		if err != nil {
			return nil, err
		}

		surveyCount, err := service.reports.CountSurveysInRange(windowStart, windowEnd)
		if err != nil {
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
	ticketCounts := make(map[string]map[string]int)
	surveyCounts := make(map[string]map[string]int)

	start, end := periodRange(period, periods, service.now)

	ticketRows, err := service.reports.ListRegisteredTicketRowsByEntityCategory(start, end)
	if err != nil {
		return nil, err
	}
	for _, item := range ticketRows {
		if ticketCounts[item.Entity] == nil {
			ticketCounts[item.Entity] = make(map[string]int)
		}
		ticketCounts[item.Entity][item.CategoryID] = item.Total
	}

	surveyRows, err := service.reports.ListRegisteredSurveyRowsByEntityCategory(start, end)
	if err != nil {
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
	entityRows, err := service.reports.ListRegisteredEntities()
	if err != nil {
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
	return service.reports.ListRegisteredCategories()
}

func periodRange(period string, periods int, nowFn func() time.Time) (time.Time, time.Time) {
	if periods <= 0 {
		periods = 5
	}
	unit := normalizePeriod(period)
	now := nowInWIB(nowFn)
	end := addPeriods(periodStart(now, unit), unit, 1)
	start := addPeriods(periodStart(now, unit), unit, -(periods - 1))
	return start, end
}

func nowInWIB(nowFn func() time.Time) time.Time {
	return nowFn().In(reportLocationWIB)
}

func (service *ReportService) resolveTemplate(
	categoryID string,
	templateID string,
	withOrdering bool,
) (*domain.SurveyTemplate, error) {
	selectedTemplateID := strings.TrimSpace(templateID)
	if selectedTemplateID == "" {
		category, err := service.categories.FindByID(categoryID)
		if err != nil {
			return nil, err
		}
		if category.SurveyTemplateID == "" {
			return nil, gorm.ErrRecordNotFound
		}
		selectedTemplateID = category.SurveyTemplateID
	}

	if withOrdering {
		return service.reports.FindTemplateWithOrderedQuestions(selectedTemplateID)
	}
	return service.surveys.FindByID(selectedTemplateID)
}

func (service *ReportService) categoryNameMap() map[string]string {
	categories := make(map[string]string)
	categoryRows, err := service.categories.List()
	if err == nil {
		for _, cat := range categoryRows {
			categories[cat.ID] = cat.Name
		}
	}
	return categories
}

func (service *ReportService) resolveCategoryName(categoryID string) string {
	category, err := service.categories.FindByID(categoryID)
	if err == nil && category.Name != "" {
		return category.Name
	}
	return categoryID
}
