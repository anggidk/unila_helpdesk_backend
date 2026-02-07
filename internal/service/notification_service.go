package service

import (
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
	"unila_helpdesk_backend/internal/util"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
	tokens        *repository.FCMTokenRepository
	now           func() time.Time
}

type FCMRegisterRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type FCMUnregisterRequest struct {
	Token string `json:"token"`
}

func NewNotificationService(notifications *repository.NotificationRepository, tokens *repository.FCMTokenRepository) *NotificationService {
	return &NotificationService{
		notifications: notifications,
		tokens:        tokens,
		now:           time.Now,
	}
}

func (service *NotificationService) List(user domain.User) ([]domain.NotificationDTO, error) {
	items, err := service.notifications.ListByUser(user.ID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.NotificationDTO, 0, len(items))
	for _, item := range items {
		result = append(result, domain.NotificationDTO{
			ID:        item.ID,
			TicketID:  item.TicketID,
			Title:     item.Title,
			Message:   item.Message,
			Timestamp: item.CreatedAt,
			IsRead:    item.IsRead,
		})
	}
	return result, nil
}

func (service *NotificationService) RegisterToken(user domain.User, req FCMRegisterRequest) error {
	if user.Role == domain.RoleGuest {
		return nil
	}
	tokenValue := strings.TrimSpace(req.Token)
	if tokenValue == "" {
		return nil
	}
	token := domain.FCMToken{
		ID:        util.NewUUID(),
		UserID:    user.ID,
		Token:     tokenValue,
		Platform:  req.Platform,
		CreatedAt: service.now(),
		UpdatedAt: service.now(),
	}
	return service.tokens.Upsert(&token)
}

func (service *NotificationService) UnregisterToken(user domain.User, req FCMUnregisterRequest) error {
	tokenValue := strings.TrimSpace(req.Token)
	if tokenValue == "" {
		return nil
	}
	return service.tokens.DeleteByUserAndToken(user.ID, tokenValue)
}
