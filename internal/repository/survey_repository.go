package repository

import (
	"strconv"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SurveyRepository struct {
	db *gorm.DB
}

type SurveyResponseFilter struct {
	Query      string
	CategoryID string
	TemplateID string
	Start      *time.Time
	End        *time.Time
}

type SurveyResponseRow struct {
	ID            string
	TicketID      int
	UserID        string
	TemplateID    string
	Score         float64
	CreatedAt     time.Time
	UserName      string
	UserEmail     string
	UserEntity    string
	ServiceID     int
	ServiceName   string
	TemplateTitle string
}

func NewSurveyRepository(db *gorm.DB) *SurveyRepository {
	return &SurveyRepository{db: db}
}

func (repo *SurveyRepository) ListTemplates() ([]domain.SurveyTemplate, error) {
	var templates []domain.SurveyTemplate
	if err := repo.db.Preload("Questions").Order("created_at desc").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (repo *SurveyRepository) FindByCategory(categoryID string) (*domain.SurveyTemplate, error) {
	parsedID, err := strconv.Atoi(strings.TrimSpace(categoryID))
	if err != nil || parsedID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var mapping domain.CategoryTemplate
	if err := repo.db.First(&mapping, "category_id = ?", parsedID).Error; err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var template domain.SurveyTemplate
	if err := repo.db.Preload("Questions").
		First(&template, "id = ?", mapping.TemplateID).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (repo *SurveyRepository) FindByID(templateID string) (*domain.SurveyTemplate, error) {
	var template domain.SurveyTemplate
	if err := repo.db.Preload("Questions").First(&template, "id = ?", templateID).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (repo *SurveyRepository) CreateTemplate(template *domain.SurveyTemplate) error {
	return repo.db.Create(template).Error
}

func (repo *SurveyRepository) ReplaceTemplate(template *domain.SurveyTemplate) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"title":       template.Title,
			"description": template.Description,
			"framework":   template.Framework,
			"updated_at":  template.UpdatedAt,
		}
		result := tx.Model(&domain.SurveyTemplate{}).Where("id = ?", template.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		incomingQuestionIDs := make([]string, 0, len(template.Questions))
		for _, question := range template.Questions {
			incomingQuestionIDs = append(incomingQuestionIDs, question.ID)
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"template_id", "text", "type", "options"}),
			}).Create(&question).Error; err != nil {
				return err
			}
		}

		deleteQuery := tx.Where("template_id = ?", template.ID).
			Where("id NOT IN (SELECT DISTINCT question_id FROM survey_response_items)")
		if len(incomingQuestionIDs) > 0 {
			deleteQuery = deleteQuery.Where("id NOT IN ?", incomingQuestionIDs)
		}
		return deleteQuery.Delete(&domain.SurveyQuestion{}).Error
	})
}

func (repo *SurveyRepository) DeleteTemplate(templateID string) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", templateID).Delete(&domain.SurveyQuestion{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&domain.SurveyTemplate{}, "id = ?", templateID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (repo *SurveyRepository) SaveResponse(
	response *domain.SurveyResponse,
	items []domain.SurveyResponseItem,
) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(response).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *SurveyRepository) HasResponse(ticketID int, userID string) (bool, error) {
	var count int64
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Where("ticket_id = ? AND number_id = ?", ticketID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *SurveyRepository) ListResponses(
	filter SurveyResponseFilter,
	page int,
	limit int,
) ([]SurveyResponseRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	base := repo.db.Table("survey_responses sr").
		Joins("JOIN users u ON u.id = sr.number_id").
		Joins("JOIN tickets t ON t.id = sr.ticket_id").
		Joins("LEFT JOIN services s ON s.id = t.id_service").
		Joins("LEFT JOIN survey_templates st ON st.id = sr.template_id")

	if strings.TrimSpace(filter.Query) != "" {
		like := "%" + strings.TrimSpace(filter.Query) + "%"
		base = base.Where("CAST(sr.ticket_id AS text) ILIKE ? OR u.name ILIKE ? OR u.email ILIKE ?", like, like, like)
	}
	if strings.TrimSpace(filter.CategoryID) != "" {
		if serviceID, err := strconv.Atoi(strings.TrimSpace(filter.CategoryID)); err == nil && serviceID > 0 {
			base = base.Where("t.id_service = ?", serviceID)
		}
	}
	if strings.TrimSpace(filter.TemplateID) != "" {
		base = base.Where("sr.template_id = ?", strings.TrimSpace(filter.TemplateID))
	}
	if filter.Start != nil {
		base = base.Where("sr.created_at >= ?", *filter.Start)
	}
	if filter.End != nil {
		base = base.Where("sr.created_at < ?", *filter.End)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []SurveyResponseRow
	if err := base.Select(`
			sr.id,
			sr.ticket_id,
			sr.number_id AS user_id,
			sr.template_id,
			sr.score,
			sr.created_at,
			u.name AS user_name,
			u.email AS user_email,
			u.entity AS user_entity,
			t.id_service AS service_id,
			s.name AS service_name,
			st.title AS template_title
		`).
		Order("sr.created_at desc").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
