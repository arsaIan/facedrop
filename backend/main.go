package main

import (
	"facedrop/config"
	"facedrop/controller"
	"facedrop/deepface"
	"facedrop/events"
	"facedrop/logger"
	"facedrop/middleware"
	"facedrop/models"
	"facedrop/repository"
	"facedrop/sender"
	"facedrop/service"
	"facedrop/storage"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
func Migrate(db *gorm.DB) {
	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&models.User{}, 
		&models.Event{}, 
		&models.Photo{}, 
		&models.UserFace{}, 
		&models.EventSubscriber{},
		&models.EventMatch{},
		); err != nil {
		logger.Fatal("Failed to migrate database", logger.Error(err))
	}
	logger.Info("Database migration completed")
}
func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Initialize logger
	logger.InitLogger(cfg.ServerConfig.Environment)
	logger.Info("Starting application", 
		logger.String("environment", cfg.ServerConfig.Environment),
		logger.Int("port", cfg.ServerConfig.Port),
	)

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Error(err))
	}
	logger.Info("Database connection established")

	if cfg.DatabaseConfig.Migrate {
		Migrate(db)
	}

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Add logging middleware
	r.Use(logger.HTTPLogger())

	s3Client, err := storage.NewS3Client(cfg)
	if err != nil {
		logger.Fatal("Failed to create storage client", logger.Error(err))
	}
	storageClient := service.NewStorageClient(s3Client)
	deepfaceClient := deepface.NewDeepFaceClient(cfg.DeepFaceConfig.URL)
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	qrService := service.NewQRService(cfg.ServerConfig.BaseURL)

	//initialize sender
	sender := sender.NewEmailSender(cfg)
	//initialize processor
	processor := events.NewEventProcessor(eventRepo, s3Client, cfg, sender)
	// Initialize services
	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo, qrService, deepfaceClient, processor)

	// Initialize controllers
	userController := controller.NewUserController(userService, cfg, storageClient)
	eventController := controller.NewEventController(eventService, storageClient, cfg)
	authMiddleware := middleware.NewAuthMiddleware(cfg)
	// Setup routes
	setupRoutes(r, userController, eventController, authMiddleware)

	// Start server
	logger.Info("Server starting", logger.String("port", fmt.Sprintf(":%d", cfg.ServerConfig.Port)))
	if err := r.Run(fmt.Sprintf(":%d", cfg.ServerConfig.Port)); err != nil {
		logger.Fatal("Failed to start server", logger.Error(err))
	}
}

func setupRoutes(r *gin.Engine, userController *controller.UserController,
	 eventController *controller.EventController, authMiddleware *middleware.AuthMiddleware) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	//public routes
	publicRoutes := r.Group("/")
	{
		publicRoutes.POST("/login", userController.Login)
		publicRoutes.POST("/register", userController.Register)
	}
	// User routes
	users := r.Group("/users").Use(authMiddleware.Auth())
	{
		users.GET("/:id", userController.GetUser)
		users.PUT("/:id", userController.UpdateUser)
		users.DELETE("/:id", userController.DeleteUser)
		users.POST("/:id/face", userController.UploadFace)
	}
	// Event routes
	eventRoutes := r.Group("/events").Use(authMiddleware.Auth())
	{
		eventRoutes.POST("", eventController.CreateEvent)
		eventRoutes.GET("", eventController.GetAllEvents)
		eventRoutes.GET("/:id", eventController.GetEvent)
		eventRoutes.PUT("/:id", eventController.UpdateEvent)
		eventRoutes.DELETE("/:id", eventController.DeleteEvent)
		eventRoutes.POST("/:id/subscribe", eventController.SubscribeToEvent)
		eventRoutes.DELETE("/:id/subscribe", eventController.UnsubscribeFromEvent)
		eventRoutes.POST("/:id/photos", eventController.AddPhoto)
		eventRoutes.GET("/:id/photos", eventController.GetEventPhotos)
		eventRoutes.GET("/:id/my-photos", eventController.GetSubscriberPhotos)
		eventRoutes.POST("/:id/ready", eventController.PushEventToReadyQueue)
		eventRoutes.GET("/:id/subscribers", eventController.GetEventSubscribers)
	}
} 