package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/fcm"
	"unila_helpdesk_backend/internal/repository"
	"unila_helpdesk_backend/internal/util"
)

type TicketService struct {
	tickets       *repository.TicketRepository
	categories    *repository.CategoryRepository
	notifications *repository.NotificationRepository
	tokens        *repository.FCMTokenRepository
	fcmClient     *fcm.Client
	initialStatus domain.TicketStatus
	now           func() time.Time
}

type TicketCreateRequest struct {
	ServiceID int                   `json:"serviceId"`
	Notes     string                `json:"notes"`
	Priority  domain.TicketPriority `json:"priority"`
	Lamp1     string                `json:"lamp1"`
}

type GuestTicketCreateRequest struct {
	Name      string                `json:"name"`
	NumberID  string                `json:"numberId"`
	Email     string                `json:"email"`
	Entity    string                `json:"entity"`
	ServiceID int                   `json:"serviceId"`
	Notes     string                `json:"notes"`
	Priority  domain.TicketPriority `json:"priority"`
	Lamp1     string                `json:"lamp1"`
	Lamp2     string                `json:"lamp2"`
}

type TicketUpdateRequest struct {
	ServiceID  *int                   `json:"serviceId"`
	Notes      *string                `json:"notes"`
	Priority   *domain.TicketPriority `json:"priority"`
	Status     *domain.TicketStatus   `json:"status"`
	StaffNotes *string                `json:"staffNotes"`
	Lamp1      *string                `json:"lamp1"`
	Lamp2      *string                `json:"lamp2"`
	IDStaff    *string                `json:"idStaff"`
}

func NewTicketService(
	tickets *repository.TicketRepository,
	categories *repository.CategoryRepository,
	notifications *repository.NotificationRepository,
	tokens *repository.FCMTokenRepository,
	fcmClient *fcm.Client,
	initialStatus domain.TicketStatus,
) *TicketService {
	return &TicketService{
		tickets:       tickets,
		categories:    categories,
		notifications: notifications,
		tokens:        tokens,
		fcmClient:     fcmClient,
		initialStatus: normalizeInitialTicketStatus(initialStatus),
		now:           time.Now,
	}
}

