package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
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

func defaultCohortLookback(unit string) int {
	switch unit {
	case "daily":
		return 30
	case "weekly":
		return 12
	case "yearly":
		return 5
	default:
		return 6
	}
}

func defaultCohortBuckets(unit string) int {
	switch unit {
	case "daily":
		return 10
	case "weekly":
		return 8
	case "yearly":
		return 5
	default:
		return 6
	}
}

func normalizeCohortLookback(unit string, value int) int {
	if value <= 0 {
		return defaultCohortLookback(unit)
	}
	return value
}

func normalizeCohortBuckets(unit string, value int) int {
	if value <= 0 {
		return defaultCohortBuckets(unit)
	}
	return value
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
		domain.StatusAssign,
	})
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	now := nowInWIB(service.now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, reportLocationWIB)
	resolvedThisMonth, err := service.reports.CountResolvedTicketsInRange(
		monthStart,
		monthStart.AddDate(0, 1, 0),
		domain.StatusDone,
	)
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	avgScore, err := service.reports.AveragePositiveSurveyScore()
	if err != nil {
		return domain.DashboardSummaryDTO{}, err
	}

	return domain.DashboardSummaryDTO{
		TotalTickets:       int(totalTickets),
		OpenTickets:        int(openTickets),
		ResolvedThisPeriod: int(resolvedThisMonth),
		AvgRating:          scoreToFivePoint(avgScore),
	}, nil
}

func (service *ReportService) ServiceSatisfactionSummary(period string, periods int) ([]domain.ServiceSatisfactionDTO, error) {
	start, end := rollingReportRange(period, service.now)

	rows, err := service.reports.ListServiceSatisfactionRows(start, end)
	if err != nil {
		return nil, err
	}

	categories := service.categoryNameMap()

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
			AvgScore:   scoreToFivePoint(row.AvgScore),
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

	start, end := rollingReportRange(period, service.now)

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

	responseItems, err := service.reports.ListResponseItemsByResponseIDs(surveyResponseIDs(responses))
	if err != nil {
		return nil, err
	}
	itemsByResponse := groupResponseItemsByResponseID(responseItems)
	questionMetas := make(map[string]domain.SurveyQuestionType, len(template.Questions))
	for _, question := range template.Questions {
		questionMetas[question.ID] = question.Type
	}

	sums := make(map[string]float64)
	scoreCounts := make(map[string]int)
	answerCounts := make(map[string]int)
	for _, response := range responses {
		for _, item := range itemsByResponse[response.ID] {
			questionType, ok := questionMetas[item.QuestionID]
			if !ok {
				continue
			}
			answerCounts[item.QuestionID]++
			if score, ok := scoreFromResponseItem(item, questionType); ok {
				sums[item.QuestionID] += score
				scoreCounts[item.QuestionID]++
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
			AvgScore:   scoreToFivePoint(avgScore),
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

	start, end := rollingReportRange(period, service.now)

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

	responseItems, err := service.reports.ListResponseItemsByResponseIDs(surveyResponseIDs(responses))
	if err != nil {
		return nil, err
	}
	itemsByResponse := groupResponseItemsByResponseID(responseItems)

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
		answersPayload, err := buildAnswerPayload(itemsByResponse[response.ID])
		if err != nil {
			return nil, err
		}
		responseDTOs = append(responseDTOs, domain.SurveySatisfactionExportResponseDTO{
			ID:        response.ID,
			TicketID:  strconv.Itoa(response.TicketID),
			UserID:    response.UserID,
			Score:     scoreToFivePoint(response.Score),
			CreatedAt: response.CreatedAt,
			Answers:   answersPayload,
		})
	}

	categoryName := service.resolveCategoryName(categoryID)

	report := &domain.SurveySatisfactionExportDTO{
		TemplateID: template.ID,
		Template:   template.Title,
		Framework:  template.Framework,
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

	usedIDs, err := service.reports.ListUsedTemplateIDsByCategory(categoryID)
	if err != nil {
		return nil, err
	}
	if len(usedIDs) == 0 {
		return []domain.SurveyTemplateDTO{}, nil
	}

	templates, err := service.reports.ListTemplatesByIDsWithQuestions(usedIDs)
	if err != nil {
		return nil, err
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].UpdatedAt.After(templates[j].UpdatedAt)
	})

	return mapSurveyTemplates(templates), nil
}

func (service *ReportService) SurveyCategoriesWithResponses() ([]domain.ServiceCategoryDTO, error) {
	usedIDs, err := service.reports.ListUsedCategoryIDs()
	if err != nil {
		return nil, err
	}
	if len(usedIDs) == 0 {
		return []domain.ServiceCategoryDTO{}, nil
	}

	usedSet := make(map[string]struct{}, len(usedIDs))
	for _, id := range usedIDs {
		if strings.TrimSpace(id) != "" {
			usedSet[id] = struct{}{}
		}
	}

	categories, err := service.categories.List()
	if err != nil {
		return nil, err
	}

	result := make([]domain.ServiceCategoryDTO, 0)
	for _, item := range categories {
		itemID := strconv.Itoa(item.ID)
		if _, ok := usedSet[itemID]; !ok {
			continue
		}
		result = append(result, domain.ServiceCategoryDTO{
			ID:           itemID,
			Name:         item.Name,
			GuestAllowed: item.GuestAllowed,
			TemplateID:   item.SurveyTemplateID,
		})
	}
	return result, nil
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
				CategoryID: strconv.Itoa(cat.ID),
				Category:   cat.Name,
				Tickets:    ticketCounts[entity][strconv.Itoa(cat.ID)],
				Surveys:    surveyCounts[entity][strconv.Itoa(cat.ID)],
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

func rollingReportRange(period string, nowFn func() time.Time) (time.Time, time.Time) {
	now := nowInWIB(nowFn)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, reportLocationWIB)
	end := dayStart.AddDate(0, 0, 1)

	switch normalizePeriod(period) {
	case "daily":
		return end.AddDate(0, 0, -1), end
	case "weekly":
		return end.AddDate(0, 0, -7), end
	case "yearly":
		return end.AddDate(0, 0, -365), end
	default:
		return end.AddDate(0, 0, -30), end
	}
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
		parsedID, err := strconv.Atoi(strings.TrimSpace(categoryID))
		if err != nil || parsedID <= 0 {
			return nil, gorm.ErrRecordNotFound
		}
		category, err := service.categories.FindByID(parsedID)
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
			categories[strconv.Itoa(cat.ID)] = cat.Name
		}
	}
	return categories
}

