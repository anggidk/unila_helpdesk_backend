package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
	"unila_helpdesk_backend/internal/util"
)

type SurveyService struct {
	surveys *repository.SurveyRepository
	tickets *repository.TicketRepository
	now     func() time.Time
}

type SurveyTemplateRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Framework   string                  `json:"framework"`
	Questions   []SurveyQuestionRequest `json:"questions"`
}

type SurveyQuestionRequest struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

type SurveyResponseRequest struct {
	TicketID string                 `json:"ticket_id"`
	Answers  map[string]interface{} `json:"answers"`
}

func NewSurveyService(
	surveys *repository.SurveyRepository,
	tickets *repository.TicketRepository,
) *SurveyService {
	return &SurveyService{
		surveys: surveys,
		tickets: tickets,
		now:     time.Now,
	}
}

func (service *SurveyService) ListTemplates() ([]domain.SurveyTemplateDTO, error) {
	templates, err := service.surveys.ListTemplates()
	if err != nil {
		return nil, err
	}
	return mapSurveyTemplates(templates), nil
}

func (service *SurveyService) TemplateByCategory(categoryID string) (domain.SurveyTemplateDTO, error) {
	template, err := service.surveys.FindByCategory(categoryID)
	if err != nil {
		return domain.SurveyTemplateDTO{}, err
	}
	return mapSurveyTemplate(*template), nil
}

func (service *SurveyService) CreateTemplate(req SurveyTemplateRequest) (domain.SurveyTemplateDTO, error) {
	if strings.TrimSpace(req.Title) == "" {
		return domain.SurveyTemplateDTO{}, errors.New("judul template wajib diisi")
	}

	template := domain.SurveyTemplate{
		ID:          util.NewID(12),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Framework:   strings.TrimSpace(req.Framework),
		CreatedAt:   service.now(),
		UpdatedAt:   service.now(),
	}

	template.Questions = buildSurveyQuestions(req.Questions, template.ID, service.now())
	if len(template.Questions) == 0 {
		return domain.SurveyTemplateDTO{}, errors.New("template wajib memiliki minimal satu pertanyaan")
	}
	if err := service.surveys.CreateTemplate(&template); err != nil {
		return domain.SurveyTemplateDTO{}, err
	}
	return mapSurveyTemplate(template), nil
}

func (service *SurveyService) UpdateTemplate(templateID string, req SurveyTemplateRequest) (domain.SurveyTemplateDTO, error) {
	if strings.TrimSpace(templateID) == "" {
		return domain.SurveyTemplateDTO{}, errors.New("template id wajib diisi")
	}
	if strings.TrimSpace(req.Title) == "" {
		return domain.SurveyTemplateDTO{}, errors.New("judul template wajib diisi")
	}

	template, err := service.surveys.FindByID(templateID)
	if err != nil {
		return domain.SurveyTemplateDTO{}, err
	}

	template.Title = strings.TrimSpace(req.Title)
	template.Description = strings.TrimSpace(req.Description)
	template.Framework = strings.TrimSpace(req.Framework)
	template.UpdatedAt = service.now()
	template.Questions = buildSurveyQuestions(req.Questions, template.ID, service.now())

	if err := service.surveys.ReplaceTemplate(template); err != nil {
		return domain.SurveyTemplateDTO{}, err
	}
	return mapSurveyTemplate(*template), nil
}

func buildSurveyQuestions(
	requests []SurveyQuestionRequest,
	templateID string,
	createdAt time.Time,
) []domain.SurveyQuestion {
	questions := make([]domain.SurveyQuestion, 0, len(requests))
	for _, question := range requests {
		if strings.TrimSpace(question.Text) == "" {
			continue
		}
		questionID := strings.TrimSpace(question.ID)
		if questionID == "" {
			questionID = util.NewID(32)
		}
		options, _ := json.Marshal(question.Options)
		questions = append(questions, domain.SurveyQuestion{
			ID:         questionID,
			TemplateID: templateID,
			Text:       strings.TrimSpace(question.Text),
			Type:       domain.SurveyQuestionType(question.Type),
			Options:    options,
			CreatedAt:  createdAt,
		})
	}
	return questions
}

func (service *SurveyService) DeleteTemplate(templateID string) error {
	if strings.TrimSpace(templateID) == "" {
		return errors.New("template id wajib diisi")
	}
	return service.surveys.DeleteTemplate(templateID)
}