func (service *TicketService) CreateTicket(ctx context.Context, user domain.User, req TicketCreateRequest) (domain.TicketDTO, error) {
	parsedEntity, err := normalizeEntity(user.Entity)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	serviceRow, err := service.resolveCategoryID(req.ServiceID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	if serviceRow.GuestAllowed {
		return domain.TicketDTO{}, errors.New("pengguna terdaftar hanya dapat membuat tiket layanan internal")
	}
	if strings.TrimSpace(req.Notes) == "" {
		return domain.TicketDTO{}, errors.New("notes wajib diisi")
	}

	ticketNumber, err := service.generateTicketNumber()
	if err != nil {
		return domain.TicketDTO{}, err
	}

	priority := normalizePriority(req.Priority)
	ticket := domain.Ticket{
		TicketNumber: ticketNumber,
		CreatedAt:    service.now(),
		Username:     cleanOptionalString(user.Username),
		NumberID:     cleanOptionalString(user.ID),
		Name:         strings.TrimSpace(user.Name),
		Email:        strings.ToLower(strings.TrimSpace(user.Email)),
		Entity:       parsedEntity,
		ServiceID:    serviceRow.ID,
		Notes:        strings.TrimSpace(req.Notes),
		Priority:     priority,
		Lamp1:        strings.TrimSpace(req.Lamp1),
		Lamp2:        "",
	}
	service.applyStatus(&ticket, service.initialStatus)
	ticket.SurveyRequired = surveyRequiredForTicket(ticket)

	if err := service.tickets.Create(&ticket); err != nil {
		return domain.TicketDTO{}, err
	}

	if ticket.SurveyRequired {
		if err := service.notifyTicketStatus(
			ctx,
			ticket,
			"Tiket Selesai Ditangani",
			fmt.Sprintf("Tiket %s selesai ditangani. Mohon isi feedback.", ticket.TicketNumber),
		); err != nil {
			log.Printf("failed to send ticket create notification: %v", err)
		}
	}
	return service.toTicketDTO(ticket, *serviceRow, scoreZero()), nil
}

func (service *TicketService) CreateGuestTicket(ctx context.Context, req GuestTicketCreateRequest) (domain.TicketDTO, error) {
	_ = ctx
	serviceRow, err := service.resolveCategoryID(req.ServiceID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	if !serviceRow.GuestAllowed {
		return domain.TicketDTO{}, errors.New("guest hanya dapat membuat tiket layanan guest")
	}
	entity, err := normalizeEntity(req.Entity)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return domain.TicketDTO{}, errors.New("nama wajib diisi")
	}
	if strings.TrimSpace(req.NumberID) == "" {
		return domain.TicketDTO{}, errors.New("numberId wajib diisi")
	}
	if strings.TrimSpace(req.Notes) == "" {
		return domain.TicketDTO{}, errors.New("notes wajib diisi")
	}
	if strings.TrimSpace(req.Lamp1) == "" || strings.TrimSpace(req.Lamp2) == "" {
		return domain.TicketDTO{}, errors.New("lamp1 dan lamp2 wajib diisi untuk guest")
	}

	ticketNumber, err := service.generateTicketNumber()
	if err != nil {
		return domain.TicketDTO{}, err
	}

	ticket := domain.Ticket{
		TicketNumber: ticketNumber,
		CreatedAt:    service.now(),
		Username:     nil,
		NumberID:     cleanOptionalString(req.NumberID),
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Entity:       entity,
		ServiceID:    serviceRow.ID,
		Notes:        strings.TrimSpace(req.Notes),
		Priority:     normalizePriority(req.Priority),
		Lamp1:        strings.TrimSpace(req.Lamp1),
		Lamp2:        strings.TrimSpace(req.Lamp2),
	}
	service.applyStatus(&ticket, service.initialStatus)
	ticket.SurveyRequired = false

	if err := service.tickets.Create(&ticket); err != nil {
		return domain.TicketDTO{}, err
	}
	return service.toTicketDTO(ticket, *serviceRow, scoreZero()), nil
}

func (service *TicketService) UpdateTicket(ctx context.Context, user domain.User, ticketID string, req TicketUpdateRequest) (domain.TicketDTO, error) {
	parsedID, err := parseTicketID(ticketID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	ticket, err := service.tickets.FindByID(parsedID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	if !ticketOwnedByUser(*ticket, user) && user.Role != domain.RoleAdmin {
		return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk memperbarui tiket ini")
	}
	if user.Role != domain.RoleAdmin && (ticket.Status == domain.StatusDone || ticket.Status == domain.StatusReject) {
		return domain.TicketDTO{}, errors.New("tiket yang sudah selesai tidak dapat diedit")
	}

	if req.ServiceID != nil {
		serviceRow, resolveErr := service.resolveCategoryID(*req.ServiceID)
		if resolveErr != nil {
			return domain.TicketDTO{}, resolveErr
		}
		if user.Role != domain.RoleAdmin && serviceRow.GuestAllowed {
			return domain.TicketDTO{}, errors.New("pengguna terdaftar hanya dapat membuat tiket layanan internal")
		}
		ticket.ServiceID = serviceRow.ID
		ticket.Service = *serviceRow
	}
	if req.Notes != nil {
		if strings.TrimSpace(*req.Notes) == "" {
			return domain.TicketDTO{}, errors.New("notes wajib diisi")
		}
		ticket.Notes = strings.TrimSpace(*req.Notes)
	}
	if req.Priority != nil {
		ticket.Priority = normalizePriority(*req.Priority)
	}
	if req.Lamp1 != nil {
		ticket.Lamp1 = strings.TrimSpace(*req.Lamp1)
	}
	if req.Lamp2 != nil && user.Role == domain.RoleAdmin {
		ticket.Lamp2 = strings.TrimSpace(*req.Lamp2)
	}
	if req.StaffNotes != nil && user.Role == domain.RoleAdmin {
		ticket.StaffNotes = strings.TrimSpace(*req.StaffNotes)
	}
	if req.IDStaff != nil && user.Role == domain.RoleAdmin {
		cleanStaff := strings.TrimSpace(*req.IDStaff)
		if cleanStaff == "" {
			ticket.StaffID = nil
		} else {
			ticket.StaffID = &cleanStaff
		}
	}

	statusChanged := false
	previousStatus := ticket.Status
	if req.Status != nil && user.Role == domain.RoleAdmin && ticket.Status != *req.Status {
		service.applyStatus(ticket, *req.Status)
		statusChanged = true
	}
	ticket.SurveyRequired = surveyRequiredForTicket(*ticket)

	if err := service.tickets.Update(ticket); err != nil {
		return domain.TicketDTO{}, err
	}
	if statusChanged {
		title, message := statusChangeNotification(ticket.TicketNumber, previousStatus, ticket.Status, ticket.SurveyRequired)
		if err := service.notifyTicketStatus(ctx, *ticket, title, message); err != nil {
			log.Printf("failed to send status notification: %v", err)
		}
	}
	serviceRow := ticket.Service
	serviceRow.GuestAllowed = isGuestServiceID(serviceRow.ID)
	return service.toTicketDTO(*ticket, serviceRow, scoreZero()), nil
}

func (service *TicketService) DeleteTicket(user domain.User, ticketID string) error {
	parsedID, err := parseTicketID(ticketID)
	if err != nil {
		return err
	}
	ticket, err := service.tickets.FindByID(parsedID)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleAdmin && !ticketOwnedByUser(*ticket, user) {
		return errors.New("tidak memiliki akses untuk menghapus tiket ini")
	}
	return service.tickets.SoftDelete(ticket.ID)
}

func (service *TicketService) GetTicket(user *domain.User, ticketID string) (domain.TicketDTO, error) {
	parsedID, err := parseTicketID(ticketID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	ticket, err := service.tickets.FindByID(parsedID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	scores, err := service.tickets.GetSurveyScores([]int{ticket.ID})
	if err != nil {
		return domain.TicketDTO{}, err
	}
	score := scores[ticket.ID]

	if user == nil {
		if !ticketIsGuest(*ticket) {
			return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk tiket ini")
		}
	}
	if user != nil && !(user.Role == domain.RoleAdmin || ticketOwnedByUser(*ticket, *user) || ticketIsGuest(*ticket)) {
		return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk tiket ini")
	}
	serviceRow := ticket.Service
	serviceRow.GuestAllowed = isGuestServiceID(serviceRow.ID)
	return service.toTicketDTO(*ticket, serviceRow, score), nil
}

func (service *TicketService) ListTickets(user domain.User) ([]domain.TicketDTO, error) {
	var (
		tickets []domain.Ticket
		err     error
	)
	if user.Role == domain.RoleAdmin {
		tickets, err = service.tickets.ListAll()
	} else {
		tickets, err = service.tickets.ListByUser(user)
	}
	if err != nil {
		return nil, err
	}
	scores, err := service.tickets.GetSurveyScores(ticketIDs(tickets))
	if err != nil {
		return nil, err
	}
	return service.mapTickets(tickets, scores), nil
}

func (service *TicketService) ListTicketsPaged(
	user domain.User,
	filter repository.TicketListFilter,
	page int,
	limit int,
) (domain.TicketPageDTO, error) {
	if limit <= 0 {
		limit = 15
	}
	if limit > 50 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}
	if user.Role != domain.RoleAdmin {
		filter.UserID = user.ID
	}

	tickets, total, err := service.tickets.ListFiltered(filter, page, limit)
	if err != nil {
		return domain.TicketPageDTO{}, err
	}
	scores, err := service.tickets.GetSurveyScores(ticketIDs(tickets))
	if err != nil {
		return domain.TicketPageDTO{}, err
	}
	return domain.TicketPageDTO{
		Items:      service.mapTickets(tickets, scores),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: util.CalcTotalPages(total, limit),
	}, nil
}

func (service *TicketService) SearchTickets(query string, guestOnly bool) ([]domain.TicketDTO, error) {
	tickets, err := service.tickets.Search(query, guestOnly)
	if err != nil {
		return nil, err
	}
	scores, err := service.tickets.GetSurveyScores(ticketIDs(tickets))
	if err != nil {
		return nil, err
	}
	return service.mapTickets(tickets, scores), nil
}

func (service *TicketService) resolveCategoryID(serviceID int) (*domain.ServiceCategory, error) {
	if serviceID <= 0 {
		return nil, errors.New("serviceId wajib diisi")
	}
	serviceRow, err := service.categories.FindByID(serviceID)
	if err != nil {
		return nil, errors.New("service tidak ditemukan")
	}
	return serviceRow, nil
}

func (service *TicketService) generateTicketNumber() (string, error) {
	const maxCreateRetries = 10
	for attempt := 0; attempt < maxCreateRetries; attempt++ {
		value, err := randomDigits(6)
		if err != nil {
			return "", err
		}
		exists, err := service.tickets.ExistsTicketNumber(value)
		if err != nil {
			return "", err
		}
		if !exists {
			return value, nil
		}
	}
	return "", errors.New("gagal membuat nomor tiket unik")
}

func randomDigits(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length tidak valid")
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	builder := strings.Builder{}
	builder.Grow(length)
	for index := range bytes {
		builder.WriteByte(byte('0' + (bytes[index] % 10)))
	}
	return builder.String(), nil
}

func (service *TicketService) applyStatus(ticket *domain.Ticket, status domain.TicketStatus) {
	status = normalizeInitialTicketStatus(status)
	ticket.Status = status
	switch status {
	case domain.StatusAssign:
		ticket.IsAssign = true
		ticket.IsDone = false
		ticket.IsReject = false
	case domain.StatusDone:
		ticket.IsAssign = true
		ticket.IsDone = true
		ticket.IsReject = false
	case domain.StatusReject:
		ticket.IsAssign = false
		ticket.IsDone = false
		ticket.IsReject = true
	default:
		ticket.IsAssign = false
		ticket.IsDone = false
		ticket.IsReject = false
	}
}

func surveyRequiredForTicket(ticket domain.Ticket) bool {
	return ticket.Status == domain.StatusDone &&
		stringOrEmpty(ticket.Username) != "" &&
		stringOrEmpty(ticket.NumberID) != ""
}

func ticketIsGuest(ticket domain.Ticket) bool {
	return stringOrEmpty(ticket.Username) == ""
}

func ticketOwnedByUser(ticket domain.Ticket, user domain.User) bool {
	return strings.EqualFold(stringOrEmpty(ticket.Username), strings.TrimSpace(user.Username)) &&
		stringOrEmpty(ticket.NumberID) == strings.TrimSpace(user.ID)
}

func normalizePriority(priority domain.TicketPriority) domain.TicketPriority {
	switch strings.ToUpper(strings.TrimSpace(string(priority))) {
	case string(domain.PriorityLow):
		return domain.PriorityLow
	case string(domain.PriorityHigh):
		return domain.PriorityHigh
	default:
		return domain.PriorityMedium
	}
}

func normalizeEntity(entity string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(entity)) {
	case domain.EntityDosen:
		return domain.EntityDosen, nil
	case domain.EntityTendik:
		return domain.EntityTendik, nil
	case domain.EntityMahasiswa:
		return domain.EntityMahasiswa, nil
	case domain.EntityLainnya:
		return domain.EntityLainnya, nil
	}
	return "", errors.New("entity tidak valid")
}

func (service *TicketService) toTicketDTO(
	ticket domain.Ticket,
	serviceRow domain.ServiceCategory,
	surveyScore float64,
) domain.TicketDTO {
	serviceName := serviceRow.Name
	if serviceName == "" {
		serviceName = ticket.Service.Name
	}
	staffID := ""
	if ticket.StaffID != nil {
		staffID = *ticket.StaffID
	}
	return domain.TicketDTO{
		ID:             strconv.Itoa(ticket.ID),
		TicketNumber:   ticket.TicketNumber,
		TicketDate:     ticket.CreatedAt,
		CreatedAt:      ticket.CreatedAt,
		Username:       stringOrEmpty(ticket.Username),
		NumberID:       stringOrEmpty(ticket.NumberID),
		Name:           ticket.Name,
		Email:          ticket.Email,
		Entity:         ticket.Entity,
		ServiceID:      ticket.ServiceID,
		ServiceName:    serviceName,
		CategoryID:     strconv.Itoa(ticket.ServiceID),
		Category:       serviceName,
		Notes:          ticket.Notes,
		StaffNotes:     ticket.StaffNotes,
		Priority:       ticket.Priority,
		Status:         ticket.Status,
		IsReject:       ticket.IsReject,
		IsAssign:       ticket.IsAssign,
		IsDone:         ticket.IsDone,
		IDStaff:        staffID,
		Lamp1:          ticket.Lamp1,
		Lamp2:          ticket.Lamp2,
		SurveyRequired: ticket.SurveyRequired,
		SurveyScore:    scoreToFivePoint(surveyScore),
	}
}

func normalizeInitialTicketStatus(status domain.TicketStatus) domain.TicketStatus {
	switch status {
	case domain.StatusWaiting, domain.StatusAssign, domain.StatusDone, domain.StatusReject:
		return status
	default:
		return domain.StatusDone
	}
}

func statusLabel(status domain.TicketStatus) string {
	switch status {
	case domain.StatusWaiting:
		return "Menunggu"
	case domain.StatusAssign:
		return "Ditugaskan"
	case domain.StatusDone:
		return "Selesai"
	case domain.StatusReject:
		return "Ditolak"
	default:
		return string(status)
	}
}

func statusChangeNotification(
	ticketNumber string,
	previous domain.TicketStatus,
	current domain.TicketStatus,
	surveyRequired bool,
) (string, string) {
	if surveyRequired {
		return "Tiket Selesai Ditangani",
			fmt.Sprintf("Tiket %s selesai ditangani. Mohon isi feedback.", ticketNumber)
	}
	return "Status Tiket Diperbarui",
		fmt.Sprintf(
			"Status tiket %s diperbarui dari %s menjadi %s.",
			ticketNumber,
			statusLabel(previous),
			statusLabel(current),
		)
}

func (service *TicketService) mapTickets(
	tickets []domain.Ticket,
	scores map[int]float64,
) []domain.TicketDTO {
	result := make([]domain.TicketDTO, 0, len(tickets))
	for _, ticket := range tickets {
		serviceRow := ticket.Service
		serviceRow.GuestAllowed = isGuestServiceID(serviceRow.ID)
		result = append(result, service.toTicketDTO(ticket, serviceRow, scores[ticket.ID]))
	}
	return result
}

func isGuestServiceID(serviceID int) bool {
	switch serviceID {
	case domain.ServiceGuestPassword, domain.ServiceGuestRegistration, domain.ServiceGuestEmail:
		return true
	default:
		return false
	}
}

func ticketIDs(tickets []domain.Ticket) []int {
	ids := make([]int, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}
	return ids
}

func parseTicketID(raw string) (int, error) {
	parsedID, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsedID <= 0 {
		return 0, errors.New("id tiket tidak valid")
	}
	return parsedID, nil
}

func (service *TicketService) notifyTicketStatus(ctx context.Context, ticket domain.Ticket, title string, message string) error {
	numberID := stringOrEmpty(ticket.NumberID)
	if numberID == "" || ticketIsGuest(ticket) {
		return nil
	}
	notification := domain.Notification{
		ID:        util.NewID(64),
		UserID:    numberID,
		TicketID:  ticket.ID,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: service.now(),
	}
	if err := service.notifications.Create(&notification); err != nil {
		log.Printf("failed to create notification: %v", err)
	}

	tokens, err := service.tokens.ListTokens(numberID)
	if err != nil {
		log.Printf("failed to list tokens: %v", err)
		return err
	}
	tokenValues := make([]string, 0, len(tokens))
	for _, token := range tokens {
		tokenValues = append(tokenValues, token.Token)
	}
	invalidTokens, sendErr := service.fcmClient.SendToTokens(ctx, tokenValues, title, message, map[string]string{
		"ticket_id": strconv.Itoa(ticket.ID),
		"title":     title,
		"body":      message,
	})
	if len(invalidTokens) > 0 {
		unique := make([]string, 0, len(invalidTokens))
		seen := make(map[string]struct{}, len(invalidTokens))
		for _, token := range invalidTokens {
			if token == "" {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			unique = append(unique, token)
		}
		if err := service.tokens.DeleteByUserAndTokens(numberID, unique); err != nil {
			log.Printf("failed to delete invalid fcm tokens user=%s count=%d: %v", numberID, len(unique), err)
		}
	}
	return sendErr
}

func cleanOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func scoreZero() float64 {
	return 0
}
