package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/db"
	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
	appservice "unila_helpdesk_backend/internal/service"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
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

type seedSurveyQuestion struct {
	ID      string
	Text    string
	Type    domain.SurveyQuestionType
	Options []string
}

type seedSurveyTemplate struct {
	ID          string
	CategoryID  int
	Title       string
	Description string
	Framework   string
	Questions   []seedSurveyQuestion
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

func normalizeSeedEntity(entity string) string {
	switch strings.ToUpper(strings.TrimSpace(entity)) {
	case domain.EntityDosen:
		return domain.EntityDosen
	case domain.EntityTendik:
		return domain.EntityTendik
	case domain.EntityMahasiswa:
		return domain.EntityMahasiswa
	default:
		return domain.EntityLainnya
	}
}

func defaultUserID(username string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(username))
	cleaned = strings.NewReplacer(" ", "", ".", "", "-", "", "_", "").Replace(cleaned)
	if len(cleaned) > 25 {
		return cleaned[:25]
	}
	return cleaned
}

func upsertUser(database *gorm.DB, seed seedUser) error {
	email := strings.ToLower(strings.TrimSpace(seed.Email))
	username := strings.ToLower(strings.TrimSpace(seed.Username))
	password := strings.TrimSpace(seed.Password)
	entity := normalizeSeedEntity(seed.Entity)

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
			"entity":        entity,
			"password_hash": hashed,
			"is_active":     true,
		}
		return database.Model(&existing).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	user := domain.User{
		ID:           defaultUserID(username),
		Username:     username,
		PasswordHash: hashed,
		Name:         seed.Name,
		Email:        email,
		Role:         seed.Role,
		Entity:       entity,
		IsActive:     true,
	}

	return database.Create(&user).Error
}

