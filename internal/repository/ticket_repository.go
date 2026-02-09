package repository

import (
	"fmt"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
}

type TicketListFilter struct {
	Query      string
	Status     *domain.TicketStatus
	CategoryID string
	Start      *time.Time
	End        *time.Time
	UserID     string
	IsGuest    *bool
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (repo *TicketRepository) Create(ticket *domain.Ticket) error {
	values := map[string]any{
		"id":              ticket.ID,
		"ticket_number":   ticket.TicketNumber,
		"user_id":         nullableTrimmed(ticket.UserID),
		"reporter_name":   strings.TrimSpace(ticket.ReporterName),
		"email":           strings.TrimSpace(ticket.Email),
		"phone":           ticket.Phone,
		"is_guest":        ticket.IsGuest,
		"title":           strings.TrimSpace(ticket.Title),
		"description":     strings.TrimSpace(ticket.Description),
		"category_id":     strings.TrimSpace(ticket.CategoryID),
		"priority":        ticket.Priority,
		"status":          ticket.Status,
		"assignee_id":     ticket.AssigneeID,
		"staff_notes":     strings.TrimSpace(ticket.StaffNotes),
		"survey_required": ticket.SurveyRequired,
		"created_at":      ticket.CreatedAt,
		"updated_at":      ticket.UpdatedAt,
	}
	return repo.db.Model(&domain.Ticket{}).Create(values).Error
}

func (repo *TicketRepository) Update(ticket *domain.Ticket) error {
	updates := map[string]any{
		"ticket_number":   ticket.TicketNumber,
		"user_id":         nullableTrimmed(ticket.UserID),
		"reporter_name":   strings.TrimSpace(ticket.ReporterName),
		"email":           strings.TrimSpace(ticket.Email),
		"phone":           ticket.Phone,
		"is_guest":        ticket.IsGuest,
		"title":           strings.TrimSpace(ticket.Title),
		"description":     strings.TrimSpace(ticket.Description),
		"category_id":     strings.TrimSpace(ticket.CategoryID),
		"priority":        ticket.Priority,
		"status":          ticket.Status,
		"assignee_id":     ticket.AssigneeID,
		"staff_notes":     strings.TrimSpace(ticket.StaffNotes),
		"survey_required": ticket.SurveyRequired,
		"updated_at":      ticket.UpdatedAt,
	}
	return repo.db.Model(&domain.Ticket{}).Where("id = ?", ticket.ID).Updates(updates).Error
}

func (repo *TicketRepository) SoftDelete(ticketID string) error {
	return repo.db.Delete(&domain.Ticket{}, "id = ?", ticketID).Error
}

func (repo *TicketRepository) FindByID(ticketID string) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := repo.db.Preload("Category").Preload("History", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc")
	}).First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (repo *TicketRepository) ListByUser(userID string) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	if err := repo.db.Preload("Category").Where("user_id = ?", userID).Order("created_at desc").Find(&tickets).Error; err != nil {
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
		qb = qb.Where("id ILIKE ? OR ticket_number ILIKE ? OR title ILIKE ?", like, like, like)
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
		qb = qb.Where("id ILIKE ? OR ticket_number ILIKE ? OR title ILIKE ? OR reporter_name ILIKE ? OR email ILIKE ?", like, like, like, like, like)
	}
	if filter.Status != nil {
		qb = qb.Where("status = ?", *filter.Status)
	}
	if filter.CategoryID != "" {
		qb = qb.Where("category_id = ?", filter.CategoryID)
	}
	if filter.UserID != "" {
		qb = qb.Where("user_id = ?", filter.UserID)
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

func (repo *TicketRepository) NextTicketSequence(year int) (int64, error) {
	seqName := fmt.Sprintf("ticket_seq_%d", year)
	createSQL := fmt.Sprintf(
		"CREATE SEQUENCE IF NOT EXISTS %s INCREMENT BY 1 MINVALUE 1 START WITH 1",
		seqName,
	)
	if err := repo.db.Exec(createSQL).Error; err != nil {
		return 0, err
	}

	likePattern := fmt.Sprintf("TK-%d-%%", year)
	var maxExisting int64
	if err := repo.db.Unscoped().
		Model(&domain.Ticket{}).
		Where("ticket_number LIKE ?", likePattern).
		Select("COALESCE(MAX(CASE WHEN split_part(ticket_number, '-', 3) ~ '^[0-9]+$' THEN split_part(ticket_number, '-', 3)::bigint ELSE 0 END), 0)").
		Scan(&maxExisting).Error; err != nil {
		return 0, err
	}

	// Keep sequence aligned with existing historical IDs without moving backward.
	if maxExisting > 0 {
		alignSQL := fmt.Sprintf(
			"SELECT setval('%s', GREATEST((SELECT last_value FROM %s), $1), true)",
			seqName,
			seqName,
		)
		if err := repo.db.Exec(alignSQL, maxExisting).Error; err != nil {
			return 0, err
		}
	}

	nextSQL := fmt.Sprintf("SELECT nextval('%s')", seqName)
	var next int64
	if err := repo.db.Raw(nextSQL).Scan(&next).Error; err != nil {
		return 0, err
	}
	return next, nil
}

func (repo *TicketRepository) AddHistory(history *domain.TicketHistory) error {
	return repo.db.Create(history).Error
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

func nullableTrimmed(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
