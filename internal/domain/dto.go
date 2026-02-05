package domain

import (
    "time"

    "gorm.io/datatypes"
)

type UserDTO struct {
    ID     string   `json:"id"`
    Name   string   `json:"name"`
    Email  string   `json:"email"`
    Role   UserRole `json:"role"`
    Entity string   `json:"entity"`
}

type TicketHistoryDTO struct {
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Timestamp   time.Time `json:"timestamp"`
}

type TicketCommentDTO struct {
    Author    string    `json:"author"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    IsStaff   bool      `json:"isStaff"`
}

type TicketDTO struct {
    ID         string            `json:"id"`
    Title      string            `json:"title"`
    Description string           `json:"description"`
    Category   string            `json:"category"`
    CategoryID string            `json:"categoryId"`
    Status     TicketStatus      `json:"status"`
    Priority   TicketPriority    `json:"priority"`
    CreatedAt  time.Time         `json:"createdAt"`
    Reporter   string            `json:"reporter"`
    IsGuest    bool              `json:"isGuest"`
    Assignee   string            `json:"assignee,omitempty"`
    History    []TicketHistoryDTO `json:"history"`
    Comments   []TicketCommentDTO `json:"comments"`
    SurveyRequired bool          `json:"surveyRequired"`
    SurveyScore float64          `json:"surveyScore"`
}

type TicketPageDTO struct {
    Items      []TicketDTO `json:"items"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    Total      int64       `json:"total"`
    TotalPages int         `json:"totalPages"`
}

type ServiceCategoryDTO struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    GuestAllowed bool   `json:"guestAllowed"`
    TemplateID   string `json:"templateId,omitempty"`
}

type SurveyQuestionDTO struct {
    ID      string   `json:"id"`
    Text    string   `json:"text"`
    Type    string   `json:"type"`
    Options []string `json:"options"`
}

type SurveyTemplateDTO struct {
    ID          string              `json:"id"`
    Title       string              `json:"title"`
    Description string              `json:"description"`
    Framework   string              `json:"framework"`
    CategoryID  string              `json:"categoryId"`
    Questions   []SurveyQuestionDTO `json:"questions"`
    CreatedAt   time.Time           `json:"createdAt"`
    UpdatedAt   time.Time           `json:"updatedAt"`
}

type NotificationDTO struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    IsRead    bool      `json:"isRead"`
}

type CohortRowDTO struct {
    Label        string  `json:"label"`
    Users        int     `json:"users"`
    Retention    []int   `json:"retention"`
    AvgScore     float64 `json:"avgScore"`
    ResponseRate float64 `json:"responseRate"`
}

type ServiceTrendDTO struct {
    Label      string  `json:"label"`
    Percentage float64 `json:"percentage"`
    Note       string  `json:"note"`
}

type DashboardSummaryDTO struct {
    TotalTickets       int     `json:"totalTickets"`
    OpenTickets        int     `json:"openTickets"`
    ResolvedThisPeriod int     `json:"resolvedThisPeriod"`
    AvgRating          float64 `json:"avgRating"`
}

type ServiceSatisfactionDTO struct {
    CategoryID string  `json:"categoryId"`
    Label      string  `json:"label"`
    AvgScore   float64 `json:"avgScore"`
    Responses  int     `json:"responses"`
    Percentage float64 `json:"percentage"`
}

type UsageCohortRowDTO struct {
    Label   string `json:"label"`
    Tickets int    `json:"tickets"`
    Surveys int    `json:"surveys"`
}

type SurveySatisfactionRowDTO struct {
    QuestionID string  `json:"questionId"`
    Question   string  `json:"question"`
    Type       string  `json:"type"`
    AvgScore   float64 `json:"avgScore"`
    Responses  int     `json:"responses"`
}

type SurveySatisfactionDTO struct {
    TemplateID string                     `json:"templateId"`
    Template   string                     `json:"template"`
    CategoryID string                     `json:"categoryId"`
    Category   string                     `json:"category"`
    Period     string                     `json:"period"`
    Start      time.Time                  `json:"start"`
    End        time.Time                  `json:"end"`
    Rows       []SurveySatisfactionRowDTO `json:"rows"`
}

type SurveySatisfactionExportQuestionDTO struct {
    ID   string `json:"id"`
    Text string `json:"text"`
    Type string `json:"type"`
}

type SurveySatisfactionExportResponseDTO struct {
    ID        string         `json:"id"`
    TicketID  string         `json:"ticketId"`
    UserID    string         `json:"userId"`
    Score     float64        `json:"score"`
    CreatedAt time.Time      `json:"createdAt"`
    Answers   datatypes.JSON `json:"answers"`
}

type SurveySatisfactionExportDTO struct {
    TemplateID string                              `json:"templateId"`
    Template   string                              `json:"template"`
    CategoryID string                              `json:"categoryId"`
    Category   string                              `json:"category"`
    Period     string                              `json:"period"`
    Start      time.Time                           `json:"start"`
    End        time.Time                           `json:"end"`
    Questions  []SurveySatisfactionExportQuestionDTO `json:"questions"`
    Responses  []SurveySatisfactionExportResponseDTO `json:"responses"`
}

type SurveyResponseItemDTO struct {
    ID          string    `json:"id"`
    TicketID    string    `json:"ticketId"`
    UserID      string    `json:"userId"`
    UserName    string    `json:"userName"`
    UserEmail   string    `json:"userEmail"`
    UserEntity  string    `json:"userEntity"`
    CategoryID  string    `json:"categoryId"`
    Category    string    `json:"category"`
    TemplateID  string    `json:"templateId"`
    Template    string    `json:"template"`
    Score       float64   `json:"score"`
    CreatedAt   time.Time `json:"createdAt"`
}

type SurveyResponsePageDTO struct {
    Items      []SurveyResponseItemDTO `json:"items"`
    Page       int                     `json:"page"`
    Limit      int                     `json:"limit"`
    Total      int64                   `json:"total"`
    TotalPages int                     `json:"totalPages"`
}

type EntityServiceDTO struct {
    Entity     string `json:"entity"`
    CategoryID string `json:"categoryId"`
    Category   string `json:"category"`
    Tickets    int    `json:"tickets"`
    Surveys    int    `json:"surveys"`
}

func ToUserDTO(user User) UserDTO {
    return UserDTO{
        ID:     user.ID,
        Name:   user.Name,
        Email:  user.Email,
        Role:   user.Role,
        Entity: user.Entity,
    }
}
