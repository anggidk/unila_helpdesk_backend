package db

import (
    "fmt"
    "log"
    "net/url"
    "strings"
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

func EnsureDatabase(cfg config.Config) error {
    parsed, err := url.Parse(cfg.DatabaseURL)
    if err != nil {
        return fmt.Errorf("invalid DATABASE_URL: %w", err)
    }
    dbName := strings.TrimPrefix(parsed.Path, "/")
    if dbName == "" {
        return fmt.Errorf("DATABASE_URL missing database name")
    }

    adminURL := *parsed
    adminURL.Path = "/postgres"
    adminDB, err := gorm.Open(postgres.Open(adminURL.String()), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect admin db: %w", err)
    }
    sqlDB, err := adminDB.DB()
    if err == nil {
        defer sqlDB.Close()
    }

    var exists bool
    if err := adminDB.Raw(
        "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)",
        dbName,
    ).Scan(&exists).Error; err != nil {
        return fmt.Errorf("failed to check database: %w", err)
    }
    if exists {
        return nil
    }

    if err := adminDB.Exec(
        fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName)),
    ).Error; err != nil {
        return fmt.Errorf("failed to create database: %w", err)
    }
    return nil
}

func quoteIdentifier(value string) string {
    return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func AutoMigrate(database *gorm.DB) error {
    // Normalize legacy schema to avoid type-mismatch errors before AutoMigrate.
    _ = database.Exec(`
        DO $$
        BEGIN
            IF to_regclass('public.tickets') IS NOT NULL THEN
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_reporter') THEN
                    ALTER TABLE tickets DROP CONSTRAINT fk_tickets_reporter;
                END IF;
                ALTER TABLE tickets ALTER COLUMN id TYPE varchar(64) USING id::varchar(64);
                ALTER TABLE tickets ALTER COLUMN reporter_id TYPE varchar(36) USING reporter_id::varchar(36);
            END IF;
            IF to_regclass('public.ticket_histories') IS NOT NULL THEN
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_history') THEN
                    ALTER TABLE ticket_histories DROP CONSTRAINT fk_tickets_history;
                END IF;
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_ticket_histories_ticket') THEN
                    ALTER TABLE ticket_histories DROP CONSTRAINT fk_ticket_histories_ticket;
                END IF;
                ALTER TABLE ticket_histories ALTER COLUMN ticket_id TYPE varchar(64) USING ticket_id::varchar(64);
            END IF;
            IF to_regclass('public.ticket_comments') IS NOT NULL THEN
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_ticket_comments_ticket') THEN
                    ALTER TABLE ticket_comments DROP CONSTRAINT fk_ticket_comments_ticket;
                END IF;
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_comments') THEN
                    ALTER TABLE ticket_comments DROP CONSTRAINT fk_tickets_comments;
                END IF;
                ALTER TABLE ticket_comments ALTER COLUMN ticket_id TYPE varchar(64) USING ticket_id::varchar(64);
            END IF;
            IF to_regclass('public.survey_responses') IS NOT NULL THEN
                IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_responses_ticket') THEN
                    ALTER TABLE survey_responses DROP CONSTRAINT fk_survey_responses_ticket;
                END IF;
                ALTER TABLE survey_responses ALTER COLUMN ticket_id TYPE varchar(64) USING ticket_id::varchar(64);
            END IF;
        END $$;
    `).Error
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
