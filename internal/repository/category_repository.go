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
    return repo.db.Create(&category).Error
}

func (repo *CategoryRepository) UpdateTemplate(categoryID string, templateID string) error {
    return repo.db.Model(&domain.ServiceCategory{}).
        Where("id = ?", categoryID).
        Update("survey_template_id", templateID).Error
}