func (service *ReportService) resolveCategoryName(categoryID string) string {
	parsedID, err := strconv.Atoi(strings.TrimSpace(categoryID))
	if err != nil || parsedID <= 0 {
		return categoryID
	}
	category, err := service.categories.FindByID(parsedID)
	if err == nil && category.Name != "" {
		return category.Name
	}
	return categoryID
}

func surveyResponseIDs(responses []domain.SurveyResponse) []string {
	ids := make([]string, 0, len(responses))
	for _, response := range responses {
		ids = append(ids, response.ID)
	}
	return ids
}

func groupResponseItemsByResponseID(
	items []domain.SurveyResponseItem,
) map[string][]domain.SurveyResponseItem {
	grouped := make(map[string][]domain.SurveyResponseItem, len(items))
	for _, item := range items {
		grouped[item.ResponseID] = append(grouped[item.ResponseID], item)
	}
	return grouped
}

func scoreFromResponseItem(
	item domain.SurveyResponseItem,
	questionType domain.SurveyQuestionType,
) (float64, bool) {
	if item.ScoreValue != nil {
		return *item.ScoreValue, true
	}
	if len(item.AnswerValue) == 0 {
		return 0, false
	}
	var value interface{}
	if err := json.Unmarshal(item.AnswerValue, &value); err != nil {
		return 0, false
	}
	return scoreFromQuestionValue(value, questionType)
}

func buildAnswerPayload(items []domain.SurveyResponseItem) ([]byte, error) {
	answers := make(map[string]interface{}, len(items))
	for _, item := range items {
		if len(item.AnswerValue) == 0 {
			continue
		}
		var value interface{}
		if err := json.Unmarshal(item.AnswerValue, &value); err != nil {
			return nil, err
		}
		answers[item.QuestionID] = value
	}
	return json.Marshal(answers)
}
