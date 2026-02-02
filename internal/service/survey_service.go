package service

import (
    "encoding/json"
    "errors"
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

    response := domain.SurveyResponse{
        ID:        util.NewUUID(),
        TicketID:  ticket.ID,
        UserID:    user.ID,
        Answers:   payload,
        CreatedAt: service.now(),
    }
    if err := service.surveys.SaveResponse(&response); err != nil {
        return err
    }

    ticket.SurveyRequired = false
    _ = service.tickets.Update(ticket)

    return nil
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
        CategoryID:  template.CategoryID,
        Questions:   questions,
    }
}
