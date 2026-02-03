package repository

import (
    "unila_helpdesk_backend/internal/domain"

    "gorm.io/gorm"
)

type RefreshTokenRepository struct {
    db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
    return &RefreshTokenRepository{db: db}
}

func (repo *RefreshTokenRepository) Create(token *domain.RefreshToken) error {
    return repo.db.Create(token).Error
}

func (repo *RefreshTokenRepository) FindByHash(hash string) (*domain.RefreshToken, error) {
    var token domain.RefreshToken
    if err := repo.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
        return nil, err
    }
    return &token, nil
}

func (repo *RefreshTokenRepository) DeleteByID(id string) error {
    return repo.db.Delete(&domain.RefreshToken{}, "id = ?", id).Error
}
