package repository

import (
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (repo *CategoryRepository) List() ([]domain.ServiceCategory, error) {
	var categories []domain.ServiceCategory
	if err := repo.db.
		Table("service_categories sc").
		Select("sc.id, sc.name, sc.guest_allowed, ct.template_id AS survey_template_id").
		Joins("LEFT JOIN category_templates ct ON ct.category_id = sc.id").
		Order("sc.name asc").
		Scan(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (repo *CategoryRepository) FindByID(id string) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.
		Table("service_categories sc").
		Select("sc.id, sc.name, sc.guest_allowed, ct.template_id AS survey_template_id").
		Joins("LEFT JOIN category_templates ct ON ct.category_id = sc.id").
		Where("sc.id = ?", id).
		First(&category).Error; err != nil {
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
	return repo.db.Model(&domain.ServiceCategory{}).Create(map[string]any{
		"id":            category.ID,
		"name":          category.Name,
		"guest_allowed": category.GuestAllowed,
	}).Error
}

func (repo *CategoryRepository) UpdateTemplate(categoryID string, templateID string) error {
	cleanCategoryID := strings.TrimSpace(categoryID)
	cleanTemplateID := strings.TrimSpace(templateID)
	if cleanCategoryID == "" {
		return gorm.ErrInvalidData
	}
	if cleanTemplateID == "" {
		return repo.db.Where("category_id = ?", cleanCategoryID).
			Delete(&domain.CategoryTemplate{}).Error
	}
	record := domain.CategoryTemplate{
		CategoryID: cleanCategoryID,
		TemplateID: cleanTemplateID,
		AssignedAt: time.Now(),
	}
	return repo.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "category_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"template_id", "assigned_at"}),
	}).Create(&record).Error
}

func (repo *CategoryRepository) BindTemplateToCategory(categoryID string, templateID string) error {
	return repo.UpdateTemplate(categoryID, templateID)
}
