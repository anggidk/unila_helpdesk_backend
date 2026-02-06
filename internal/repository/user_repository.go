package repository

import (
	"strings"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

// UserRepositoryInterface defines the contract for user data access
type UserRepositoryInterface interface {
	Create(user *domain.User) error
	UpsertByEmail(user *domain.User) error
	FindByID(id string) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindByUsername(username string) (*domain.User, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Create(user *domain.User) error {
	return repo.db.Create(user).Error
}

func (repo *UserRepository) UpsertByEmail(user *domain.User) error {
	var existing domain.User
	err := repo.db.Where("email = ?", strings.ToLower(user.Email)).First(&existing).Error

	if err == nil {
		// User ada, update
		user.ID = existing.ID
		return repo.db.Model(&existing).Updates(map[string]any{
			"name":     user.Name,
			"role":     user.Role,
			"entity":   user.Entity,
			"username": user.Username,
		}).Error
	} else if err != gorm.ErrRecordNotFound {
		// Error tak terduga
		return err
	}

	// User tidak ada, buat baru
	return repo.db.Create(user).Error
}

func (repo *UserRepository) FindByID(id string) (*domain.User, error) {
	var user domain.User
	if err := repo.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := repo.db.Where("email = ?", strings.ToLower(email)).First(&user).Error; err != nil {
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
