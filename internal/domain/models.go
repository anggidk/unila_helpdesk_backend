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
	StatusInProgress TicketStatus = "inProgress"
	StatusResolved   TicketStatus = "resolved"
)

const (
	QuestionLikert         SurveyQuestionType = "likert"
	QuestionLikertQuality  SurveyQuestionType = "likertQuality"
	QuestionLikert3Puas    SurveyQuestionType = "likert3Puas"
	QuestionLikert3        SurveyQuestionType = "likert3"
	QuestionLikert4Puas    SurveyQuestionType = "likert4Puas"
	QuestionLikert4        SurveyQuestionType = "likert4"
	QuestionYesNo          SurveyQuestionType = "yesNo"
	QuestionMultipleChoice SurveyQuestionType = "multipleChoice"
	QuestionText           SurveyQuestionType = "text"
)

type User struct {
	ID           string   `gorm:"primaryKey;size:10"`
	Username     string   `gorm:"size:60;uniqueIndex"`
	PasswordHash string   `gorm:"type:text"`
	Name         string   `gorm:"size:120"`
	Email        string   `gorm:"size:180;uniqueIndex"`
	Role         UserRole `gorm:"size:20"`
	Entity       string   `gorm:"size:120"`
	IsActive     bool     `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type ServiceCategory struct {
	ID               string `gorm:"primaryKey;size:6"`
	Name             string `gorm:"size:120"`
	GuestAllowed     bool
	SurveyTemplateID string `gorm:"size:12"`
}

type Ticket struct {
	ID             string         `gorm:"primaryKey;size:32"`
	TicketNumber   string         `gorm:"size:20;uniqueIndex"`
	UserID         string         `gorm:"size:10;index"`
	ReporterName   string         `gorm:"size:120"`
	Email          string         `gorm:"size:180"`
	Phone          *string        `gorm:"size:20"`
	IsGuest        bool           `gorm:"default:false"`
	Title          string         `gorm:"size:180"`
	Description    string         `gorm:"type:text"`
	CategoryID     string         `gorm:"size:6;index"`
	Priority       TicketPriority `gorm:"size:20"`
	Status         TicketStatus   `gorm:"size:20"`
	StaffNotes     string         `gorm:"type:text"`
	SurveyRequired bool           `gorm:"default:false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Category ServiceCategory `gorm:"foreignKey:CategoryID"`
	History  []TicketHistory `gorm:"foreignKey:TicketID"`
}

type TicketHistory struct {
	ID          string `gorm:"primaryKey;size:64"`
	TicketID    string `gorm:"size:32;index"`
	Title       string `gorm:"size:120"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
}

type Attachment struct {
	ID          string `gorm:"primaryKey;size:32"`
	TicketID    string `gorm:"size:32;index"`
	Filename    string `gorm:"size:180"`
	ContentType string `gorm:"size:80"`
	Size        int64
	Data        []byte `gorm:"type:bytea"`
	CreatedAt   time.Time
}

type SurveyTemplate struct {
	ID          string           `gorm:"primaryKey;size:12"`
	Title       string           `gorm:"size:160"`
	Description string           `gorm:"type:text"`
	Framework   string           `gorm:"size:80"`
	CategoryID  string           `gorm:"size:6;index"`
	Questions   []SurveyQuestion `gorm:"foreignKey:TemplateID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SurveyQuestion struct {
	ID         string             `gorm:"primaryKey;size:32"`
	TemplateID string             `gorm:"size:12;index"`
	Text       string             `gorm:"type:text"`
	Type       SurveyQuestionType `gorm:"size:24"`
	Options    datatypes.JSON     `gorm:"type:jsonb"`
	CreatedAt  time.Time
}

type SurveyResponse struct {
	ID         string         `gorm:"primaryKey;size:32"`
	TicketID   string         `gorm:"size:32;index"`
	UserID     string         `gorm:"size:10;index"`
	TemplateID string         `gorm:"size:12;index"`
	Answers    datatypes.JSON `gorm:"type:jsonb"`
	Score      float64        `gorm:"default:0"`
	CreatedAt  time.Time
}

type Notification struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"size:10;index"`
	TicketID  string `gorm:"size:32;index"`
	Title     string `gorm:"size:160"`
	Message   string `gorm:"type:text"`
	IsRead    bool   `gorm:"default:false"`
	CreatedAt time.Time
}

type FCMToken struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"size:10;index"`
	Token     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    string    `gorm:"size:10;index"`
	TokenHash string    `gorm:"size:64;uniqueIndex"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}
