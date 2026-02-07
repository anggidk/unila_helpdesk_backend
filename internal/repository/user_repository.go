package repository

import (
	"strings"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) FindByID(id string) (*domain.User, error) {
	var user domain.User
	if err := repo.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := repo.db.Where("username = ?", strings.ToLower(username)).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
