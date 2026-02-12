package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/db"
	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/util"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type seedUser struct {
	Username string
	Password string
	Name     string
	Email    string
	Role     domain.UserRole
	Entity   string
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

func upsertUser(database *gorm.DB, seed seedUser) error {
	email := strings.ToLower(strings.TrimSpace(seed.Email))
	username := strings.ToLower(strings.TrimSpace(seed.Username))
	password := strings.TrimSpace(seed.Password)

	if email == "" || username == "" {
		return fmt.Errorf("username dan email wajib diisi")
	}
	if password == "" {
		return fmt.Errorf("password wajib diisi untuk %s", username)
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return err
	}

	var existing domain.User
	err = database.Where("email = ?", email).First(&existing).Error
	if err == nil {
		updates := map[string]any{
			"username":      username,
			"name":          seed.Name,
			"role":          seed.Role,
			"entity":        seed.Entity,
			"password_hash": hashed,
			"is_active":     true,
		}
		return database.Model(&existing).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	user := domain.User{
		ID:           util.NewID(10),
		Username:     username,
		PasswordHash: hashed,
		Name:         seed.Name,
		Email:        email,
		Role:         seed.Role,
		Entity:       seed.Entity,
		IsActive:     true,
	}

	return database.Create(&user).Error
}

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("ensure database failed: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	seeds := []seedUser{
		// ===== Admin (3) =====
		{
			Username: "admin1",
			Password: "admin123",
			Name:     "Admin 1",
			Email:    "admin1@unila.ac.id",
			Role:     domain.RoleAdmin,
			Entity:   "Admin",
		},
		{
			Username: "admin2",
			Password: "admin123",
			Name:     "Admin 2",
			Email:    "admin2@unila.ac.id",
			Role:     domain.RoleAdmin,
			Entity:   "Admin",
		},
		{
			Username: "admin3",
			Password: "admin123",
			Name:     "Admin 3",
			Email:    "admin3@unila.ac.id",
			Role:     domain.RoleAdmin,
			Entity:   "Admin",
		},

		// ===== Mahasiswa (35) =====
		{Username: "mahasiswa1", Password: "mahasiswa123", Name: "Mahasiswa 1", Email: "mahasiswa1@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa2", Password: "mahasiswa123", Name: "Mahasiswa 2", Email: "mahasiswa2@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa3", Password: "mahasiswa123", Name: "Mahasiswa 3", Email: "mahasiswa3@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa4", Password: "mahasiswa123", Name: "Mahasiswa 4", Email: "mahasiswa4@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa5", Password: "mahasiswa123", Name: "Mahasiswa 5", Email: "mahasiswa5@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa6", Password: "mahasiswa123", Name: "Mahasiswa 6", Email: "mahasiswa6@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa7", Password: "mahasiswa123", Name: "Mahasiswa 7", Email: "mahasiswa7@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa8", Password: "mahasiswa123", Name: "Mahasiswa 8", Email: "mahasiswa8@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa9", Password: "mahasiswa123", Name: "Mahasiswa 9", Email: "mahasiswa9@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa10", Password: "mahasiswa123", Name: "Mahasiswa 10", Email: "mahasiswa10@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa11", Password: "mahasiswa123", Name: "Mahasiswa 11", Email: "mahasiswa11@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa12", Password: "mahasiswa123", Name: "Mahasiswa 12", Email: "mahasiswa12@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa13", Password: "mahasiswa123", Name: "Mahasiswa 13", Email: "mahasiswa13@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa14", Password: "mahasiswa123", Name: "Mahasiswa 14", Email: "mahasiswa14@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa15", Password: "mahasiswa123", Name: "Mahasiswa 15", Email: "mahasiswa15@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa16", Password: "mahasiswa123", Name: "Mahasiswa 16", Email: "mahasiswa16@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa17", Password: "mahasiswa123", Name: "Mahasiswa 17", Email: "mahasiswa17@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa18", Password: "mahasiswa123", Name: "Mahasiswa 18", Email: "mahasiswa18@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa19", Password: "mahasiswa123", Name: "Mahasiswa 19", Email: "mahasiswa19@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa20", Password: "mahasiswa123", Name: "Mahasiswa 20", Email: "mahasiswa20@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa21", Password: "mahasiswa123", Name: "Mahasiswa 21", Email: "mahasiswa21@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa22", Password: "mahasiswa123", Name: "Mahasiswa 22", Email: "mahasiswa22@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa23", Password: "mahasiswa123", Name: "Mahasiswa 23", Email: "mahasiswa23@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa24", Password: "mahasiswa123", Name: "Mahasiswa 24", Email: "mahasiswa24@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa25", Password: "mahasiswa123", Name: "Mahasiswa 25", Email: "mahasiswa25@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa26", Password: "mahasiswa123", Name: "Mahasiswa 26", Email: "mahasiswa26@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa27", Password: "mahasiswa123", Name: "Mahasiswa 27", Email: "mahasiswa27@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa28", Password: "mahasiswa123", Name: "Mahasiswa 28", Email: "mahasiswa28@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa29", Password: "mahasiswa123", Name: "Mahasiswa 29", Email: "mahasiswa29@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa30", Password: "mahasiswa123", Name: "Mahasiswa 30", Email: "mahasiswa30@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa31", Password: "mahasiswa123", Name: "Mahasiswa 31", Email: "mahasiswa31@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa32", Password: "mahasiswa123", Name: "Mahasiswa 32", Email: "mahasiswa32@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa33", Password: "mahasiswa123", Name: "Mahasiswa 33", Email: "mahasiswa33@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa34", Password: "mahasiswa123", Name: "Mahasiswa 34", Email: "mahasiswa34@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa35", Password: "mahasiswa123", Name: "Mahasiswa 35", Email: "mahasiswa35@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},

		// ===== Dosen (7) =====
		{Username: "dosen1", Password: "dosen123", Name: "Dosen 1", Email: "dosen1@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen2", Password: "dosen123", Name: "Dosen 2", Email: "dosen2@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen3", Password: "dosen123", Name: "Dosen 3", Email: "dosen3@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen4", Password: "dosen123", Name: "Dosen 4", Email: "dosen4@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen5", Password: "dosen123", Name: "Dosen 5", Email: "dosen5@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen6", Password: "dosen123", Name: "Dosen 6", Email: "dosen6@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen7", Password: "dosen123", Name: "Dosen 7", Email: "dosen7@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},

		// ===== Tendik (5) =====
		{Username: "tendik1", Password: "tendik123", Name: "Tendik 1", Email: "tendik1@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik2", Password: "tendik123", Name: "Tendik 2", Email: "tendik2@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik3", Password: "tendik123", Name: "Tendik 3", Email: "tendik3@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik4", Password: "tendik123", Name: "Tendik 4", Email: "tendik4@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik5", Password: "tendik123", Name: "Tendik 5", Email: "tendik5@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
	}

	for _, seed := range seeds {
		if err := upsertUser(database, seed); err != nil {
			log.Fatalf("seed user %s gagal: %v", seed.Username, err)
		}
	}

	log.Println("seed users selesai")
}