func (service *SurveyService) SubmitSurvey(user domain.User, req SurveyResponseRequest) error {
	if user.Role != domain.RoleRegistered {
		return errors.New("hanya pengguna terdaftar yang dapat mengisi survey")
	}

	ticketID, err := strconv.Atoi(strings.TrimSpace(req.TicketID))
	if err != nil || ticketID <= 0 {
		return errors.New("ticket_id wajib diisi")
	}

	ticket, err := service.tickets.FindByID(ticketID)
	if err != nil {
		return err
	}
	if ticket.Status != domain.StatusDone {
		return errors.New("survey hanya tersedia untuk tiket selesai")
	}
	if stringOrEmpty(ticket.Username) == "" || stringOrEmpty(ticket.NumberID) == "" {
		return errors.New("survey tidak tersedia untuk tiket guest")
	}
	if stringOrEmpty(ticket.NumberID) != strings.TrimSpace(user.ID) {
		return errors.New("tidak memiliki akses untuk tiket ini")
	}
	if username := strings.TrimSpace(user.Username); username != "" &&
		!strings.EqualFold(stringOrEmpty(ticket.Username), username) {
		return errors.New("tidak memiliki akses untuk tiket ini")
	}

	hasResponse, err := service.surveys.HasResponse(ticket.ID, user.ID)
	if err != nil {
		return err
	}
	if hasResponse {
		return errors.New("survey sudah diisi")
	}

	template, err := service.surveys.FindByCategory(strconv.Itoa(ticket.ServiceID))
	if err != nil || template == nil || strings.TrimSpace(template.ID) == "" {
		return errors.New("template survei untuk layanan tiket tidak ditemukan")
	}

	responseID := util.NewID(32)
	createdAt := service.now()
	items, err := buildSurveyResponseItems(responseID, req.Answers, template, createdAt)
	if err != nil {
		return err
	}
	if len(template.Questions) > 0 && len(items) == 0 {
		return errors.New("jawaban wajib diisi")
	}

	response := domain.SurveyResponse{
		ID:         responseID,
		TicketID:   ticket.ID,
		UserID:     user.ID,
		TemplateID: template.ID,
		Score:      calculateSurveyScore(req.Answers, template),
		CreatedAt:  createdAt,
	}
	if err := service.surveys.SaveResponse(&response, items); err != nil {
		return err
	}

	ticket.SurveyRequired = false
	return service.tickets.Update(ticket)
}

func (service *SurveyService) ListResponsesPaged(
	filter repository.SurveyResponseFilter,
	page int,
	limit int,
) (domain.SurveyResponsePageDTO, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	rows, total, err := service.surveys.ListResponses(filter, page, limit)
	if err != nil {
		return domain.SurveyResponsePageDTO{}, err
	}
	items := make([]domain.SurveyResponseItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.SurveyResponseItemDTO{
			ID:         row.ID,
			TicketID:   strconv.Itoa(row.TicketID),
			UserID:     row.UserID,
			UserName:   row.UserName,
			UserEmail:  row.UserEmail,
			UserEntity: row.UserEntity,
			CategoryID: strconv.Itoa(row.ServiceID),
			Category:   row.ServiceName,
			TemplateID: row.TemplateID,
			Template:   row.TemplateTitle,
			Score:      scoreToFivePoint(row.Score),
			CreatedAt:  row.CreatedAt,
		})
	}
	totalPages := util.CalcTotalPages(total, limit)
	return domain.SurveyResponsePageDTO{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func mapSurveyTemplates(templates []domain.SurveyTemplate) []domain.SurveyTemplateDTO {
	result := make([]domain.SurveyTemplateDTO, 0, len(templates))
	for _, template := range templates {
		result = append(result, mapSurveyTemplate(template))
	}
	return result
}

func mapSurveyTemplate(template domain.SurveyTemplate) domain.SurveyTemplateDTO {
	questions := make([]domain.SurveyQuestionDTO, 0, len(template.Questions))
	for _, question := range template.Questions {
		var options []string
		_ = json.Unmarshal(question.Options, &options)
		questions = append(questions, domain.SurveyQuestionDTO{
			ID:      question.ID,
			Text:    question.Text,
			Type:    string(question.Type),
			Options: options,
		})
	}
	return domain.SurveyTemplateDTO{
		ID:          template.ID,
		Title:       template.Title,
		Description: template.Description,
		Framework:   template.Framework,
		Questions:   questions,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func calculateSurveyScore(answers map[string]interface{}, template *domain.SurveyTemplate) float64 {
	if len(answers) == 0 || template == nil || len(template.Questions) == 0 {
		return 0
	}
	var total float64
	var count int
	for _, question := range template.Questions {
		value, ok := answers[question.ID]
		if !ok {
			continue
		}
		if score, ok := scoreFromQuestionValue(value, question.Type); ok {
			total += score
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func buildSurveyResponseItems(
	responseID string,
	answers map[string]interface{},
	template *domain.SurveyTemplate,
	createdAt time.Time,
) ([]domain.SurveyResponseItem, error) {
	if len(answers) == 0 || template == nil || len(template.Questions) == 0 {
		return []domain.SurveyResponseItem{}, nil
	}
	items := make([]domain.SurveyResponseItem, 0, len(template.Questions))
	for _, question := range template.Questions {
		value, ok := answers[question.ID]
		if !ok {
			continue
		}
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}

		item := domain.SurveyResponseItem{
			ID:          util.NewID(32),
			ResponseID:  responseID,
			QuestionID:  question.ID,
			AnswerValue: payload,
			CreatedAt:   createdAt,
		}
		if score, ok := scoreFromQuestionValue(value, question.Type); ok {
			scoreValue := score
			item.ScoreValue = &scoreValue
		}
		items = append(items, item)
	}
	return items, nil
}
