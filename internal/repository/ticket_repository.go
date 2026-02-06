package repository

import (
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

// TicketRepositoryInterface defines the contract for ticket data access
type TicketRepositoryInterface interface {
	Create(ticket *domain.Ticket) error
	Update(ticket *domain.Ticket) error
	SoftDelete(ticketID string) error
	FindByID(ticketID string) (*domain.Ticket, error)
	ListByUser(userID string) ([]domain.Ticket, error)
	ListAll() ([]domain.Ticket, error)
	Search(query string, isGuest bool) ([]domain.Ticket, error)
	ListFiltered(filter TicketListFilter, page int, limit int) ([]domain.Ticket, int64, error)
	CountForYear(year int) (int64, error)
	AddHistory(history *domain.TicketHistory) error
	AddComment(comment *domain.TicketComment) error
	UpdateStatus(ticketID string, status domain.TicketStatus, surveyRequired bool) error
	GetSurveyScores(ticketIDs []string) (map[string]float64, error)
}

type TicketRepository struct {
	db *gorm.DB
}

type TicketListFilter struct {
	Query      string
	Status     *domain.TicketStatus
	CategoryID string
	Start      *time.Time
	End        *time.Time
	ReporterID string
	IsGuest    *bool
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

func (repo *TicketRepository) ListFiltered(
	filter TicketListFilter,
	page int,
	limit int,
) ([]domain.Ticket, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	qb := repo.db.Model(&domain.Ticket{})
	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		qb = qb.Where("id ILIKE ? OR title ILIKE ?", like, like)
	}
	if filter.Status != nil {
		qb = qb.Where("status = ?", *filter.Status)
	}
	if filter.CategoryID != "" {
		qb = qb.Where("category_id = ?", filter.CategoryID)
	}
	if filter.ReporterID != "" {
		qb = qb.Where("reporter_id = ?", filter.ReporterID)
	}
	if filter.IsGuest != nil {
		qb = qb.Where("is_guest = ?", *filter.IsGuest)
	}
	if filter.Start != nil {
		qb = qb.Where("created_at >= ?", *filter.Start)
	}
	if filter.End != nil {
		qb = qb.Where("created_at < ?", *filter.End)
	}

	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tickets []domain.Ticket
	if err := qb.Preload("Category").
		Order("created_at desc").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&tickets).Error; err != nil {
		return nil, 0, err
	}
	return tickets, total, nil
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

func (repo *TicketRepository) GetSurveyScores(ticketIDs []string) (map[string]float64, error) {
	scores := make(map[string]float64)
	if len(ticketIDs) == 0 {
		return scores, nil
	}
	type row struct {
		TicketID string
		AvgScore float64
	}
	var rows []row
	if err := repo.db.Model(&domain.SurveyResponse{}).
		Select("ticket_id, AVG(score) as avg_score").
		Where("ticket_id IN ?", ticketIDs).
		Group("ticket_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		scores[item.TicketID] = item.AvgScore
	}
	return scores, nil
}
