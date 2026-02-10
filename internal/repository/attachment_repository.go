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
	query := repo.db
	if attachment.TicketID == "" {
		query = query.Omit("TicketID")
	}
	return query.Create(attachment).Error
}

func (repo *AttachmentRepository) FindByID(id string) (*domain.Attachment, error) {
	var attachment domain.Attachment
	if err := repo.db.First(&attachment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (repo *AttachmentRepository) ListByTicketID(ticketID string) ([]domain.Attachment, error) {
	if ticketID == "" {
		return []domain.Attachment{}, nil
	}
	var attachments []domain.Attachment
	if err := repo.db.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&attachments).Error; err != nil {
		return nil, err
	}
	return attachments, nil
}

func (repo *AttachmentRepository) ListByTicketIDs(ticketIDs []string) (map[string][]domain.Attachment, error) {
	result := make(map[string][]domain.Attachment)
	if len(ticketIDs) == 0 {
		return result, nil
	}
	var attachments []domain.Attachment
	if err := repo.db.Where("ticket_id IN ?", ticketIDs).Order("created_at asc").Find(&attachments).Error; err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		result[attachment.TicketID] = append(result[attachment.TicketID], attachment)
	}
	return result, nil
}

func (repo *AttachmentRepository) AttachToTicket(ids []string, ticketID string) error {
	if len(ids) == 0 || ticketID == "" {
		return nil
	}
	return repo.db.Model(&domain.Attachment{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"ticket_id": ticketID}).Error
}
