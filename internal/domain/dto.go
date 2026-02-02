package domain

import "time"

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
}

type ServiceCategoryDTO struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    GuestAllowed bool   `json:"guestAllowed"`
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
    CategoryID  string              `json:"categoryId"`
    Questions   []SurveyQuestionDTO `json:"questions"`
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

func ToUserDTO(user User) UserDTO {
    return UserDTO{
        ID:     user.ID,
        Name:   user.Name,
        Email:  user.Email,
        Role:   user.Role,
        Entity: user.Entity,
    }
}
