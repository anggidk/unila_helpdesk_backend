package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
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
	attachments   *repository.AttachmentRepository
	fcmClient     *fcm.Client
	initialStatus domain.TicketStatus
	now           func() time.Time
}

type TicketCreateRequest struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Category    string                `json:"category"`
	Priority    domain.TicketPriority `json:"priority"`
	Attachments []string              `json:"attachments"`
}

type GuestTicketCreateRequest struct {
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Category     string                `json:"category"`
	Priority     domain.TicketPriority `json:"priority"`
	Attachments  []string              `json:"attachments"`
	ReporterName string                `json:"reporter_name"`
}

type TicketUpdateRequest struct {
	Title       *string                `json:"title"`
	Description *string                `json:"description"`
	Category    *string                `json:"category"`
	Priority    *domain.TicketPriority `json:"priority"`
	Status      *domain.TicketStatus   `json:"status"`
	Assignee    *string                `json:"assignee"`
}

func NewTicketService(
	tickets *repository.TicketRepository,
	categories *repository.CategoryRepository,
	notifications *repository.NotificationRepository,
	tokens *repository.FCMTokenRepository,
	attachments *repository.AttachmentRepository,
	fcmClient *fcm.Client,
	initialStatus domain.TicketStatus,
) *TicketService {
	return &TicketService{
		tickets:       tickets,
		categories:    categories,
		notifications: notifications,
		tokens:        tokens,
		attachments:   attachments,
		fcmClient:     fcmClient,
		initialStatus: normalizeInitialTicketStatus(initialStatus),
		now:           time.Now,
	}
}

type ticketCoreParams struct {
	title          string
	description    string
	category       string
	priority       domain.TicketPriority
	attachments    []string
	reporterID     string
	reporterName   string
	isGuest        bool
	surveyEligible bool
	historyNote    string
}

func (service *TicketService) createTicketCore(params ticketCoreParams) (domain.Ticket, *domain.ServiceCategory, error) {
	if strings.TrimSpace(params.title) == "" {
		return domain.Ticket{}, nil, errors.New("judul tiket wajib diisi")
	}
	if strings.TrimSpace(params.description) == "" {
		return domain.Ticket{}, nil, errors.New("deskripsi tiket wajib diisi")
	}

	category, err := service.resolveCategory(params.category)
	if err != nil {
		return domain.Ticket{}, nil, err
	}

	if params.isGuest && !category.GuestAllowed {
		return domain.Ticket{}, nil, errors.New("guest hanya dapat membuat tiket keanggotaan")
	}

	priority := params.priority
	if priority == "" {
		priority = domain.PriorityMedium
	}

	const maxCreateRetries = 5
	for attempt := 0; attempt < maxCreateRetries; attempt++ {
		ticketID, err := service.generateTicketID()
		if err != nil {
			return domain.Ticket{}, nil, err
		}

		ticket := domain.Ticket{
			ID:             ticketID,
			Title:          strings.TrimSpace(params.title),
			Description:    strings.TrimSpace(params.description),
			CategoryID:     category.ID,
			Priority:       priority,
			Status:         service.initialStatus,
			ReporterID:     params.reporterID,
			ReporterName:   params.reporterName,
			IsGuest:        params.isGuest,
			SurveyRequired: params.surveyEligible && !params.isGuest && service.initialStatus == domain.StatusResolved,
			CreatedAt:      service.now(),
			UpdatedAt:      service.now(),
		}
		if payload := marshalAttachments(params.attachments); payload != nil {
			ticket.Attachments = payload
		}

		if err := service.tickets.Create(&ticket); err != nil {
			if isDuplicateTicketIDError(err) {
				continue
			}
			return domain.Ticket{}, nil, err
		}

		_ = service.attachments.AttachToTicket(attachmentIDsFromRefs(params.attachments), ticket.ID)
		_ = service.addHistory(ticket.ID, "Ticket Created", params.historyNote)
		_ = service.addHistory(ticket.ID, "Status Updated", fmt.Sprintf("Status diperbarui ke %s", ticket.Status))
		return ticket, category, nil
	}

	return domain.Ticket{}, nil, errors.New("gagal membuat nomor tiket unik, silakan coba lagi")
}

