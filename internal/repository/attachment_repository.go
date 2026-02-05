package repository

import (
    "unila_helpdesk_backend/internal/domain"

    "gorm.io/gorm"
)

type AttachmentRepository struct {
    db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
    return &AttachmentRepository{db: db}
}

func (repo *AttachmentRepository) Create(attachment *domain.Attachment) error {
    return repo.db.Create(attachment).Error
}

func (repo *AttachmentRepository) FindByID(id string) (*domain.Attachment, error) {
    var attachment domain.Attachment
    if err := repo.db.First(&attachment, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &attachment, nil
}

func (repo *AttachmentRepository) AttachToTicket(ids []string, ticketID string) error {
    if len(ids) == 0 || ticketID == "" {
        return nil
    }
    return repo.db.Model(&domain.Attachment{}).
        Where("id IN ?", ids).
        Updates(map[string]any{"ticket_id": ticketID}).Error
}
