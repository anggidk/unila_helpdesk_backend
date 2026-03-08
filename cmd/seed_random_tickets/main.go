package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
	"sort"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/db"
	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
	"unila_helpdesk_backend/internal/util"

	"github.com/joho/godotenv"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type seedUser struct {
	ID       string
	Username string
	Name     string
	Email    string
	Entity   string
}

type seededQuestionnaire struct {
	Response domain.SurveyResponse
	Items    []domain.SurveyResponseItem
}

var (
	registeredServiceIDs = []int{
		domain.ServiceInternet,
		domain.ServiceWebsiteDown,
		domain.ServiceSistemInformasi,
		domain.ServiceSIAKADU,
		domain.ServiceLainnya,
	}
	priorityPool = []domain.TicketPriority{
		domain.PriorityLow,
		domain.PriorityLow,
		domain.PriorityMedium,
		domain.PriorityMedium,
		domain.PriorityHigh,
	}
	ticketNotesByService = map[int][]string{
		domain.ServiceInternet: {
			"Akses internet di gedung fakultas sering terputus saat jam kerja.",
			"Wi-Fi kampus tersambung tetapi tidak dapat membuka halaman web internal.",
			"Koneksi internet di laboratorium sangat lambat sejak pagi.",
			"Jaringan internet ruang administrasi tidak stabil saat unggah dokumen.",
		},
		domain.ServiceWebsiteDown: {
			"Website unit kerja tidak dapat diakses dan menampilkan error timeout.",
			"Halaman utama website fakultas tidak bisa dibuka dari jaringan kampus maupun luar kampus.",
			"Website layanan publik menampilkan pesan internal server error.",
			"Subdomain unit tidak dapat diakses setelah pembaruan konten.",
		},
		domain.ServiceSistemInformasi: {
			"Sistem informasi internal gagal memproses data saat menu laporan dibuka.",
			"Pengajuan pada sistem informasi tidak dapat disimpan walaupun data sudah lengkap.",
			"Beberapa fitur pada sistem informasi tidak dapat diakses oleh pengguna.",
			"Halaman dashboard sistem informasi memuat sangat lama dan sering gagal.",
		},
		domain.ServiceSIAKADU: {
			"SIAKADU tidak dapat menampilkan jadwal kuliah pada akun saya.",
			"Menu pengisian KRS di SIAKADU gagal disimpan.",
			"Nilai mata kuliah belum muncul pada halaman hasil studi SIAKADU.",
			"Akses login ke SIAKADU berhasil tetapi halaman utama kosong.",
		},
		domain.ServiceLainnya: {
			"Saya memerlukan bantuan pada layanan TI kampus yang belum tercakup kategori lain.",
			"Terdapat kendala pada layanan digital kampus dan membutuhkan tindak lanjut teknis.",
			"Permintaan bantuan terkait layanan teknologi informasi perlu ditangani lebih lanjut.",
			"Ada masalah operasional layanan kampus yang memerlukan dukungan helpdesk.",
		},
	}
	staffNotesPool = []string{
		"Petugas sudah melakukan verifikasi awal dan layanan dinyatakan selesai.",
		"Permasalahan berhasil ditangani sesuai kebutuhan pelapor.",
		"Tim helpdesk telah melakukan tindak lanjut dan memastikan layanan kembali normal.",
		"Penanganan selesai dan pengguna telah menerima konfirmasi perbaikan.",
	}
	suggestionPool = []string{
		"Layanan sudah baik, mohon dipertahankan.",
		"Informasi progres penanganan dapat dibuat lebih rinci.",
		"Respons petugas sudah baik, semoga waktu penanganan bisa lebih cepat.",
		"Sistem sudah membaik, mohon kualitas layanan dijaga secara konsisten.",
		"Saran saya, notifikasi perkembangan layanan dibuat lebih jelas.",
	}
)

