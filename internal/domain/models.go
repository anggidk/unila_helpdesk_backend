package domain

import (
    "time"

    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type UserRole string

type TicketPriority string

type TicketStatus string

type SurveyQuestionType string

const (
    RoleRegistered UserRole = "registered"
    RoleGuest      UserRole = "guest"
    RoleAdmin      UserRole = "admin"
)

const (
    PriorityLow    TicketPriority = "low"
    PriorityMedium TicketPriority = "medium"
    PriorityHigh   TicketPriority = "high"
)

const (
    StatusWaiting    TicketStatus = "waiting"
    StatusProcessing TicketStatus = "processing"
    StatusInProgress TicketStatus = "inProgress"
    StatusResolved   TicketStatus = "resolved"
)

const (
    QuestionLikert         SurveyQuestionType = "likert"
    QuestionYesNo          SurveyQuestionType = "yesNo"
    QuestionMultipleChoice SurveyQuestionType = "multipleChoice"
    QuestionText           SurveyQuestionType = "text"
)

type User struct {
    ID           string         `gorm:"primaryKey;type:varchar(36)"`
    Username     string         `gorm:"size:60;uniqueIndex"`
    PasswordHash string         `gorm:"type:text"`
    Name         string         `gorm:"size:120"`
    Email        string         `gorm:"size:180;uniqueIndex"`
    Role         UserRole       `gorm:"size:20"`
    Entity       string         `gorm:"size:120"`
    IsActive     bool           `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type ServiceCategory struct {
    ID           string `gorm:"primaryKey;size:60"`
    Name         string `gorm:"size:120"`
    GuestAllowed bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Ticket struct {
    ID             string         `gorm:"primaryKey;size:64"`
    Title          string         `gorm:"size:180"`
    Description    string         `gorm:"type:text"`
    CategoryID     string         `gorm:"size:60;index"`
    Priority       TicketPriority `gorm:"size:20"`
    Status         TicketStatus   `gorm:"size:20"`
    ReporterID     string         `gorm:"size:36;index"`
    ReporterName   string         `gorm:"size:120"`
    IsGuest        bool           `gorm:"default:false"`
    Assignee       string         `gorm:"size:120"`
    SurveyRequired bool           `gorm:"default:false"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt `gorm:"index"`

    Category ServiceCategory `gorm:"foreignKey:CategoryID"`
    History  []TicketHistory `gorm:"foreignKey:TicketID"`
    Comments []TicketComment `gorm:"foreignKey:TicketID"`
}

type TicketHistory struct {
    ID          string    `gorm:"primaryKey;type:varchar(36)"`
    TicketID    string    `gorm:"size:64;index"`
    Title       string    `gorm:"size:160"`
    Description string    `gorm:"type:text"`
    Timestamp   time.Time `gorm:"index"`
    CreatedAt   time.Time
}

type TicketComment struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)"`
    TicketID  string    `gorm:"size:64;index"`
    Author    string    `gorm:"size:120"`
    Message   string    `gorm:"type:text"`
    IsStaff   bool
    Timestamp time.Time `gorm:"index"`
    CreatedAt time.Time
}

type SurveyTemplate struct {
    ID          string           `gorm:"primaryKey;size:64"`
    Title       string           `gorm:"size:160"`
    Description string           `gorm:"type:text"`
    CategoryID  string           `gorm:"size:60;index"`
    Questions   []SurveyQuestion `gorm:"foreignKey:TemplateID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type SurveyQuestion struct {
    ID         string             `gorm:"primaryKey;size:64"`
    TemplateID string             `gorm:"size:64;index"`
    Text       string             `gorm:"type:text"`
    Type       SurveyQuestionType `gorm:"size:24"`
    Options    datatypes.JSON     `gorm:"type:jsonb"`
    CreatedAt  time.Time
}

type SurveyResponse struct {
    ID        string         `gorm:"primaryKey;type:varchar(36)"`
    TicketID  string         `gorm:"size:64;index"`
    UserID    string         `gorm:"size:36;index"`
    Answers   datatypes.JSON `gorm:"type:jsonb"`
    Score     float64        `gorm:"default:0"`
    CreatedAt time.Time
}

type Notification struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)"`
    UserID    string    `gorm:"size:36;index"`
    Title     string    `gorm:"size:160"`
    Message   string    `gorm:"type:text"`
    IsRead    bool      `gorm:"default:false"`
    CreatedAt time.Time
}

type FCMToken struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)"`
    UserID    string    `gorm:"size:36;index"`
    Token     string    `gorm:"type:text"`
    Platform  string    `gorm:"size:40"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
