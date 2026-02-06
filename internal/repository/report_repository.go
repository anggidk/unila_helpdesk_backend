package repository

import (
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

type CategoryCountRow struct {
	CategoryID string
	Total      int64
}

type ServiceSatisfactionRow struct {
	CategoryID string
	AvgScore   float64
	Responses  int
}

type EntityCategoryTotalRow struct {
	Entity     string
	CategoryID string
	Total      int
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (repo *ReportRepository) ListSurveyResponsesByCreatedRange(start time.Time, end time.Time) ([]domain.SurveyResponse, error) {
	var responses []domain.SurveyResponse
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Find(&responses).Error; err != nil {
		return nil, err
	}
	return responses, nil
}

func (repo *ReportRepository) ListActiveUsersInRange(userIDs []string, start time.Time, end time.Time) ([]string, error) {
	activeUsers := make([]string, 0)
	if len(userIDs) == 0 {
		return activeUsers, nil
	}
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Where("user_id IN ?", userIDs).
		Where("created_at >= ? AND created_at < ?", start, end).
		Distinct().
		Pluck("user_id", &activeUsers).Error; err != nil {
		return nil, err
	}
	return activeUsers, nil
}

func (repo *ReportRepository) ListTicketTotalsByCategory(start time.Time, end time.Time) ([]CategoryCountRow, error) {
	var rows []CategoryCountRow
	if err := repo.db.Model(&domain.Ticket{}).
		Select("category_id as category_id, count(*) as total").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("category_id").
		Order("total desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (repo *ReportRepository) ListCategories() ([]domain.ServiceCategory, error) {
	var categories []domain.ServiceCategory
	if err := repo.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (repo *ReportRepository) CountTickets() (int64, error) {
	var total int64
	if err := repo.db.Model(&domain.Ticket{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *ReportRepository) CountOpenTickets(resolvedStatus domain.TicketStatus) (int64, error) {
	var total int64
	if err := repo.db.Model(&domain.Ticket{}).
		Where("status != ?", resolvedStatus).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *ReportRepository) CountResolvedTicketsInRange(start time.Time, end time.Time, resolvedStatus domain.TicketStatus) (int64, error) {
	var total int64
	if err := repo.db.Model(&domain.Ticket{}).
		Where("status = ?", resolvedStatus).
		Where("updated_at >= ? AND updated_at < ?", start, end).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *ReportRepository) AveragePositiveSurveyScore() (float64, error) {
	var avgScore float64
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Where("score > 0").
		Select("COALESCE(AVG(score), 0)").
		Scan(&avgScore).Error; err != nil {
		return 0, err
	}
	return avgScore, nil
}

func (repo *ReportRepository) ListServiceSatisfactionRows(start time.Time, end time.Time) ([]ServiceSatisfactionRow, error) {
	var rows []ServiceSatisfactionRow
	if err := repo.db.Raw(`
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
	return rows, nil
}

func (repo *ReportRepository) FindCategoryByID(categoryID string) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.First(&category, "id = ?", categoryID).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (repo *ReportRepository) FindTemplateWithQuestions(templateID string) (*domain.SurveyTemplate, error) {
	var template domain.SurveyTemplate
	if err := repo.db.Preload("Questions").First(&template, "id = ?", templateID).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (repo *ReportRepository) FindTemplateWithOrderedQuestions(templateID string) (*domain.SurveyTemplate, error) {
	var template domain.SurveyTemplate
	if err := repo.db.Preload("Questions", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("created_at asc")
	}).First(&template, "id = ?", templateID).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (repo *ReportRepository) ListSurveyResponsesByTicketCategoryAndTemplate(
	start time.Time,
	end time.Time,
	categoryID string,
	templateID string,
	orderAsc bool,
) ([]domain.SurveyResponse, error) {
	var responses []domain.SurveyResponse
	query := repo.db.Model(&domain.SurveyResponse{}).
		Joins("JOIN tickets t ON t.id = survey_responses.ticket_id").
		Where("survey_responses.created_at >= ? AND survey_responses.created_at < ?", start, end)
	if categoryID != "" {
		query = query.Where("t.category_id = ?", categoryID)
	}
	if templateID != "" {
		query = query.Where("survey_responses.template_id = ?", templateID)
	}
	orderBy := "survey_responses.created_at desc"
	if orderAsc {
		orderBy = "survey_responses.created_at asc"
	}
	if err := query.Order(orderBy).Find(&responses).Error; err != nil {
		return nil, err
	}
	return responses, nil
}

func (repo *ReportRepository) ListUsedTemplateIDsByCategory(categoryID string) ([]string, error) {
	usedIDs := make([]string, 0)
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Joins("JOIN tickets t ON t.id = survey_responses.ticket_id").
		Where("t.category_id = ? AND survey_responses.template_id <> ''", categoryID).
		Distinct().
		Pluck("survey_responses.template_id", &usedIDs).Error; err != nil {
		return nil, err
	}
	return usedIDs, nil
}

func (repo *ReportRepository) ListTemplatesByIDsWithQuestions(ids []string) ([]domain.SurveyTemplate, error) {
	if len(ids) == 0 {
		return []domain.SurveyTemplate{}, nil
	}
	var templates []domain.SurveyTemplate
	if err := repo.db.Preload("Questions").
		Where("id IN ?", ids).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (repo *ReportRepository) CountTicketsInRange(start time.Time, end time.Time) (int64, error) {
	var total int64
	if err := repo.db.Model(&domain.Ticket{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *ReportRepository) CountSurveysInRange(start time.Time, end time.Time) (int64, error) {
	var total int64
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (repo *ReportRepository) ListRegisteredTicketRowsByEntityCategory(start time.Time, end time.Time) ([]EntityCategoryTotalRow, error) {
	var rows []EntityCategoryTotalRow
	if err := repo.db.Raw(`
        SELECT u.entity AS entity, t.category_id AS category_id, COUNT(*) AS total
        FROM tickets t
        JOIN users u ON u.id = t.reporter_id
        WHERE u.role = 'registered'
          AND t.created_at >= ? AND t.created_at < ?
        GROUP BY u.entity, t.category_id
    `, start, end).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (repo *ReportRepository) ListRegisteredSurveyRowsByEntityCategory(start time.Time, end time.Time) ([]EntityCategoryTotalRow, error) {
	var rows []EntityCategoryTotalRow
	if err := repo.db.Raw(`
        SELECT u.entity AS entity, t.category_id AS category_id, COUNT(*) AS total
        FROM survey_responses sr
        JOIN users u ON u.id = sr.user_id
        JOIN tickets t ON t.id = sr.ticket_id
        WHERE u.role = 'registered'
          AND sr.created_at >= ? AND sr.created_at < ?
        GROUP BY u.entity, t.category_id
    `, start, end).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (repo *ReportRepository) ListRegisteredEntities() ([]string, error) {
	entities := make([]string, 0)
	if err := repo.db.Model(&domain.User{}).
		Distinct("entity").
		Where("role = ?", domain.RoleRegistered).
		Where("entity <> ''").
		Order("entity asc").
		Pluck("entity", &entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (repo *ReportRepository) ListRegisteredCategories() ([]domain.ServiceCategory, error) {
	var categories []domain.ServiceCategory
	if err := repo.db.Model(&domain.ServiceCategory{}).
		Where("guest_allowed = ?", false).
		Order("name asc").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