func (service *TicketService) CreateTicket(ctx context.Context, user domain.User, req TicketCreateRequest) (domain.TicketDTO, error) {
	ticket, category, err := service.createTicketCore(ticketCoreParams{
		title:          req.Title,
		description:    req.Description,
		category:       req.Category,
		priority:       req.Priority,
		attachments:    req.Attachments,
		reporterID:     user.ID,
		reporterName:   user.Name,
		isGuest:        user.Role == domain.RoleGuest,
		surveyEligible: user.Role == domain.RoleRegistered,
		historyNote:    "Dilaporkan oleh pengguna",
	})
	if err != nil {
		return domain.TicketDTO{}, err
	}

	if !ticket.IsGuest {
		if err := service.notifyTicketStatus(
			ctx,
			ticket,
			"Tiket Berhasil Dibuat",
			fmt.Sprintf("Tiket %s berhasil dibuat dengan status %s.", ticket.ID, statusLabel(ticket.Status)),
		); err != nil {
			log.Printf("failed to send create notification: %v", err)
		}
	}

	return service.toTicketDTO(ticket, *category, 0), nil
}

func (service *TicketService) CreateGuestTicket(ctx context.Context, req GuestTicketCreateRequest) (domain.TicketDTO, error) {
	reporterName := strings.TrimSpace(req.ReporterName)
	if reporterName == "" {
		reporterName = "Guest User"
	}

	ticket, category, err := service.createTicketCore(ticketCoreParams{
		title:        req.Title,
		description:  req.Description,
		category:     req.Category,
		priority:     req.Priority,
		attachments:  req.Attachments,
		reporterID:   "",
		reporterName: reporterName,
		isGuest:      true,
		historyNote:  "Dilaporkan oleh guest",
	})
	if err != nil {
		return domain.TicketDTO{}, err
	}

	return service.toTicketDTO(ticket, *category, 0), nil
}

func (service *TicketService) UpdateTicket(ctx context.Context, user domain.User, ticketID string, req TicketUpdateRequest) (domain.TicketDTO, error) {
	ticket, err := service.tickets.FindByID(ticketID)
	if err != nil {
		return domain.TicketDTO{}, err
	}

	if user.Role != domain.RoleAdmin && ticket.ReporterID != user.ID {
		return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk memperbarui tiket ini")
	}

	// User biasa tidak bisa mengedit tiket yang sudah selesai atau ditutup
	if user.Role != domain.RoleAdmin && (ticket.Status == domain.StatusResolved) {
		return domain.TicketDTO{}, errors.New("tiket yang sudah selesai tidak dapat diedit")
	}

	if req.Title != nil {
		ticket.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		ticket.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		category, err := service.resolveCategory(*req.Category)
		if err != nil {
			return domain.TicketDTO{}, err
		}
		if user.Role == domain.RoleGuest && !category.GuestAllowed {
			return domain.TicketDTO{}, errors.New("guest hanya dapat membuat tiket keanggotaan")
		}
		ticket.CategoryID = category.ID
		ticket.Category = *category
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}

	statusChanged := false
	previousStatus := ticket.Status
	historyTitle := "Ticket Updated"
	historyDesc := "Perubahan tiket diperbarui"

	// User biasa tidak bisa mengubah status, assignee - hanya admin
	if user.Role == domain.RoleAdmin {
		if req.Status != nil && ticket.Status != *req.Status {
			ticket.Status = *req.Status
			statusChanged = true
		}
		if req.Assignee != nil {
			ticket.Assignee = strings.TrimSpace(*req.Assignee)
		}
	}

	if statusChanged {
		historyTitle = "Status Updated"
		historyDesc = fmt.Sprintf("Status diperbarui dari %s ke %s", statusLabel(previousStatus), statusLabel(ticket.Status))
	}

	ticket.UpdatedAt = service.now()
	if err := service.tickets.Update(ticket); err != nil {
		return domain.TicketDTO{}, err
	}

	if err := service.addHistory(ticket.ID, historyTitle, historyDesc); err != nil {
		log.Printf("failed to add ticket history: %v", err)
	}

	if statusChanged {
		surveyRequired := ticket.Status == domain.StatusResolved && !ticket.IsGuest
		ticket.SurveyRequired = surveyRequired
		if err := service.tickets.UpdateStatus(ticket.ID, ticket.Status, surveyRequired); err != nil {
			log.Printf("failed to update status: %v", err)
		}
		title, message := statusChangeNotification(ticket.ID, previousStatus, ticket.Status, surveyRequired)
		if err := service.notifyTicketStatus(ctx, *ticket, title, message); err != nil {
			log.Printf("failed to send status notification: %v", err)
		}
	}

	return service.toTicketDTO(*ticket, ticket.Category, 0), nil
}

func (service *TicketService) DeleteTicket(user domain.User, ticketID string) error {
	ticket, err := service.tickets.FindByID(ticketID)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleAdmin && ticket.ReporterID != user.ID {
		return errors.New("tidak memiliki akses untuk menghapus tiket ini")
	}
	return service.tickets.SoftDelete(ticketID)
}

