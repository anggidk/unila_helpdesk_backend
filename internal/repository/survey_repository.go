package repository

import (
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

// SurveyRepositoryInterface defines the contract for survey data access
type SurveyRepositoryInterface interface {
	ListTemplates() ([]domain.SurveyTemplate, error)
	FindByCategory(categoryID string) (*domain.SurveyTemplate, error)
	FindByID(templateID string) (*domain.SurveyTemplate, error)
	CreateTemplate(template *domain.SurveyTemplate) error
	UpdateTemplate(template *domain.SurveyTemplate) error
	ReplaceTemplate(template *domain.SurveyTemplate) error
	DeleteTemplate(templateID string) error
	SaveResponse(response *domain.SurveyResponse) error
	HasResponse(ticketID string, userID string) (bool, error)
	ListResponses(filter SurveyResponseFilter, page int, limit int) ([]SurveyResponseRow, int64, error)
}

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
	TicketID      string
	UserID        string
	TemplateID    string
	Score         float64
	CreatedAt     time.Time
	UserName      string
	UserEmail     string
	UserEntity    string
	CategoryID    string
	CategoryName  string
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
	var category domain.ServiceCategory
	if err := repo.db.First(&category, "id = ?", categoryID).Error; err != nil {
		return nil, err
	}
	if category.SurveyTemplateID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var template domain.SurveyTemplate
	if err := repo.db.Preload("Questions").
		First(&template, "id = ?", category.SurveyTemplateID).Error; err != nil {
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

func (repo *SurveyRepository) UpdateTemplate(template *domain.SurveyTemplate) error {
	return repo.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(template).Error
}

func (repo *SurveyRepository) ReplaceTemplate(template *domain.SurveyTemplate) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"title":       template.Title,
			"description": template.Description,
			"category_id": template.CategoryID,
			"updated_at":  template.UpdatedAt,
		}
		result := tx.Model(&domain.SurveyTemplate{}).Where("id = ?", template.ID).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("template_id = ?", template.ID).Delete(&domain.SurveyQuestion{}).Error; err != nil {
			return err
		}
		if len(template.Questions) > 0 {
			if err := tx.Create(&template.Questions).Error; err != nil {
				return err
			}
		}
		return nil
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

func (repo *SurveyRepository) SaveResponse(response *domain.SurveyResponse) error {
	return repo.db.Create(response).Error
}

func (repo *SurveyRepository) HasResponse(ticketID string, userID string) (bool, error) {
	var count int64
	if err := repo.db.Model(&domain.SurveyResponse{}).Where("ticket_id = ? AND user_id = ?", ticketID, userID).Count(&count).Error; err != nil {
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
		Joins("JOIN users u ON u.id = sr.user_id").
		Joins("JOIN tickets t ON t.id = sr.ticket_id").
		Joins("LEFT JOIN service_categories sc ON sc.id = t.category_id").
		Joins("LEFT JOIN survey_templates st ON st.id = sr.template_id")

	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		base = base.Where("sr.ticket_id ILIKE ? OR u.name ILIKE ? OR u.email ILIKE ?", like, like, like)
	}
	if filter.CategoryID != "" {
		base = base.Where("t.category_id = ?", filter.CategoryID)
	}
	if filter.TemplateID != "" {
		base = base.Where("sr.template_id = ?", filter.TemplateID)
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
            sr.user_id,
            sr.template_id,
            sr.score,
            sr.created_at,
            u.name as user_name,
            u.email as user_email,
            u.entity as user_entity,
            t.category_id as category_id,
            sc.name as category_name,
            st.title as template_title
        `).
		Order("sr.created_at desc").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
