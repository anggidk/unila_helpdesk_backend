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
		if err := tx.Exec(migration20260209Baseline).Error; err != nil {
			return err
		}
		return tx.Exec(migration20260311LampToText).Error
	})
}

func MustAutoMigrate(database *gorm.DB) {
	if err := AutoMigrate(database); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}
}

const migration20260209Baseline = `
CREATE TABLE IF NOT EXISTS public.users (
    id varchar(25) PRIMARY KEY,
    username varchar(64) UNIQUE,
    password_hash text,
    name varchar(100),
    email varchar(255) UNIQUE,
    role varchar(20),
    entity varchar(20),
    is_active boolean DEFAULT true,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.services (
    id int PRIMARY KEY,
    name varchar(120) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.staff (
    id varchar(25) PRIMARY KEY,
    username varchar(100) UNIQUE,
    name varchar(100),
    nip varchar(25),
    divisi varchar(100),
    role varchar(50),
    photo varchar(255),
    hp varchar(25)
);

CREATE TABLE IF NOT EXISTS public.tickets (
    id bigserial PRIMARY KEY,
    ticket_number varchar(6) NOT NULL,
    ticket_date timestamptz NOT NULL,
    username varchar(64),
    number_id varchar(25) NOT NULL,
    name varchar(100) NOT NULL,
    email varchar(255) NOT NULL DEFAULT '',
    entity varchar(20) NOT NULL,
    id_service int NOT NULL,
    notes text NOT NULL,
    staff_notes text NOT NULL DEFAULT '',
    priority varchar(20) NOT NULL,
    is_reject boolean NOT NULL DEFAULT false,
    is_assign boolean NOT NULL DEFAULT false,
    is_done boolean NOT NULL DEFAULT false,
    id_staff varchar(25),
    status varchar(50) NOT NULL,
    lamp1 varchar(255) NOT NULL DEFAULT '',
    lamp2 varchar(255) NOT NULL DEFAULT '',
    survey_required boolean NOT NULL DEFAULT false
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
    category_id int PRIMARY KEY,
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
    number_id varchar(25),
    ticket_id bigint,
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
    number_id varchar(25),
    ticket_id bigint,
    title varchar(160),
    message text,
    is_read boolean DEFAULT false,
    created_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.fcm_tokens (
    id varchar(64) PRIMARY KEY,
    number_id varchar(25),
    token text,
    created_at timestamptz,
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id varchar(64) PRIMARY KEY,
    number_id varchar(25),
    token_hash varchar(64),
    expires_at timestamptz,
    created_at timestamptz
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_users_id_username')
        AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ux_users_id_username') THEN
        ALTER TABLE public.users
            RENAME CONSTRAINT uq_users_id_username TO ux_users_id_username;
    ELSIF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ux_users_id_username') THEN
        ALTER TABLE public.users
            ADD CONSTRAINT ux_users_id_username UNIQUE (id, username);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_service_id') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT fk_tickets_service_id
            FOREIGN KEY (id_service) REFERENCES public.services(id)
            ON UPDATE NO ACTION ON DELETE NO ACTION;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_staff_id') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT fk_tickets_staff_id
            FOREIGN KEY (id_staff) REFERENCES public.staff(id)
            ON UPDATE NO ACTION ON DELETE SET NULL;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_number_id') THEN
        ALTER TABLE public.tickets
            DROP CONSTRAINT fk_tickets_number_id;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_user_id') THEN
        ALTER TABLE public.tickets
            DROP CONSTRAINT fk_tickets_user_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tickets_number_id_username') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT fk_tickets_number_id_username
            FOREIGN KEY (number_id, username) REFERENCES public.users(id, username)
            ON UPDATE NO ACTION ON DELETE NO ACTION;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_number_id') THEN
        ALTER TABLE public.notifications
            ADD CONSTRAINT fk_notifications_number_id
            FOREIGN KEY (number_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_ticket_id') THEN
        ALTER TABLE public.notifications
            ADD CONSTRAINT fk_notifications_ticket_id
            FOREIGN KEY (ticket_id) REFERENCES public.tickets(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_fcm_tokens_number_id') THEN
        ALTER TABLE public.fcm_tokens
            ADD CONSTRAINT fk_fcm_tokens_number_id
            FOREIGN KEY (number_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_refresh_tokens_number_id') THEN
        ALTER TABLE public.refresh_tokens
            ADD CONSTRAINT fk_refresh_tokens_number_id
            FOREIGN KEY (number_id) REFERENCES public.users(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_category_templates_category_id') THEN
        ALTER TABLE public.category_templates
            ADD CONSTRAINT fk_category_templates_category_id
            FOREIGN KEY (category_id) REFERENCES public.services(id)
            ON UPDATE NO ACTION ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_category_templates_template_id') THEN
        ALTER TABLE public.category_templates
            ADD CONSTRAINT fk_category_templates_template_id
            FOREIGN KEY (template_id) REFERENCES public.survey_templates(id)
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

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_survey_responses_number_id') THEN
        ALTER TABLE public.survey_responses
            ADD CONSTRAINT fk_survey_responses_number_id
            FOREIGN KEY (number_id) REFERENCES public.users(id)
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
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_entity_values') THEN
        ALTER TABLE public.users
            ADD CONSTRAINT chk_users_entity_values
            CHECK (entity IN ('DOSEN', 'TENDIK', 'MAHASISWA', 'LAINNYA'));
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_entity_values') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_entity_values
            CHECK (entity IN ('DOSEN', 'TENDIK', 'MAHASISWA', 'LAINNYA'));
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_priority_values') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_priority_values
            CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH'));
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_status_values') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_status_values
            CHECK (status IN ('WAITING', 'ASSIGN', 'DONE', 'REJECT'));
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_tickets_guest_marker') THEN
        ALTER TABLE public.tickets
            ADD CONSTRAINT chk_tickets_guest_marker
            CHECK (
                NULLIF(BTRIM(number_id), '') IS NOT NULL AND
                (username IS NULL OR NULLIF(BTRIM(username), '') IS NOT NULL)
            );
    END IF;
END $$;

INSERT INTO public.services (id, name) VALUES
    (1, 'Lupa Password SSO'),
    (2, 'Registrasi SSO'),
    (3, 'Email Resmi Unila'),
    (4, 'Jaringan Internet'),
    (5, 'Website Down'),
    (6, 'Sistem Informasi'),
    (7, 'SIAKADU'),
    (99, 'Lainnya')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

CREATE UNIQUE INDEX IF NOT EXISTS ux_tickets_ticket_number ON public.tickets(ticket_number);
CREATE INDEX IF NOT EXISTS ix_tickets_ticket_date ON public.tickets(ticket_date);
CREATE INDEX IF NOT EXISTS ix_tickets_service_id ON public.tickets(id_service);
CREATE INDEX IF NOT EXISTS ix_tickets_status ON public.tickets(status);
CREATE INDEX IF NOT EXISTS ix_tickets_username ON public.tickets(username);
CREATE INDEX IF NOT EXISTS ix_tickets_number_id ON public.tickets(number_id);
CREATE INDEX IF NOT EXISTS ix_survey_responses_ticket_id ON public.survey_responses(ticket_id);
CREATE INDEX IF NOT EXISTS ix_survey_responses_number_id ON public.survey_responses(number_id);
CREATE INDEX IF NOT EXISTS ix_survey_responses_template_id ON public.survey_responses(template_id);
CREATE INDEX IF NOT EXISTS ix_survey_response_items_response_id ON public.survey_response_items(response_id);
CREATE INDEX IF NOT EXISTS ix_survey_response_items_question_id ON public.survey_response_items(question_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_survey_response_items_response_question ON public.survey_response_items(response_id, question_id);
CREATE INDEX IF NOT EXISTS ix_category_templates_template_id ON public.category_templates(template_id);
CREATE INDEX IF NOT EXISTS ix_refresh_tokens_number_id ON public.refresh_tokens(number_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_refresh_tokens_token_hash ON public.refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS ix_refresh_tokens_expires_at ON public.refresh_tokens(expires_at);

CREATE TABLE IF NOT EXISTS public.uploads (
    id varchar(30) PRIMARY KEY,
    original_name varchar(255) NOT NULL DEFAULT '',
    content_type varchar(100) NOT NULL DEFAULT '',
    data bytea NOT NULL,
    size bigint NOT NULL DEFAULT 0,
    created_at timestamptz
);
`

// migration20260311LampToText mengubah kolom lamp1/lamp2 dari varchar(255) ke text
// agar bisa menyimpan base64 file yang dikirim langsung dari frontend.
const migration20260311LampToText = `
ALTER TABLE public.tickets ALTER COLUMN lamp1 TYPE text;
ALTER TABLE public.tickets ALTER COLUMN lamp2 TYPE text;
`