func (service *TicketService) GetTicket(user *domain.User, ticketID string) (domain.TicketDTO, error) {
	ticket, err := service.tickets.FindByID(ticketID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	scores, err := service.tickets.GetSurveyScores([]string{ticket.ID})
	if err != nil {
		return domain.TicketDTO{}, err
	}
	score := scores[ticket.ID]
	if user == nil {
		return service.toTicketDTO(*ticket, ticket.Category, score), nil
	}
	if user.Role == domain.RoleAdmin || ticket.ReporterID == user.ID {
		return service.toTicketDTO(*ticket, ticket.Category, score), nil
	}
	if ticket.IsGuest {
		return service.toTicketDTO(*ticket, ticket.Category, score), nil
	}
	return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk tiket ini")
}

func (service *TicketService) ListTickets(user domain.User) ([]domain.TicketDTO, error) {
	var tickets []domain.Ticket
	var err error
	if user.Role == domain.RoleAdmin {
		tickets, err = service.tickets.ListAll()
	} else {
		tickets, err = service.tickets.ListByUser(user.ID)
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
		filter.ReporterID = user.ID
	}

	tickets, total, err := service.tickets.ListFiltered(filter, page, limit)
	if err != nil {
		return domain.TicketPageDTO{}, err
	}
	scores, err := service.tickets.GetSurveyScores(ticketIDs(tickets))
	if err != nil {
		return domain.TicketPageDTO{}, err
	}
	totalPages := util.CalcTotalPages(total, limit)
	return domain.TicketPageDTO{
		Items:      service.mapTickets(tickets, scores),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
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

func (service *TicketService) AddComment(user domain.User, ticketID string, message string) (domain.TicketDTO, error) {
	if strings.TrimSpace(message) == "" {
		return domain.TicketDTO{}, errors.New("komentar tidak boleh kosong")
	}
	ticket, err := service.tickets.FindByID(ticketID)
	if err != nil {
		return domain.TicketDTO{}, err
	}
	if user.Role != domain.RoleAdmin && ticket.ReporterID != user.ID {
		return domain.TicketDTO{}, errors.New("tidak memiliki akses untuk menambah komentar")
	}

	comment := domain.TicketComment{
		ID:        util.NewUUID(),
		TicketID:  ticket.ID,
		Author:    user.Name,
		Message:   strings.TrimSpace(message),
		IsStaff:   user.Role == domain.RoleAdmin,
		Timestamp: service.now(),
	}
	if err := service.tickets.AddComment(&comment); err != nil {
		return domain.TicketDTO{}, err
	}
	return service.toTicketDTO(*ticket, ticket.Category, 0), nil
}

func (service *TicketService) resolveCategory(value string) (*domain.ServiceCategory, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("kategori wajib diisi")
	}
	category, err := service.categories.FindByID(value)
	if err == nil {
		return category, nil
	}
	category, err = service.categories.FindByName(value)
	if err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}
	return category, nil
}

func (service *TicketService) generateTicketID() (string, error) {
	year := service.now().Year()
	sequence, err := service.tickets.NextTicketSequence(year)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TK-%d-%03d", year, sequence), nil
}

func (service *TicketService) toTicketDTO(ticket domain.Ticket, category domain.ServiceCategory, surveyScore float64) domain.TicketDTO {
	history := make([]domain.TicketHistoryDTO, 0, len(ticket.History))
	for _, item := range ticket.History {
		history = append(history, domain.TicketHistoryDTO{
			Title:       item.Title,
			Description: item.Description,
			Timestamp:   item.Timestamp,
		})
	}
	comments := make([]domain.TicketCommentDTO, 0, len(ticket.Comments))
	for _, item := range ticket.Comments {
		comments = append(comments, domain.TicketCommentDTO{
			Author:    item.Author,
			Message:   item.Message,
			Timestamp: item.Timestamp,
			IsStaff:   item.IsStaff,
		})
	}

	categoryName := category.Name
	categoryID := category.ID
	if categoryName == "" {
		categoryName = ticket.Category.Name
	}
	if categoryID == "" {
		categoryID = ticket.Category.ID
	}

	surveyScore = normalizeLegacyScore(surveyScore)
	attachments := []string{}
	if len(ticket.Attachments) > 0 {
		_ = json.Unmarshal(ticket.Attachments, &attachments)
	}
	return domain.TicketDTO{
		ID:             ticket.ID,
		Title:          ticket.Title,
		Description:    ticket.Description,
		Category:       categoryName,
		CategoryID:     categoryID,
		Status:         ticket.Status,
		Priority:       ticket.Priority,
		CreatedAt:      ticket.CreatedAt,
		Reporter:       ticket.ReporterName,
		IsGuest:        ticket.IsGuest,
		Assignee:       ticket.Assignee,
		Attachments:    attachments,
		History:        history,
		Comments:       comments,
		SurveyRequired: ticket.SurveyRequired,
		SurveyScore:    surveyScore,
	}
}

func marshalAttachments(values []string) []byte {
	if len(values) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item != "" {
			cleaned = append(cleaned, item)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	payload, err := json.Marshal(cleaned)
	if err != nil {
		return nil
	}
	return payload
}

func attachmentIDsFromRefs(refs []string) []string {
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		cleaned := strings.TrimSpace(ref)
		if cleaned == "" {
			continue
		}
		parsed, err := url.Parse(cleaned)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			ids = append(ids, cleaned)
			continue
		}
		base := path.Base(parsed.Path)
		if base != "" && base != "." && base != "/" && base != "uploads" {
			ids = append(ids, base)
		}
	}
	return ids
}

