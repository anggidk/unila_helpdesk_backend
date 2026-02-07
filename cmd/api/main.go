package main

import (
	"log"
	"net/http"
	"strings"

	"unila_helpdesk_backend/internal/config"
	"unila_helpdesk_backend/internal/db"
	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/fcm"
	"unila_helpdesk_backend/internal/handler"
	"unila_helpdesk_backend/internal/middleware"
	"unila_helpdesk_backend/internal/repository"
	"unila_helpdesk_backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	validateConfig(cfg)

	// Set Gin mode berdasarkan environment
	if strings.EqualFold(cfg.Environment, "production") {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("database ensure failed: %v", err)
	}
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	db.MustAutoMigrate(database)

	userRepo := repository.NewUserRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	ticketRepo := repository.NewTicketRepository(database)
	surveyRepo := repository.NewSurveyRepository(database)
	notificationRepo := repository.NewNotificationRepository(database)
	tokenRepo := repository.NewFCMTokenRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	attachmentRepo := repository.NewAttachmentRepository(database)
	reportRepo := repository.NewReportRepository(database)

	for _, category := range service.DefaultCategories() {
		_ = categoryRepo.Upsert(category)
	}
	if err := service.CleanupDeprecatedCategories(database, service.CategoryLainnya); err != nil {
		log.Fatalf("cleanup deprecated categories failed: %v", err)
	}

	authService := service.NewAuthService(cfg, userRepo, refreshTokenRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	fcmClient := fcm.NewClient(cfg.FCMEnabled, cfg.FCMCredentials)
	ticketService := service.NewTicketService(
		ticketRepo,
		categoryRepo,
		notificationRepo,
		tokenRepo,
		attachmentRepo,
		fcmClient,
		domain.TicketStatus(cfg.TicketInitialStatus),
	)
	surveyService := service.NewSurveyService(surveyRepo, ticketRepo)
	notificationService := service.NewNotificationService(notificationRepo, tokenRepo)
	reportService := service.NewReportService(reportRepo, categoryRepo, surveyRepo)

	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	surveyHandler := handler.NewSurveyHandler(surveyService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	reportHandler := handler.NewReportHandler(reportService)
	uploadHandler := handler.NewUploadHandler(cfg.BaseURL, attachmentRepo)

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	corsOrigins := strings.Split(cfg.CORSOrigins, ",")
	for i, origin := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(origin)
	}
	router.Use(middleware.CORSMiddleware(corsOrigins))

	authRequired := middleware.AuthMiddleware(authService, userRepo, true)
	authOptional := middleware.AuthMiddleware(authService, userRepo, false)

	api := router.Group("")
	public := api.Group("", authOptional)
	authGroup := api.Group("", authRequired)
	adminGroup := api.Group("", authRequired, middleware.RequireRole(domain.RoleAdmin))
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler.RegisterRoutes(api)
	categoryHandler.RegisterRoutes(public)
	categoryHandler.RegisterAdminRoutes(adminGroup)
	ticketHandler.RegisterRoutes(public, authGroup)
	surveyHandler.RegisterRoutes(public, authGroup, adminGroup)
	notificationHandler.RegisterRoutes(authGroup)
	reportHandler.RegisterRoutes(adminGroup)
	uploadHandler.RegisterRoutes(public)

	log.Printf("%s running on :%s", cfg.AppName, cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func validateConfig(cfg config.Config) {
	// Validate required fields
	if strings.TrimSpace(cfg.AppName) == "" {
		log.Fatal("APP_NAME is required")
	}
	if strings.TrimSpace(cfg.Environment) == "" {
		log.Fatal("APP_ENV is required")
	}
	if strings.TrimSpace(cfg.HTTPPort) == "" {
		log.Fatal("HTTP_PORT is required")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		log.Fatal("BASE_URL is required")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.JWTExpiry == 0 {
		log.Fatal("JWT_EXPIRY is required")
	}
	if cfg.JWTRefreshExpiry == 0 {
		log.Fatal("JWT_REFRESH_EXPIRY is required")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.DatabaseMaxConns == 0 {
		log.Fatal("DB_MAX_CONNS is required and must be > 0")
	}
	if cfg.DatabaseIdleConns == 0 {
		log.Fatal("DB_IDLE_CONNS is required and must be > 0")
	}
	if strings.TrimSpace(cfg.CORSOrigins) == "" {
		log.Fatal("CORS_ORIGINS is required")
	}

	// Production-specific validation
	if strings.EqualFold(cfg.Environment, "production") {
		corsOrigins := strings.Split(cfg.CORSOrigins, ",")
		for _, origin := range corsOrigins {
			if strings.TrimSpace(origin) == "*" {
				log.Fatal("CORS_ORIGINS cannot use '*' in production")
			}
		}
	}
}
