package repository

import (
	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (repo *NotificationRepository) ListByUser(userID string) ([]domain.Notification, error) {
	var notifications []domain.Notification
	if err := repo.db.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (repo *NotificationRepository) Create(notification *domain.Notification) error {
	return repo.db.Create(notification).Error
}

type FCMTokenRepository struct {
	db *gorm.DB
}

func NewFCMTokenRepository(db *gorm.DB) *FCMTokenRepository {
	return &FCMTokenRepository{db: db}
}

func (repo *FCMTokenRepository) Upsert(token *domain.FCMToken) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("token = ?", token.Token).Delete(&domain.FCMToken{}).Error; err != nil {
			return err
		}
		return tx.Create(token).Error
	})
}

func (repo *FCMTokenRepository) ListTokens(userID string) ([]domain.FCMToken, error) {
	var tokens []domain.FCMToken
	if err := repo.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (repo *FCMTokenRepository) DeleteByUserAndTokens(userID string, tokens []string) error {
	if userID == "" || len(tokens) == 0 {
		return nil
	}
	return repo.db.Where("user_id = ? AND token IN ?", userID, tokens).Delete(&domain.FCMToken{}).Error
}

func (repo *FCMTokenRepository) DeleteByUserAndToken(userID string, token string) error {
	if userID == "" || token == "" {
		return nil
	}
	return repo.db.Where("user_id = ? AND token = ?", userID, token).Delete(&domain.FCMToken{}).Error
}
