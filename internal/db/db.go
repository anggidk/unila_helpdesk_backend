package db

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/config"

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
	return database.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(migration20260209Baseline).Error
	})
}

func MustAutoMigrate(database *gorm.DB) {
	if err := AutoMigrate(database); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}
}

const migration20260209Baseline = `
CREATE TABLE IF NOT EXISTS public.users (
    id varchar(10) PRIMARY KEY,
    username varchar(60) UNIQUE,
    password_hash text,
    name varchar(120),
    email varchar(180) UNIQUE,
    role varchar(20),
    entity varchar(120),
    is_active boolean DEFAULT true,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.service_categories (
    id varchar(6) PRIMARY KEY,
    name varchar(120),
    guest_allowed boolean
);

CREATE TABLE IF NOT EXISTS public.tickets (
    id varchar(32) PRIMARY KEY,
    user_id varchar(10),
    category_id varchar(6),
    ticket_number varchar(20),
    reporter_name varchar(120),
    email varchar(180),
    phone varchar(20),
    is_guest boolean DEFAULT false,
    title varchar(180),
    description text,
    priority varchar(20),
    status varchar(20),
    staff_notes text,
    survey_required boolean DEFAULT false,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.ticket_histories (
    id varchar(64) PRIMARY KEY,
    ticket_id varchar(32),
    title varchar(120),
    description text,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.attachments (
    id varchar(32) PRIMARY KEY,
    ticket_id varchar(32),
    filename varchar(180),
    content_type varchar(80),
    size bigint,
    data bytea,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.survey_templates (
    id varchar(12) PRIMARY KEY,
    title varchar(160),
    description text,
    framework varchar(80),
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.category_templates (
    category_id varchar(6) PRIMARY KEY,
    template_id varchar(12) NOT NULL,
    assigned_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.survey_questions (
    id varchar(32) PRIMARY KEY,
    template_id varchar(12),
    text text,
    type varchar(24),
    options jsonb,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.survey_responses (
    id varchar(32) PRIMARY KEY,
    user_id varchar(10),
    ticket_id varchar(32),
    template_id varchar(12),
    score numeric,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.survey_response_items (
    id varchar(32) PRIMARY KEY,
    response_id varchar(32),
    question_id varchar(32),
    answer_value jsonb,
    score_value numeric,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.notifications (
    id varchar(64) PRIMARY KEY,
    user_id varchar(10),
    ticket_id varchar(32),
    title varchar(160),
    message text,
    is_read boolean DEFAULT false,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.fcm_tokens (
    id varchar(64) PRIMARY KEY,
    user_id varchar(10),
    token text,
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id varchar(64) PRIMARY KEY,
    user_id varchar(10),
    token_hash varchar(64),
    expires_at timestamptz,
    created_at timestamptz
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_attachments_ticket_id') THEN
        ALTER TABLE public.attachments
            ADD CONSTRAINT fk_attachments_ticket_id
            FOREIGN KEY (ticket_id) REFERENCES public.tickets(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_fcm_tokens_user_id') THEN
        ALTER TABLE public.fcm_tokens
            ADD CONSTRAINT fk_fcm_tokens_user_id
            FOREIGN KEY (user_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_user_id') THEN
        ALTER TABLE public.notifications
            ADD CONSTRAINT fk_notifications_user_id
            FOREIGN KEY (user_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_ticket_id') THEN
        ALTER TABLE public.notifications
            ADD CONSTRAINT fk_notifications_ticket_id
            FOREIGN KEY (ticket_id) REFERENCES public.tickets(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_refresh_tokens_user_id') THEN
        ALTER TABLE public.refresh_tokens
            ADD CONSTRAINT fk_refresh_tokens_user_id
            FOREIGN KEY (user_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_questions_template_id') THEN
        ALTER TABLE public.survey_questions
            ADD CONSTRAINT fk_survey_questions_template_id
            FOREIGN KEY (template_id) REFERENCES public.survey_templates(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_responses_ticket_id') THEN
        ALTER TABLE public.survey_responses
            ADD CONSTRAINT fk_survey_responses_ticket_id
            FOREIGN KEY (ticket_id) REFERENCES public.tickets(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_responses_user_id') THEN
        ALTER TABLE public.survey_responses
            ADD CONSTRAINT fk_survey_responses_user_id
            FOREIGN KEY (user_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_responses_template_id') THEN
        ALTER TABLE public.survey_responses
            ADD CONSTRAINT fk_survey_responses_template_id
            FOREIGN KEY (template_id) REFERENCES public.survey_templates(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_response_items_response_id') THEN
        ALTER TABLE public.survey_response_items
            ADD CONSTRAINT fk_survey_response_items_response_id
            FOREIGN KEY (response_id) REFERENCES public.survey_responses(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_response_items_question_id') THEN
        ALTER TABLE public.survey_response_items
            ADD CONSTRAINT fk_survey_response_items_question_id
            FOREIGN KEY (question_id) REFERENCES public.survey_questions(id)
            ON UPDATE NO ACTION ON DELETE NO ACTION;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_ticket_histories_ticket_id') THEN
        ALTER TABLE public.ticket_histories
            ADD CONSTRAINT fk_ticket_histories_ticket_id
            FOREIGN KEY (ticket_id) REFERENCES public.tickets(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_user_id') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT fk_tickets_user_id
            FOREIGN KEY (user_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_category_id') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT fk_tickets_category_id
            FOREIGN KEY (category_id) REFERENCES public.service_categories(id)
            ON UPDATE NO ACTION ON DELETE NO ACTION;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_category_templates_category_id') THEN
        ALTER TABLE public.category_templates
            ADD CONSTRAINT fk_category_templates_category_id
            FOREIGN KEY (category_id) REFERENCES public.service_categories(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_category_templates_template_id') THEN
        ALTER TABLE public.category_templates
            ADD CONSTRAINT fk_category_templates_template_id
            FOREIGN KEY (template_id) REFERENCES public.survey_templates(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_guest_identity') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_guest_identity
            CHECK (
                is_guest = false OR
                (user_id IS NULL AND email IS NOT NULL AND phone IS NOT NULL)
            );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_registered_user') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_registered_user
            CHECK (
                is_guest = true OR user_id IS NOT NULL
            );
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_tickets_ticket_number ON public.tickets(ticket_number);
CREATE INDEX IF NOT EXISTS ix_tickets_user_id ON public.tickets(user_id);
CREATE INDEX IF NOT EXISTS ix_tickets_category_id ON public.tickets(category_id);
CREATE INDEX IF NOT EXISTS ix_tickets_status ON public.tickets(status);
CREATE INDEX IF NOT EXISTS ix_tickets_created_at ON public.tickets(created_at);
CREATE INDEX IF NOT EXISTS ix_attachments_ticket_id ON public.attachments(ticket_id);
CREATE INDEX IF NOT EXISTS ix_ticket_histories_ticket_id_created_at ON public.ticket_histories(ticket_id, created_at);
CREATE INDEX IF NOT EXISTS ix_survey_responses_ticket_id ON public.survey_responses(ticket_id);
CREATE INDEX IF NOT EXISTS ix_survey_responses_user_id ON public.survey_responses(user_id);
CREATE INDEX IF NOT EXISTS ix_survey_responses_template_id ON public.survey_responses(template_id);
CREATE INDEX IF NOT EXISTS ix_survey_response_items_response_id ON public.survey_response_items(response_id);
CREATE INDEX IF NOT EXISTS ix_survey_response_items_question_id ON public.survey_response_items(question_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_survey_response_items_response_question ON public.survey_response_items(response_id, question_id);
CREATE INDEX IF NOT EXISTS ix_category_templates_template_id ON public.category_templates(template_id);
CREATE INDEX IF NOT EXISTS ix_refresh_tokens_user_id ON public.refresh_tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_refresh_tokens_token_hash ON public.refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS ix_refresh_tokens_expires_at ON public.refresh_tokens(expires_at);
`
