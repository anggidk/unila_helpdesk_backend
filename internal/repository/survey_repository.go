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
    var template domain.SurveyTemplate
    if err := repo.db.Preload("Questions").First(&template, "category_id = ?", categoryID).Error; err != nil {
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
