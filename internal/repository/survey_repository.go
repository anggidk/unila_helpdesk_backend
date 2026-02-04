package repository

import (
    "unila_helpdesk_backend/internal/domain"

    "gorm.io/gorm"
)

type SurveyRepository struct {
    db *gorm.DB
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
