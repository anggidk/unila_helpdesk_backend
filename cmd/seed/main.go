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
        ID:           util.NewUUID(),
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
    cfg := config.Load()

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
        {
            Username: "admin",
            Password: "Admin123!",
            Name:     "Admin Helpdesk",
            Email:    "admin@helpdesk.local",
            Role:     domain.RoleAdmin,
            Entity:   "Admin",
        },
        {
            Username: "mhs01",
            Password: "Mhs123!",
            Name:     "Mahasiswa 01",
            Email:    "mhs01@unila.local",
            Role:     domain.RoleRegistered,
            Entity:   "Mahasiswa",
        },
        {
            Username: "dsn01",
            Password: "Dsn123!",
            Name:     "Dosen 01",
            Email:    "dsn01@unila.local",
            Role:     domain.RoleRegistered,
            Entity:   "Dosen",
        },
        {
            Username: "tdk01",
            Password: "Tdk123!",
            Name:     "Tendik 01",
            Email:    "tdk01@unila.local",
            Role:     domain.RoleRegistered,
            Entity:   "Tendik",
        },
    }

    for _, seed := range seeds {
        if err := upsertUser(database, seed); err != nil {
            log.Fatalf("seed user %s gagal: %v", seed.Username, err)
        }
    }

    log.Println("seed users selesai")
}