func seedSurveyTemplates() []seedSurveyTemplate {
	likert4Options := []string{
		"Sangat Tidak Setuju",
		"Tidak Setuju",
		"Setuju",
		"Sangat Setuju",
	}
	description := "Gunakan skala Likert 4: Sangat Tidak Setuju, Tidak Setuju, Setuju, dan Sangat Setuju."

	return []seedSurveyTemplate{
		{
			ID:          "TPLNETWK001",
			CategoryID:  domain.ServiceInternet,
			Title:       "Survei Kepuasan Layanan Jaringan Internet",
			Description: description,
			Framework:   "Likert 4",
			Questions: []seedSurveyQuestion{
				{ID: "internet_q1", Text: "Gangguan jaringan internet saya ditangani dengan cepat.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q2", Text: "Petugas memahami masalah jaringan yang saya laporkan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q3", Text: "Informasi perkembangan penanganan gangguan disampaikan dengan jelas.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q4", Text: "Koneksi internet kembali dapat digunakan setelah penanganan dilakukan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q5", Text: "Stabilitas jaringan internet setelah perbaikan sudah memadai.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q6", Text: "Kecepatan akses internet setelah perbaikan sesuai dengan kebutuhan saya.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q7", Text: "Solusi yang diberikan sesuai dengan masalah jaringan yang saya alami.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q8", Text: "Secara keseluruhan, saya puas terhadap layanan penanganan jaringan internet.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "internet_q9", Text: "Saran atau masukan Anda untuk peningkatan layanan jaringan internet.", Type: domain.QuestionText, Options: []string{}},
			},
		},
		{
			ID:          "TPLWEBTD001",
			CategoryID:  domain.ServiceWebsiteDown,
			Title:       "Survei Kepuasan Layanan Penanganan Website Down",
			Description: description,
			Framework:   "Likert 4",
			Questions: []seedSurveyQuestion{
				{ID: "webdown_q1", Text: "Laporan gangguan website saya ditanggapi dengan cepat.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q2", Text: "Petugas memahami masalah website yang saya laporkan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q3", Text: "Informasi mengenai status gangguan website disampaikan dengan jelas.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q4", Text: "Website kembali dapat diakses setelah penanganan dilakukan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q5", Text: "Fungsi utama website kembali berjalan dengan baik.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q6", Text: "Waktu penanganan gangguan website sesuai dengan harapan saya.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q7", Text: "Solusi yang diberikan sesuai dengan masalah website yang terjadi.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q8", Text: "Secara keseluruhan, saya puas terhadap layanan penanganan website down.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "webdown_q9", Text: "Saran atau masukan Anda untuk peningkatan layanan penanganan website.", Type: domain.QuestionText, Options: []string{}},
			},
		},
		{
			ID:          "TPLSYSIF001",
			CategoryID:  domain.ServiceSistemInformasi,
			Title:       "Survei Kepuasan Layanan Sistem Informasi",
			Description: description,
			Framework:   "Likert 4",
			Questions: []seedSurveyQuestion{
				{ID: "sysinfo_q1", Text: "Permasalahan sistem informasi saya ditangani dengan cepat.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q2", Text: "Petugas memahami kendala yang saya alami pada sistem informasi.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q3", Text: "Penjelasan atau informasi terkait penanganan diberikan dengan jelas.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q4", Text: "Sistem informasi kembali dapat digunakan dengan baik setelah penanganan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q5", Text: "Fitur atau fungsi yang bermasalah sudah berjalan normal kembali.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q6", Text: "Solusi yang diberikan membantu saya menyelesaikan kendala pada sistem.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q7", Text: "Penanganan yang dilakukan sesuai dengan kebutuhan saya.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q8", Text: "Secara keseluruhan, saya puas terhadap layanan sistem informasi.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "sysinfo_q9", Text: "Saran atau masukan Anda untuk peningkatan layanan sistem informasi.", Type: domain.QuestionText, Options: []string{}},
			},
		},
		{
			ID:          "TPLSIAK001",
			CategoryID:  domain.ServiceSIAKADU,
			Title:       "Survei Kepuasan Layanan SIAKADU",
			Description: description,
			Framework:   "Likert 4",
			Questions: []seedSurveyQuestion{
				{ID: "siakadu_q1", Text: "Kendala pada SIAKADU saya ditangani dengan cepat.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q2", Text: "Petugas memahami masalah yang saya alami pada SIAKADU.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q3", Text: "Informasi mengenai proses penanganan disampaikan dengan jelas.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q4", Text: "SIAKADU kembali dapat diakses setelah masalah ditangani.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q5", Text: "Fitur utama SIAKADU yang saya butuhkan kembali berjalan dengan baik.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q6", Text: "Penanganan yang dilakukan membantu saya melanjutkan aktivitas akademik.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q7", Text: "Solusi yang diberikan sesuai dengan kendala SIAKADU yang saya laporkan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q8", Text: "Secara keseluruhan, saya puas terhadap layanan penanganan SIAKADU.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "siakadu_q9", Text: "Saran atau masukan Anda untuk peningkatan layanan SIAKADU.", Type: domain.QuestionText, Options: []string{}},
			},
		},
		{
			ID:          "TPLOTHR001",
			CategoryID:  domain.ServiceLainnya,
			Title:       "Survei Kepuasan Layanan Lainnya",
			Description: description,
			Framework:   "Likert 4",
			Questions: []seedSurveyQuestion{
				{ID: "lainnya_q1", Text: "Keluhan atau permintaan layanan saya ditanggapi dengan cepat.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q2", Text: "Petugas memahami kebutuhan atau masalah yang saya sampaikan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q3", Text: "Informasi terkait proses penanganan diberikan dengan jelas.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q4", Text: "Penanganan dilakukan secara profesional dan sopan.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q5", Text: "Solusi yang diberikan sesuai dengan masalah atau kebutuhan saya.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q6", Text: "Permasalahan saya berhasil ditangani dengan baik.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q7", Text: "Waktu penyelesaian layanan sesuai dengan harapan saya.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q8", Text: "Secara keseluruhan, saya puas terhadap layanan yang saya terima.", Type: domain.QuestionLikert4, Options: likert4Options},
				{ID: "lainnya_q9", Text: "Saran atau masukan Anda untuk peningkatan layanan pada kategori ini.", Type: domain.QuestionText, Options: []string{}},
			},
		},
	}
}

func upsertSurveyTemplate(
	repo *repository.SurveyRepository,
	seed seedSurveyTemplate,
) error {
	now := time.Now()
	questions := make([]domain.SurveyQuestion, 0, len(seed.Questions))
	for _, question := range seed.Questions {
		options, err := json.Marshal(question.Options)
		if err != nil {
			return fmt.Errorf("marshal options %s: %w", question.ID, err)
		}
		questions = append(questions, domain.SurveyQuestion{
			ID:         question.ID,
			TemplateID: seed.ID,
			Text:       question.Text,
			Type:       question.Type,
			Options:    datatypes.JSON(options),
			CreatedAt:  now,
		})
	}

	existing, err := repo.FindByID(seed.ID)
	if err == nil {
		existing.Title = seed.Title
		existing.Description = seed.Description
		existing.Framework = seed.Framework
		existing.UpdatedAt = now
		existing.Questions = questions
		return repo.ReplaceTemplate(existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	template := domain.SurveyTemplate{
		ID:          seed.ID,
		Title:       seed.Title,
		Description: seed.Description,
		Framework:   seed.Framework,
		Questions:   questions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return repo.CreateTemplate(&template)
}

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if !strings.EqualFold(cfg.Environment, "production") {
		if err := db.EnsureDatabase(cfg); err != nil {
			log.Fatalf("ensure database failed: %v", err)
		}
	} else {
		log.Printf("skip database ensure in production")
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	categoryRepo := repository.NewCategoryRepository(database)
	surveyRepo := repository.NewSurveyRepository(database)

	for _, category := range appservice.DefaultCategories() {
		if err := categoryRepo.Upsert(category); err != nil {
			log.Fatalf("seed service %d gagal: %v", category.ID, err)
		}
	}

	for _, template := range seedSurveyTemplates() {
		if err := upsertSurveyTemplate(surveyRepo, template); err != nil {
			log.Fatalf("seed survey template %s gagal: %v", template.ID, err)
		}
		if err := categoryRepo.BindTemplateToCategory(template.CategoryID, template.ID); err != nil {
			log.Fatalf("bind template %s ke category %d gagal: %v", template.ID, template.CategoryID, err)
		}
	}

	seeds := []seedUser{
		// ===== Admin (10) =====
		{Username: "admin1", Password: "admin123", Name: "Admin 1", Email: "admin1@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin2", Password: "admin123", Name: "Admin 2", Email: "admin2@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin3", Password: "admin123", Name: "Admin 3", Email: "admin3@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin4", Password: "admin123", Name: "Admin 4", Email: "admin4@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin5", Password: "admin123", Name: "Admin 5", Email: "admin5@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin6", Password: "admin123", Name: "Admin 6", Email: "admin6@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin7", Password: "admin123", Name: "Admin 7", Email: "admin7@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin8", Password: "admin123", Name: "Admin 8", Email: "admin8@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin9", Password: "admin123", Name: "Admin 9", Email: "admin9@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},
		{Username: "admin10", Password: "admin123", Name: "Admin 10", Email: "admin10@unila.ac.id", Role: domain.RoleAdmin, Entity: "Admin"},

		// ===== Mahasiswa (40) =====
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
		{Username: "mahasiswa36", Password: "mahasiswa123", Name: "Mahasiswa 36", Email: "mahasiswa36@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa37", Password: "mahasiswa123", Name: "Mahasiswa 37", Email: "mahasiswa37@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa38", Password: "mahasiswa123", Name: "Mahasiswa 38", Email: "mahasiswa38@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa39", Password: "mahasiswa123", Name: "Mahasiswa 39", Email: "mahasiswa39@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},
		{Username: "mahasiswa40", Password: "mahasiswa123", Name: "Mahasiswa 40", Email: "mahasiswa40@unila.ac.id", Role: domain.RoleRegistered, Entity: "Mahasiswa"},

		// ===== Dosen (10) =====
		{Username: "dosen1", Password: "dosen123", Name: "Dosen 1", Email: "dosen1@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen2", Password: "dosen123", Name: "Dosen 2", Email: "dosen2@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen3", Password: "dosen123", Name: "Dosen 3", Email: "dosen3@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen4", Password: "dosen123", Name: "Dosen 4", Email: "dosen4@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen5", Password: "dosen123", Name: "Dosen 5", Email: "dosen5@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen6", Password: "dosen123", Name: "Dosen 6", Email: "dosen6@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen7", Password: "dosen123", Name: "Dosen 7", Email: "dosen7@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen8", Password: "dosen123", Name: "Dosen 8", Email: "dosen8@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen9", Password: "dosen123", Name: "Dosen 9", Email: "dosen9@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},
		{Username: "dosen10", Password: "dosen123", Name: "Dosen 10", Email: "dosen10@unila.ac.id", Role: domain.RoleRegistered, Entity: "Dosen"},

		// ===== Tendik (10) =====
		{Username: "tendik1", Password: "tendik123", Name: "Tendik 1", Email: "tendik1@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik2", Password: "tendik123", Name: "Tendik 2", Email: "tendik2@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik3", Password: "tendik123", Name: "Tendik 3", Email: "tendik3@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik4", Password: "tendik123", Name: "Tendik 4", Email: "tendik4@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik5", Password: "tendik123", Name: "Tendik 5", Email: "tendik5@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik6", Password: "tendik123", Name: "Tendik 6", Email: "tendik6@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik7", Password: "tendik123", Name: "Tendik 7", Email: "tendik7@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik8", Password: "tendik123", Name: "Tendik 8", Email: "tendik8@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik9", Password: "tendik123", Name: "Tendik 9", Email: "tendik9@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
		{Username: "tendik10", Password: "tendik123", Name: "Tendik 10", Email: "tendik10@unila.ac.id", Role: domain.RoleRegistered, Entity: "Tendik"},
	}

	for _, seed := range seeds {
		if err := upsertUser(database, seed); err != nil {
			log.Fatalf("seed user %s gagal: %v", seed.Username, err)
		}
	}

	log.Println("seed users selesai")
}