func isDuplicateTicketIDError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate key value") &&
		strings.Contains(message, "tickets_pkey") {
		return true
	}
	return false
}

func normalizeInitialTicketStatus(status domain.TicketStatus) domain.TicketStatus {
	switch status {
	case domain.StatusWaiting, domain.StatusInProgress, domain.StatusResolved:
		return status
	case domain.StatusProcessing:
		return domain.StatusInProgress
	default:
		return domain.StatusResolved
	}
}

func statusLabel(status domain.TicketStatus) string {
	switch status {
	case domain.StatusWaiting:
		return "Menunggu"
	case domain.StatusInProgress:
		return "Progres"
	case domain.StatusProcessing:
		return "Progres"
	case domain.StatusResolved:
		return "Selesai"
	default:
		return string(status)
	}
}

func statusChangeNotification(
	ticketID string,
	previous domain.TicketStatus,
	current domain.TicketStatus,
	surveyRequired bool,
) (string, string) {
	if surveyRequired {
		return "Tiket Selesai Ditangani",
			fmt.Sprintf("Tiket %s selesai ditangani. Mohon isi feedback.", ticketID)
	}
	return "Status Tiket Diperbarui",
		fmt.Sprintf(
			"Tiket %s berubah dari %s ke %s.",
			ticketID,
			statusLabel(previous),
			statusLabel(current),
		)
}

func (service *TicketService) mapTickets(tickets []domain.Ticket, scores map[string]float64) []domain.TicketDTO {
	result := make([]domain.TicketDTO, 0, len(tickets))
	for _, ticket := range tickets {
		score := scores[ticket.ID]
		result = append(result, service.toTicketDTO(ticket, ticket.Category, score))
	}
	return result
}

func ticketIDs(tickets []domain.Ticket) []string {
	ids := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}
	return ids
}

func (service *TicketService) addHistory(ticketID, title, description string) error {
	return service.tickets.AddHistory(&domain.TicketHistory{
		ID:          util.NewUUID(),
		TicketID:    ticketID,
		Title:       title,
		Description: description,
		Timestamp:   service.now(),
	})
}

func (service *TicketService) notifyTicketStatus(ctx context.Context, ticket domain.Ticket, title string, message string) error {
	if strings.TrimSpace(ticket.ReporterID) == "" {
		return nil
	}
	notification := domain.Notification{
		ID:        util.NewUUID(),
		UserID:    ticket.ReporterID,
		TicketID:  ticket.ID,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: service.now(),
	}
	if err := service.notifications.Create(&notification); err != nil {
		log.Printf("failed to create notification: %v", err)
	}

	tokens, err := service.tokens.ListTokens(ticket.ReporterID)
	if err != nil {
		log.Printf("failed to list tokens: %v", err)
		return err
	}
	tokenValues := make([]string, 0, len(tokens))
	for _, token := range tokens {
		tokenValues = append(tokenValues, token.Token)
	}
	invalidTokens, sendErr := service.fcmClient.SendToTokens(ctx, tokenValues, title, message, map[string]string{
		"ticket_id": ticket.ID,
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
		if err := service.tokens.DeleteByUserAndTokens(ticket.ReporterID, unique); err != nil {
			log.Printf("failed to delete invalid fcm tokens user=%s count=%d: %v", ticket.ReporterID, len(unique), err)
		} else {
			log.Printf("deleted invalid fcm tokens user=%s count=%d", ticket.ReporterID, len(unique))
		}
	}
	return sendErr
}
