package repository

import (
    "time"

    "unila_helpdesk_backend/internal/domain"

    "gorm.io/gorm"
)

type TicketRepository struct {
    db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
    return &TicketRepository{db: db}
}

func (repo *TicketRepository) Create(ticket *domain.Ticket) error {
    return repo.db.Create(ticket).Error
}

func (repo *TicketRepository) Update(ticket *domain.Ticket) error {
    return repo.db.Save(ticket).Error
}

func (repo *TicketRepository) SoftDelete(ticketID string) error {
    return repo.db.Delete(&domain.Ticket{}, "id = ?", ticketID).Error
}

func (repo *TicketRepository) FindByID(ticketID string) (*domain.Ticket, error) {
    var ticket domain.Ticket
    if err := repo.db.Preload("Category").Preload("History", func(db *gorm.DB) *gorm.DB {
        return db.Order("timestamp desc")
    }).Preload("Comments", func(db *gorm.DB) *gorm.DB {
        return db.Order("timestamp asc")
    }).First(&ticket, "id = ?", ticketID).Error; err != nil {
        return nil, err
    }
    return &ticket, nil
}

func (repo *TicketRepository) ListByUser(userID string) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    if err := repo.db.Preload("Category").Where("reporter_id = ?", userID).Order("created_at desc").Find(&tickets).Error; err != nil {
        return nil, err
    }
    return tickets, nil
}

func (repo *TicketRepository) ListAll() ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    if err := repo.db.Preload("Category").Order("created_at desc").Find(&tickets).Error; err != nil {
        return nil, err
    }
    return tickets, nil
}

func (repo *TicketRepository) Search(query string, isGuest bool) ([]domain.Ticket, error) {
    var tickets []domain.Ticket
    qb := repo.db.Preload("Category").Order("created_at desc")
    if query != "" {
        like := "%" + query + "%"
        qb = qb.Where("id ILIKE ? OR title ILIKE ?", like, like)
    }
    if isGuest {
        qb = qb.Where("is_guest = ?", true)
    }
    if err := qb.Find(&tickets).Error; err != nil {
        return nil, err
    }
    return tickets, nil
}

func (repo *TicketRepository) CountForYear(year int) (int64, error) {
    var count int64
    start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
    if err := repo.db.Model(&domain.Ticket{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error; err != nil {
        return 0, err
    }
    return count, nil
}

func (repo *TicketRepository) AddHistory(history *domain.TicketHistory) error {
    return repo.db.Create(history).Error
}

func (repo *TicketRepository) AddComment(comment *domain.TicketComment) error {
    return repo.db.Create(comment).Error
}

func (repo *TicketRepository) UpdateStatus(ticketID string, status domain.TicketStatus, surveyRequired bool) error {
    return repo.db.Model(&domain.Ticket{}).Where("id = ?", ticketID).Updates(map[string]any{
        "status":          status,
        "survey_required": surveyRequired,
    }).Error
}
