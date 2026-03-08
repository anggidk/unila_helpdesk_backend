package repository

import (
	"strconv"
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
	return repo.db.Create(ticket).Error
}

func (repo *TicketRepository) Update(ticket *domain.Ticket) error {
	updates := map[string]any{
		"ticket_number":   ticket.TicketNumber,
		"ticket_date":     ticket.CreatedAt,
		"username":        nullableStringValue(ticket.Username),
		"number_id":       nullableStringValue(ticket.NumberID),
		"name":            strings.TrimSpace(ticket.Name),
		"email":           strings.TrimSpace(ticket.Email),
		"entity":          strings.TrimSpace(ticket.Entity),
		"id_service":      ticket.ServiceID,
		"notes":           strings.TrimSpace(ticket.Notes),
		"staff_notes":     strings.TrimSpace(ticket.StaffNotes),
		"priority":        ticket.Priority,
		"is_reject":       ticket.IsReject,
		"is_assign":       ticket.IsAssign,
		"is_done":         ticket.IsDone,
		"id_staff":        nullableTrimmed(ticket.StaffID),
		"status":          ticket.Status,
		"lamp1":           strings.TrimSpace(ticket.Lamp1),
		"lamp2":           strings.TrimSpace(ticket.Lamp2),
		"survey_required": ticket.SurveyRequired,
	}
	return repo.db.Model(&domain.Ticket{}).Where("id = ?", ticket.ID).Updates(updates).Error
}

func nullableTrimmed(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableStringValue(value *string) any {
	return nullableTrimmed(value)
}

func (repo *TicketRepository) SoftDelete(ticketID int) error {
	return repo.db.Delete(&domain.Ticket{}, "id = ?", ticketID).Error
}

func (repo *TicketRepository) FindByID(ticketID int) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := repo.db.Preload("Service").First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}
	ticket.Service.GuestAllowed = serviceIsGuestAllowed(ticket.Service.ID)
	return &ticket, nil
}

func (repo *TicketRepository) ListByUser(user domain.User) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	if err := repo.db.Preload("Service").
		Where("number_id = ? AND username = ?", user.ID, user.Username).
		Order("ticket_date desc").
		Find(&tickets).Error; err != nil {
		return nil, err
	}
	for index := range tickets {
		tickets[index].Service.GuestAllowed = serviceIsGuestAllowed(tickets[index].Service.ID)
	}
	return tickets, nil
}

func (repo *TicketRepository) ListAll() ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	if err := repo.db.Preload("Service").Order("ticket_date desc").Find(&tickets).Error; err != nil {
		return nil, err
	}
	for index := range tickets {
		tickets[index].Service.GuestAllowed = serviceIsGuestAllowed(tickets[index].Service.ID)
	}
	return tickets, nil
}

func (repo *TicketRepository) Search(query string, guestOnly bool) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	qb := repo.db.Preload("Service").Order("ticket_date desc")
	if guestOnly {
		qb = qb.Where("username IS NULL")
	}
	if strings.TrimSpace(query) != "" {
		like := "%" + strings.TrimSpace(query) + "%"
		qb = qb.Where("CAST(id AS text) ILIKE ? OR ticket_number ILIKE ? OR notes ILIKE ? OR name ILIKE ? OR email ILIKE ?", like, like, like, like, like)
	}
	if err := qb.Find(&tickets).Error; err != nil {
		return nil, err
	}
	for index := range tickets {
		tickets[index].Service.GuestAllowed = serviceIsGuestAllowed(tickets[index].Service.ID)
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
	if strings.TrimSpace(filter.Query) != "" {
		like := "%" + strings.TrimSpace(filter.Query) + "%"
		qb = qb.Where("CAST(id AS text) ILIKE ? OR ticket_number ILIKE ? OR notes ILIKE ? OR name ILIKE ? OR email ILIKE ?", like, like, like, like, like)
	}
	if filter.Status != nil {
		qb = qb.Where("status = ?", *filter.Status)
	}
	if strings.TrimSpace(filter.CategoryID) != "" {
		if serviceID, err := strconv.Atoi(strings.TrimSpace(filter.CategoryID)); err == nil && serviceID > 0 {
			qb = qb.Where("id_service = ?", serviceID)
		}
	}
	if strings.TrimSpace(filter.UserID) != "" {
		qb = qb.Where("number_id = ?", strings.TrimSpace(filter.UserID))
	}
	if filter.IsGuest != nil {
		if *filter.IsGuest {
			qb = qb.Where("username IS NULL")
		} else {
			qb = qb.Where("username IS NOT NULL")
		}
	}
	if filter.Start != nil {
		qb = qb.Where("ticket_date >= ?", *filter.Start)
	}
	if filter.End != nil {
		qb = qb.Where("ticket_date < ?", *filter.End)
	}

	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tickets []domain.Ticket
	if err := qb.Preload("Service").
		Order("ticket_date desc").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&tickets).Error; err != nil {
		return nil, 0, err
	}
	for index := range tickets {
		tickets[index].Service.GuestAllowed = serviceIsGuestAllowed(tickets[index].Service.ID)
	}
	return tickets, total, nil
}

func (repo *TicketRepository) ExistsTicketNumber(ticketNumber string) (bool, error) {
	var count int64
	if err := repo.db.Model(&domain.Ticket{}).Where("ticket_number = ?", strings.TrimSpace(ticketNumber)).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *TicketRepository) UpdateStatus(ticketID int, status domain.TicketStatus, surveyRequired bool) error {
	isAssign, isDone, isReject := ticketStatusFlags(status)
	return repo.db.Model(&domain.Ticket{}).Where("id = ?", ticketID).Updates(map[string]any{
		"status":          status,
		"is_assign":       isAssign,
		"is_done":         isDone,
		"is_reject":       isReject,
		"survey_required": surveyRequired,
	}).Error
}

func (repo *TicketRepository) GetSurveyScores(ticketIDs []int) (map[int]float64, error) {
	scores := make(map[int]float64)
	if len(ticketIDs) == 0 {
		return scores, nil
	}
	type row struct {
		TicketID int
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

func ticketStatusFlags(status domain.TicketStatus) (bool, bool, bool) {
	switch status {
	case domain.StatusAssign:
		return true, false, false
	case domain.StatusDone:
		return true, true, false
	case domain.StatusReject:
		return false, false, true
	default:
		return false, false, false
	}
}
