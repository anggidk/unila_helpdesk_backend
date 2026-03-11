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
	PriorityLow    TicketPriority = "LOW"
	PriorityMedium TicketPriority = "MEDIUM"
	PriorityHigh   TicketPriority = "HIGH"
)

const (
	StatusWaiting TicketStatus = "WAITING"
	StatusAssign  TicketStatus = "ASSIGN"
	StatusDone    TicketStatus = "DONE"
	StatusReject  TicketStatus = "REJECT"
)

const (
	EntityDosen     = "DOSEN"
	EntityTendik    = "TENDIK"
	EntityMahasiswa = "MAHASISWA"
	EntityLainnya   = "LAINNYA"
)

const (
	ServiceGuestPassword     = 1
	ServiceGuestRegistration = 2
	ServiceGuestEmail        = 3
	ServiceInternet          = 4
	ServiceWebsiteDown       = 5
	ServiceSistemInformasi   = 6
	ServiceSIAKADU           = 7
	ServiceLainnya           = 99
)

const (
	QuestionLikert3        SurveyQuestionType = "likert3"
	QuestionLikert4        SurveyQuestionType = "likert4"
	QuestionLikert5        SurveyQuestionType = "likert5"
	QuestionYesNo          SurveyQuestionType = "yesNo"
	QuestionMultipleChoice SurveyQuestionType = "multipleChoice"
	QuestionText           SurveyQuestionType = "text"
)

type User struct {
	ID           string   `gorm:"primaryKey;size:25"`
	Username     string   `gorm:"size:64;uniqueIndex"`
	PasswordHash string   `gorm:"type:text"`
	Name         string   `gorm:"size:100"`
	Email        string   `gorm:"size:255;uniqueIndex"`
	Role         UserRole `gorm:"size:20"`
	Entity       string   `gorm:"size:20"`
	IsActive     bool     `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type Staff struct {
	ID       string `gorm:"primaryKey;size:25"`
	Username string `gorm:"size:100;uniqueIndex"`
	Name     string `gorm:"size:100"`
	NIP      string `gorm:"size:25"`
	Divisi   string `gorm:"size:100"`
	Role     string `gorm:"size:50"`
	Photo    string `gorm:"size:255"`
	HP       string `gorm:"size:25"`
}

type ServiceCategory struct {
	ID               int    `gorm:"primaryKey;column:id"`
	Name             string `gorm:"size:120"`
	GuestAllowed     bool   `gorm:"-"`
	SurveyTemplateID string `gorm:"column:survey_template_id;->"`
}

func (ServiceCategory) TableName() string {
	return "services"
}

type CategoryTemplate struct {
	CategoryID int    `gorm:"primaryKey;column:category_id"`
	TemplateID string `gorm:"size:12;index"`
	AssignedAt time.Time
}

type Ticket struct {
	ID             int            `gorm:"primaryKey;autoIncrement"`
	TicketNumber   string         `gorm:"size:6;uniqueIndex"`
	Username       *string        `gorm:"size:64"`
	NumberID       *string        `gorm:"column:number_id;size:25"`
	Name           string         `gorm:"size:100"`
	Email          string         `gorm:"size:255"`
	Entity         string         `gorm:"size:20"`
	ServiceID      int            `gorm:"column:id_service;index"`
	Notes          string         `gorm:"column:notes;type:text"`
	StaffNotes     string         `gorm:"type:text"`
	Priority       TicketPriority `gorm:"size:20"`
	IsReject       bool           `gorm:"column:is_reject;default:false"`
	IsAssign       bool           `gorm:"column:is_assign;default:false"`
	IsDone         bool           `gorm:"column:is_done;default:false"`
	StaffID        *string        `gorm:"column:id_staff;size:25"`
	Status         TicketStatus   `gorm:"size:50"`
	Lamp1          string         `gorm:"type:text"`
	Lamp2          string         `gorm:"type:text"`
	SurveyRequired bool           `gorm:"default:false"`
	CreatedAt      time.Time      `gorm:"column:ticket_date"`

	Service ServiceCategory `gorm:"foreignKey:ServiceID"`
}

type SurveyTemplate struct {
	ID          string           `gorm:"primaryKey;size:12"`
	Title       string           `gorm:"size:160"`
	Description string           `gorm:"type:text"`
	Framework   string           `gorm:"size:80"`
	IsActive    bool             `gorm:"column:is_active;default:true"`
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
	ID         string  `gorm:"primaryKey;size:32"`
	TicketID   int     `gorm:"index"`
	UserID     string  `gorm:"column:number_id;size:25;index"`
	TemplateID string  `gorm:"size:12;index"`
	Score      float64 `gorm:"default:0"`
	CreatedAt  time.Time
	Items      []SurveyResponseItem `gorm:"foreignKey:ResponseID"`
}

type SurveyResponseItem struct {
	ID          string         `gorm:"primaryKey;size:32"`
	ResponseID  string         `gorm:"size:32;index"`
	QuestionID  string         `gorm:"size:32;index"`
	AnswerValue datatypes.JSON `gorm:"type:jsonb"`
	ScoreValue  *float64
	CreatedAt   time.Time
}

type Notification struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"column:number_id;size:25;index"`
	TicketID  int    `gorm:"index"`
	Title     string `gorm:"size:160"`
	Message   string `gorm:"type:text"`
	IsRead    bool   `gorm:"default:false"`
	CreatedAt time.Time
}

type FCMToken struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"column:number_id;size:25;index"`
	Token     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    string    `gorm:"column:number_id;size:25;index"`
	TokenHash string    `gorm:"size:64;uniqueIndex"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}
