package repository

import (
	"strings"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (repo *CategoryRepository) List() ([]domain.ServiceCategory, error) {
	var categories []domain.ServiceCategory
	if err := repo.db.Order("name asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (repo *CategoryRepository) FindByID(id string) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (repo *CategoryRepository) FindByName(name string) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.Where("lower(name) = ?", strings.ToLower(name)).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (repo *CategoryRepository) Upsert(category domain.ServiceCategory) error {
	var existing domain.ServiceCategory
	if err := repo.db.First(&existing, "id = ?", category.ID).Error; err == nil {
		return repo.db.Model(&existing).Updates(map[string]any{
			"name":          category.Name,
			"guest_allowed": category.GuestAllowed,
		}).Error
	}

	var surveyTemplateID any = nil
	if strings.TrimSpace(category.SurveyTemplateID) != "" {
		surveyTemplateID = strings.TrimSpace(category.SurveyTemplateID)
	}
	return repo.db.Model(&domain.ServiceCategory{}).Create(map[string]any{
		"id":                 category.ID,
		"name":               category.Name,
		"guest_allowed":      category.GuestAllowed,
		"survey_template_id": surveyTemplateID,
	}).Error
}

func (repo *CategoryRepository) UpdateTemplate(categoryID string, templateID string) error {
	var value any = nil
	if strings.TrimSpace(templateID) != "" {
		value = strings.TrimSpace(templateID)
	}
	return repo.db.Model(&domain.ServiceCategory{}).
		Where("id = ?", categoryID).
		Update("survey_template_id", value).Error
}

func (repo *CategoryRepository) BindTemplateToCategory(categoryID string, templateID string) error {
	cleanCategoryID := strings.TrimSpace(categoryID)
	cleanTemplateID := strings.TrimSpace(templateID)
	if cleanCategoryID == "" || cleanTemplateID == "" {
		return gorm.ErrInvalidData
	}

	result := repo.db.Model(&domain.ServiceCategory{}).
		Where("id = ?", cleanCategoryID).
		Update("survey_template_id", cleanTemplateID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
