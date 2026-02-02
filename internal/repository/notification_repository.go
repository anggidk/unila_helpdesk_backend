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

func (repo *NotificationRepository) MarkRead(notificationID string, userID string) error {
    return repo.db.Model(&domain.Notification{}).Where("id = ? AND user_id = ?", notificationID, userID).Update("is_read", true).Error
}

type FCMTokenRepository struct {
    db *gorm.DB
}

func NewFCMTokenRepository(db *gorm.DB) *FCMTokenRepository {
    return &FCMTokenRepository{db: db}
}

func (repo *FCMTokenRepository) Upsert(token *domain.FCMToken) error {
    var existing domain.FCMToken
    if err := repo.db.Where("user_id = ? AND token = ?", token.UserID, token.Token).First(&existing).Error; err == nil {
        return repo.db.Model(&existing).Updates(map[string]any{
            "platform": token.Platform,
        }).Error
    }
    return repo.db.Create(token).Error
}

func (repo *FCMTokenRepository) ListTokens(userID string) ([]domain.FCMToken, error) {
    var tokens []domain.FCMToken
    if err := repo.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
        return nil, err
    }
    return tokens, nil
}
