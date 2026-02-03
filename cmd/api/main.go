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

	for _, category := range service.DefaultCategories() {
		_ = categoryRepo.Upsert(category)
	}

	authService := service.NewAuthService(cfg, userRepo, refreshTokenRepo)
	fcmClient := fcm.NewClient(cfg.FCMEnabled, cfg.FCMCredentials)
	ticketService := service.NewTicketService(ticketRepo, categoryRepo, notificationRepo, tokenRepo, fcmClient)
	surveyService := service.NewSurveyService(surveyRepo, ticketRepo)
	notificationService := service.NewNotificationService(notificationRepo, tokenRepo)
	reportService := service.NewReportService(database)

	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)
	surveyHandler := handler.NewSurveyHandler(surveyService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	reportHandler := handler.NewReportHandler(reportService)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

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
	ticketHandler.RegisterRoutes(public, authGroup)
	surveyHandler.RegisterRoutes(public, authGroup, adminGroup)
	notificationHandler.RegisterRoutes(authGroup)
	reportHandler.RegisterRoutes(adminGroup)

	log.Printf("%s running on :%s", cfg.AppName, cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func validateConfig(cfg config.Config) {
	if strings.TrimSpace(cfg.JWTSecret) == "" || cfg.JWTSecret == "change-me" {
		log.Fatal("JWT_SECRET wajib di-set dan tidak boleh menggunakan default")
	}
	if strings.EqualFold(cfg.Environment, "production") {
		for _, origin := range cfg.CORSOrigins {
			if origin == "*" {
				log.Fatal("CORS_ORIGINS tidak boleh menggunakan '*' di production")
			}
		}
	}
}
