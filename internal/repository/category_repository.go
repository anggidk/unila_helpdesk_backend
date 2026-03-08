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

func serviceIsGuestAllowed(serviceID int) bool {
	switch serviceID {
	case domain.ServiceGuestPassword, domain.ServiceGuestRegistration, domain.ServiceGuestEmail:
		return true
	default:
		return false
	}
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (repo *CategoryRepository) List() ([]domain.ServiceCategory, error) {
	var categories []domain.ServiceCategory
	if err := repo.db.
		Table("services s").
		Select("s.id, s.name, ct.template_id AS survey_template_id").
		Joins("LEFT JOIN category_templates ct ON ct.category_id = s.id").
		Order("s.id asc").
		Scan(&categories).Error; err != nil {
		return nil, err
	}
	for index := range categories {
		categories[index].GuestAllowed = serviceIsGuestAllowed(categories[index].ID)
	}
	return categories, nil
}

func (repo *CategoryRepository) FindByID(id int) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.
		Table("services s").
		Select("s.id, s.name, ct.template_id AS survey_template_id").
		Joins("LEFT JOIN category_templates ct ON ct.category_id = s.id").
		Where("s.id = ?", id).
		First(&category).Error; err != nil {
		return nil, err
	}
	category.GuestAllowed = serviceIsGuestAllowed(category.ID)
	return &category, nil
}

func (repo *CategoryRepository) FindByName(name string) (*domain.ServiceCategory, error) {
	var category domain.ServiceCategory
	if err := repo.db.Table("services").Where("lower(name) = ?", strings.ToLower(name)).First(&category).Error; err != nil {
		return nil, err
	}
	category.GuestAllowed = serviceIsGuestAllowed(category.ID)
	return &category, nil
}

func (repo *CategoryRepository) Upsert(category domain.ServiceCategory) error {
	var existing domain.ServiceCategory
	if err := repo.db.Table("services").First(&existing, "id = ?", category.ID).Error; err == nil {
		return repo.db.Model(&existing).Updates(map[string]any{
			"name": category.Name,
		}).Error
	}
	return repo.db.Table("services").Create(map[string]any{
		"id":   category.ID,
		"name": category.Name,
	}).Error
}

func (repo *CategoryRepository) UpdateTemplate(categoryID int, templateID string) error {
	cleanTemplateID := strings.TrimSpace(templateID)
	if categoryID <= 0 {
		return gorm.ErrInvalidData
	}
	if cleanTemplateID == "" {
		return repo.db.Where("category_id = ?", categoryID).
			Delete(&domain.CategoryTemplate{}).Error
	}
	record := domain.CategoryTemplate{
		CategoryID: categoryID,
		TemplateID: cleanTemplateID,
		AssignedAt: time.Now(),
	}
	return repo.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "category_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"template_id", "assigned_at"}),
	}).Create(&record).Error
}

func (repo *CategoryRepository) BindTemplateToCategory(categoryID int, templateID string) error {
	return repo.UpdateTemplate(categoryID, templateID)
}