func main() {
	count := flag.Int("count", 300, "jumlah tiket random yang akan dibuat")
	days := flag.Int("days", 60, "rentang hari ke belakang untuk data tiket")
	flag.Parse()

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if !strings.EqualFold(cfg.Environment, "production") {
		if err := db.EnsureDatabase(cfg); err != nil {
			log.Fatalf("ensure database failed: %v", err)
		}
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	usersByEntity, err := loadRegisteredUsersByEntity(database)
	if err != nil {
		log.Fatalf("load registered users failed: %v", err)
	}
	if len(usersByEntity) == 0 {
		log.Fatalf("tidak ada user registered. jalankan go run ./cmd/seed terlebih dahulu")
	}

	surveyRepo := repository.NewSurveyRepository(database)
	templatesByService, err := loadTemplatesByService(surveyRepo, registeredServiceIDs)
	if err != nil {
		log.Fatalf("load survey templates failed: %v", err)
	}

	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	if err := seedRandomTickets(database, rng, usersByEntity, templatesByService, *count, *days); err != nil {
		log.Fatalf("seed random tickets failed: %v", err)
	}

	log.Printf("berhasil membuat %d tiket random beserta response survei untuk %d hari terakhir", *count, *days)
}

func loadRegisteredUsersByEntity(database *gorm.DB) (map[string][]seedUser, error) {
	var users []domain.User
	if err := database.
		Where("role = ? AND is_active = ?", domain.RoleRegistered, true).
		Order("entity asc, username asc").
		Find(&users).Error; err != nil {
		return nil, err
	}

	grouped := map[string][]seedUser{}
	for _, user := range users {
		entity := normalizeEntity(user.Entity)
		grouped[entity] = append(grouped[entity], seedUser{
			ID:       strings.TrimSpace(user.ID),
			Username: strings.TrimSpace(user.Username),
			Name:     strings.TrimSpace(user.Name),
			Email:    strings.TrimSpace(user.Email),
			Entity:   entity,
		})
	}

	for entity, entityUsers := range grouped {
		if len(entityUsers) == 0 {
			delete(grouped, entity)
		}
	}
	return grouped, nil
}

func loadTemplatesByService(
	repo *repository.SurveyRepository,
	serviceIDs []int,
) (map[int]*domain.SurveyTemplate, error) {
	templates := make(map[int]*domain.SurveyTemplate, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		template, err := repo.FindByCategory(fmt.Sprintf("%d", serviceID))
		if err != nil {
			return nil, fmt.Errorf("template kategori %d tidak ditemukan: %w", serviceID, err)
		}
		if template == nil || strings.TrimSpace(template.ID) == "" || len(template.Questions) == 0 {
			return nil, fmt.Errorf("template kategori %d belum lengkap", serviceID)
		}
		templates[serviceID] = template
	}
	return templates, nil
}

func seedRandomTickets(
	database *gorm.DB,
	rng *mathrand.Rand,
	usersByEntity map[string][]seedUser,
	templatesByService map[int]*domain.SurveyTemplate,
	count int,
	days int,
) error {
	if count <= 0 {
		return fmt.Errorf("count harus lebih besar dari 0")
	}
	if days <= 0 {
		return fmt.Errorf("days harus lebih besar dari 0")
	}

	entities := entityOrder(usersByEntity)
	if len(entities) == 0 {
		return fmt.Errorf("tidak ada entitas registered yang dapat digunakan")
	}

	return database.Transaction(func(tx *gorm.DB) error {
		for index := 0; index < count; index++ {
			entity := entities[index%len(entities)]
			user := pickUser(rng, usersByEntity[entity])
			serviceID := registeredServiceIDs[rng.Intn(len(registeredServiceIDs))]
			template := templatesByService[serviceID]
			if template == nil {
				return fmt.Errorf("template untuk service %d tidak tersedia", serviceID)
			}

			ticketTime := randomTimestampWithinLastDays(rng, days)
			responseTime := randomResponseTime(rng, ticketTime)
			ticketNumber, err := generateUniqueTicketNumber(tx, rng)
			if err != nil {
				return err
			}

			username := user.Username
			numberID := user.ID
			ticket := domain.Ticket{
				TicketNumber:   ticketNumber,
				Username:       &username,
				NumberID:       &numberID,
				Name:           user.Name,
				Email:          user.Email,
				Entity:         user.Entity,
				ServiceID:      serviceID,
				Notes:          pickString(rng, ticketNotesByService[serviceID]),
				StaffNotes:     pickString(rng, staffNotesPool),
				Priority:       priorityPool[rng.Intn(len(priorityPool))],
				IsReject:       false,
				IsAssign:       true,
				IsDone:         true,
				Status:         domain.StatusDone,
				Lamp1:          "",
				Lamp2:          "",
				SurveyRequired: false,
				CreatedAt:      ticketTime,
			}
			if err := tx.Create(&ticket).Error; err != nil {
				return fmt.Errorf("create ticket %d: %w", index+1, err)
			}

			questionnaire, err := buildSeededQuestionnaire(rng, template, ticket.ID, user.ID, responseTime)
			if err != nil {
				return fmt.Errorf("build questionnaire untuk ticket %s: %w", ticket.TicketNumber, err)
			}
			if err := tx.Create(&questionnaire.Response).Error; err != nil {
				return fmt.Errorf("create survey response untuk ticket %s: %w", ticket.TicketNumber, err)
			}
			if len(questionnaire.Items) > 0 {
				if err := tx.Create(&questionnaire.Items).Error; err != nil {
					return fmt.Errorf("create survey items untuk ticket %s: %w", ticket.TicketNumber, err)
				}
			}
		}
		return nil
	})
}

func entityOrder(usersByEntity map[string][]seedUser) []string {
	entities := make([]string, 0, len(usersByEntity))
	for entity, users := range usersByEntity {
		if len(users) == 0 {
			continue
		}
		entities = append(entities, entity)
	}

	orderWeight := map[string]int{
		domain.EntityMahasiswa: 1,
		domain.EntityDosen:     2,
		domain.EntityTendik:    3,
		domain.EntityLainnya:   4,
	}
	sort.Slice(entities, func(i, j int) bool {
		left := orderWeight[entities[i]]
		right := orderWeight[entities[j]]
		if left == right {
			return entities[i] < entities[j]
		}
		return left < right
	})
	return entities
}

func pickUser(rng *mathrand.Rand, users []seedUser) seedUser {
	return users[rng.Intn(len(users))]
}

func generateUniqueTicketNumber(tx *gorm.DB, rng *mathrand.Rand) (string, error) {
	const maxRetries = 20
	for attempt := 0; attempt < maxRetries; attempt++ {
		value := fmt.Sprintf("%06d", rng.Intn(1000000))
		var count int64
		if err := tx.Model(&domain.Ticket{}).Where("ticket_number = ?", value).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return value, nil
		}
	}
	return "", fmt.Errorf("gagal membuat nomor tiket unik")
}

