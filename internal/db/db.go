package db

import (
    "log"
    "time"

    "unila_helpdesk_backend/internal/config"
    "unila_helpdesk_backend/internal/domain"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
    database, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    sqlDB, err := database.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxOpenConns(cfg.DatabaseMaxConns)
    sqlDB.SetMaxIdleConns(cfg.DatabaseIdleConns)
    sqlDB.SetConnMaxLifetime(30 * time.Minute)

    return database, nil
}

func AutoMigrate(database *gorm.DB) error {
    return database.AutoMigrate(
        &domain.User{},
        &domain.ServiceCategory{},
        &domain.Ticket{},
        &domain.TicketHistory{},
        &domain.TicketComment{},
        &domain.SurveyTemplate{},
        &domain.SurveyQuestion{},
        &domain.SurveyResponse{},
        &domain.Notification{},
        &domain.FCMToken{},
    )
}

func MustAutoMigrate(database *gorm.DB) {
    if err := AutoMigrate(database); err != nil {
        log.Fatalf("auto migrate failed: %v", err)
    }
}
