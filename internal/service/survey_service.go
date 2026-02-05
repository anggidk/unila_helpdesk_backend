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
    surveys   *repository.SurveyRepository
    tickets   *repository.TicketRepository
    now       func() time.Time
}

type SurveyTemplateRequest struct {
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Framework   string                 `json:"framework"`
    CategoryID  string                 `json:"categoryId"`
    Questions   []SurveyQuestionRequest `json:"questions"`
}

type SurveyQuestionRequest struct {
    Text    string   `json:"text"`
    Type    string   `json:"type"`
    Options []string `json:"options"`
}

type SurveyResponseRequest struct {
    TicketID string                 `json:"ticket_id"`
    Answers  map[string]interface{} `json:"answers"`
}

func NewSurveyService(surveys *repository.SurveyRepository, tickets *repository.TicketRepository) *SurveyService {
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
    if strings.TrimSpace(req.CategoryID) == "" {
        return domain.SurveyTemplateDTO{}, errors.New("kategori wajib diisi")
    }

    template := domain.SurveyTemplate{
        ID:          util.NewUUID(),
        Title:       strings.TrimSpace(req.Title),
        Description: strings.TrimSpace(req.Description),
        Framework:   strings.TrimSpace(req.Framework),
        CategoryID:  req.CategoryID,
        CreatedAt:   service.now(),
        UpdatedAt:   service.now(),
    }

    questions := make([]domain.SurveyQuestion, 0, len(req.Questions))
    for _, question := range req.Questions {
        if strings.TrimSpace(question.Text) == "" {
            continue
        }
        options, _ := json.Marshal(question.Options)
        questions = append(questions, domain.SurveyQuestion{
            ID:         util.NewUUID(),
            TemplateID: template.ID,
            Text:       strings.TrimSpace(question.Text),
            Type:       domain.SurveyQuestionType(question.Type),
            Options:    options,
            CreatedAt:  service.now(),
        })
    }
    template.Questions = questions

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
    if strings.TrimSpace(req.CategoryID) == "" {
        return domain.SurveyTemplateDTO{}, errors.New("kategori wajib diisi")
    }

    template, err := service.surveys.FindByID(templateID)
    if err != nil {
        return domain.SurveyTemplateDTO{}, err
    }

    template.Title = strings.TrimSpace(req.Title)
    template.Description = strings.TrimSpace(req.Description)
    template.Framework = strings.TrimSpace(req.Framework)
    template.CategoryID = req.CategoryID
    template.UpdatedAt = service.now()

    questions := make([]domain.SurveyQuestion, 0, len(req.Questions))
    for _, question := range req.Questions {
        if strings.TrimSpace(question.Text) == "" {
            continue
        }
        options, _ := json.Marshal(question.Options)
        questions = append(questions, domain.SurveyQuestion{
            ID:         util.NewUUID(),
            TemplateID: template.ID,
            Text:       strings.TrimSpace(question.Text),
            Type:       domain.SurveyQuestionType(question.Type),
            Options:    options,
            CreatedAt:  service.now(),
        })
    }
    template.Questions = questions

    if err := service.surveys.ReplaceTemplate(template); err != nil {
        return domain.SurveyTemplateDTO{}, err
    }
    return mapSurveyTemplate(*template), nil
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
    if strings.TrimSpace(req.TicketID) == "" {
        return errors.New("ticket_id wajib diisi")
    }
    ticket, err := service.tickets.FindByID(req.TicketID)
    if err != nil {
        return err
    }
    if ticket.ReporterID != user.ID {
        return errors.New("tidak memiliki akses untuk tiket ini")
    }
    if ticket.Status != domain.StatusResolved {
        return errors.New("survey hanya tersedia untuk tiket selesai")
    }

    hasResponse, err := service.surveys.HasResponse(ticket.ID, user.ID)
    if err != nil {
        return err
    }
    if hasResponse {
        return errors.New("survey sudah diisi")
    }

    payload, err := json.Marshal(req.Answers)
    if err != nil {
        return err
    }

    template, _ := service.surveys.FindByCategory(ticket.CategoryID)
    templateID := ""
    if template != nil {
        templateID = template.ID
    }

    response := domain.SurveyResponse{
        ID:        util.NewUUID(),
        TicketID:  ticket.ID,
        UserID:    user.ID,
        TemplateID: templateID,
        Answers:   payload,
        Score:     calculateSurveyScore(req.Answers, template),
        CreatedAt: service.now(),
    }
    if err := service.surveys.SaveResponse(&response); err != nil {
        return err
    }

    ticket.SurveyRequired = false
    _ = service.tickets.Update(ticket)

    return nil
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
            TicketID:   row.TicketID,
            UserID:     row.UserID,
            UserName:   row.UserName,
            UserEmail:  row.UserEmail,
            UserEntity: row.UserEntity,
            CategoryID: row.CategoryID,
            Category:   row.CategoryName,
            TemplateID: row.TemplateID,
            Template:   row.TemplateTitle,
            Score:      row.Score,
            CreatedAt:  row.CreatedAt,
        })
    }
    totalPages := 0
    if limit > 0 {
        totalPages = int((total + int64(limit) - 1) / int64(limit))
    }
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
        CategoryID:  template.CategoryID,
        Questions:   questions,
        CreatedAt:   template.CreatedAt,
        UpdatedAt:   template.UpdatedAt,
    }
}

func calculateSurveyScore(answers map[string]interface{}, template *domain.SurveyTemplate) float64 {
    if len(answers) == 0 {
        return 0
    }
    if template == nil || len(template.Questions) == 0 {
        return calculateLegacyScore(answers)
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

func calculateLegacyScore(answers map[string]interface{}) float64 {
    if len(answers) == 0 {
        return 0
    }
    var total float64
    var count int
    for _, value := range answers {
        switch v := value.(type) {
		case float64:
			if v >= 1 && v <= 5 {
				total += normalizeToHundred(v, 5)
				count++
			}
		case int:
			if v >= 1 && v <= 5 {
				total += normalizeToHundred(float64(v), 5)
				count++
			}
		case bool:
			if v {
				total += 100
			} else {
				total += 0
			}
			count++
        case string:
            cleaned := strings.ToLower(strings.TrimSpace(v))
			if cleaned == "ya" || cleaned == "yes" || cleaned == "true" {
				total += 100
				count++
				continue
			}
			if cleaned == "tidak" || cleaned == "no" || cleaned == "false" {
				total += 0
				count++
				continue
			}
			if parsed, err := strconv.ParseFloat(cleaned, 64); err == nil {
				if parsed >= 1 && parsed <= 5 {
					total += normalizeToHundred(parsed, 5)
					count++
				}
			}
        }
    }
    if count == 0 {
        return 0
    }
    return total / float64(count)
}