func randomTimestampWithinLastDays(rng *mathrand.Rand, days int) time.Time {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayOffset := rng.Intn(days)
	start := startOfToday.AddDate(0, 0, -dayOffset)
	end := start.Add(24 * time.Hour)
	if end.After(now) {
		end = now
	}
	if !end.After(start) {
		return now
	}
	secondsRange := int(end.Sub(start).Seconds())
	if secondsRange <= 1 {
		return start
	}
	return start.Add(time.Duration(rng.Intn(secondsRange)) * time.Second)
}

func randomResponseTime(rng *mathrand.Rand, ticketTime time.Time) time.Time {
	now := time.Now()
	minutesLater := rng.Intn(18*60) + 30
	responseTime := ticketTime.Add(time.Duration(minutesLater) * time.Minute)
	if responseTime.After(now) {
		responseTime = now
	}
	if responseTime.Before(ticketTime) {
		return ticketTime
	}
	return responseTime
}

func buildSeededQuestionnaire(
	rng *mathrand.Rand,
	template *domain.SurveyTemplate,
	ticketID int,
	userID string,
	createdAt time.Time,
) (seededQuestionnaire, error) {
	responseID := util.NewID(32)
	items := make([]domain.SurveyResponseItem, 0, len(template.Questions))
	var total float64
	var scoredCount int

	for _, question := range template.Questions {
		answer, scoreValue, err := buildQuestionAnswer(rng, question)
		if err != nil {
			return seededQuestionnaire{}, err
		}
		payload, err := json.Marshal(answer)
		if err != nil {
			return seededQuestionnaire{}, err
		}

		item := domain.SurveyResponseItem{
			ID:          util.NewID(32),
			ResponseID:  responseID,
			QuestionID:  question.ID,
			AnswerValue: datatypes.JSON(payload),
			CreatedAt:   createdAt,
		}
		if scoreValue != nil {
			item.ScoreValue = scoreValue
			total += *scoreValue
			scoredCount++
		}
		items = append(items, item)
	}

	score := 0.0
	if scoredCount > 0 {
		score = total / float64(scoredCount)
	}

	return seededQuestionnaire{
		Response: domain.SurveyResponse{
			ID:         responseID,
			TicketID:   ticketID,
			UserID:     userID,
			TemplateID: template.ID,
			Score:      score,
			CreatedAt:  createdAt,
		},
		Items: items,
	}, nil
}

func buildQuestionAnswer(
	rng *mathrand.Rand,
	question domain.SurveyQuestion,
) (interface{}, *float64, error) {
	switch question.Type {
	case domain.QuestionLikert4:
		rawValue := weightedLikert4Value(rng)
		score := normalizeLikert(rawValue, 4)
		return rawValue, &score, nil
	case domain.QuestionText:
		return pickString(rng, suggestionPool), nil, nil
	default:
		return nil, nil, fmt.Errorf("tipe pertanyaan %s belum didukung untuk seed", question.Type)
	}
}

func weightedLikert4Value(rng *mathrand.Rand) int {
	roll := rng.Intn(100)
	switch {
	case roll < 10:
		return 2
	case roll < 65:
		return 3
	default:
		return 4
	}
}

func normalizeLikert(value int, max int) float64 {
	if max <= 1 {
		return 5
	}
	return 1 + ((float64(value)-1)*4)/float64(max-1)
}

func normalizeEntity(entity string) string {
	switch strings.ToUpper(strings.TrimSpace(entity)) {
	case domain.EntityMahasiswa:
		return domain.EntityMahasiswa
	case domain.EntityDosen:
		return domain.EntityDosen
	case domain.EntityTendik:
		return domain.EntityTendik
	default:
		return domain.EntityLainnya
	}
}

func pickString(rng *mathrand.Rand, items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[rng.Intn(len(items))]
}
