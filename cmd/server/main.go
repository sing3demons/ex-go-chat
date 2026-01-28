package main

import (
	"context"
	"net/http"
	"os"

	"realtime-chat-system/config"
	"realtime-chat-system/internal/handler"
	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/internal/websocket"
	"realtime-chat-system/pkg/auth"
	"realtime-chat-system/pkg/database"
	"realtime-chat-system/pkg/kp"
	"realtime-chat-system/pkg/logger"
	redisClient "realtime-chat-system/pkg/redis"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	// Initialize logger
	log := logger.New()
	log.Info("Starting Real-time Chat System...")

	// Load configuration
	cfg := config.Load()
	log.Infof("Configuration loaded: Server Port=%s, Database=%s", cfg.Server.Port, cfg.Database.Database)

	// Connect to MongoDB
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database.URI, cfg.Database.Database, cfg.Database.Timeout)
	if err != nil {
		log.Errorf("Failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			log.Errorf("Failed to disconnect from MongoDB: %v", err)
		}
	}()
	log.Info("Connected to MongoDB successfully")

	// Create indexes
	if err := db.CreateIndexes(context.Background()); err != nil {
		log.Errorf("Failed to create indexes: %v", err)
		os.Exit(1)
	}
	log.Info("Database indexes created successfully")

	// Connect to Redis
	redisClient, err := redisClient.NewClient(&cfg.Redis)
	if err != nil {
		log.Errorf("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Errorf("Failed to close Redis connection: %v", err)
		}
	}()
	log.Info("Connected to Redis successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Database)
	roomRepo := repository.NewRoomRepository(db.Database)
	messageRepo := repository.NewMessageRepository(db.Database)
	notificationRepo := repository.NewNotificationRepository(db.Database)

	// Initialize cache repositories (choose between memory or Redis)
	var messageCacheRepo repository.MessageCacheRepository
	// var _ repository.TypingRepository
	// var _ repository.SessionRepository
	// var _ repository.RateLimitRepository

	// Initialize presence repository (choose between memory or Redis)
	var presenceRepo repository.PresenceRepository
	var roomCacheRepo repository.RoomCacheRepository
	var userCacheRepo repository.UserCacheRepository
	if redisClient != nil {
		presenceRepo = repository.NewRedisPresenceRepository(redisClient)
		messageCacheRepo = repository.NewRedisMessageCacheRepository(redisClient)
		roomCacheRepo = repository.NewRedisRoomCacheRepository(redisClient)
		userCacheRepo = repository.NewRedisUserCacheRepository(redisClient)
		// typingRepo = repository.NewRedisTypingRepository(redisClient)
		// sessionRepo = repository.NewRedisSessionRepository(redisClient)
		// rateLimitRepo = repository.NewRedisRateLimitRepository(redisClient)
		log.Info("Using Redis repositories (including room and user cache)")
	} else {
		presenceRepo = repository.NewMemoryPresenceRepository()
		// For memory implementations, we could create memory versions
		// For now, set to nil to disable caching features
		messageCacheRepo = nil
		roomCacheRepo = repository.NewNoOpRoomCacheRepository()
		userCacheRepo = repository.NewNoOpUserCacheRepository()
		log.Info("Using memory presence repository, caching disabled")
	}

	log.Info("Repositories initialized")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiration)
	log.Info("JWT manager initialized")

	// Initialize services
	userService := service.NewUserService(userRepo, userCacheRepo)
	authService := service.NewAuthService(userRepo, userService, jwtManager)
	roomService := service.NewRoomService(roomRepo, roomCacheRepo)
	messageService := service.NewMessageService(messageRepo, roomRepo, messageCacheRepo)
	presenceService := service.NewPresenceService(presenceRepo, log)
	notificationService := service.NewNotificationService(notificationRepo)
	log.Info("Services initialized")

	// Initialize WebSocket hub and message handler
	hub := websocket.NewHub(presenceService, log) // We'll set the message handler after creating it
	wsMessageHandler := websocket.NewWSMessageHandler(hub, messageService, roomService, presenceService, notificationService, log)
	hub.SetMessageHandler(wsMessageHandler)
	// สร้าง Redis broadcaster (optional, สำหรับ multi-server)
	// redisBroadcaster := websocket.NewRedisBroadcaster(
	// 	redisClient.GetClient(), // Redis client
	// 	"server-1",              // unique server ID (ใช้ hostname หรือ env variable ก็ได้)
	// 	log,
	// )
	// Initialize Redis broadcaster for WebSocket
	// เชื่อม broadcaster เข้า hub
	// hub.SetRedisBroadcaster(redisBroadcaster)

	log.Info("WebSocket hub initialized")

	// Start hub in background
	go hub.Run()
	log.Info("WebSocket hub started")

	// Start presence heartbeat monitor
	go presenceService.StartHeartbeatMonitor(context.Background())
	log.Info("Presence heartbeat monitor started")

	// Initialize handlers
	// authMw      *middleware.AuthMiddleware
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)
	roomHandler := handler.NewRoomHandler(roomService, authMiddleware, hub)
	messageHandler := handler.NewMessageHandler(messageService, roomService, authMiddleware)
	notificationHandler := handler.NewNotificationHandler(notificationService, authMiddleware)
	userHandler := handler.NewUserHandler(userRepo, userService, roomService, authMiddleware, hub)
	wsHandler := websocket.NewHandler(hub, authService, roomService, presenceService, log)
	log.Info("Handlers initialized")

	loggerConfig := logger.LoggerConfig{
		Summary: logger.LogOutputConfig{Path: "./logs/summary/", Console: true, File: true},
		Detail:  logger.LogOutputConfig{Path: "./logs/detail/", Console: true, File: true},
		Rotation: logger.RotationConfig{
			MaxSize:    50 * 1024 * 1024, // 50MB
			MaxAge:     7,                // 7 days
			MaxBackups: 5,
			Compress:   true,
		},
	}
	app := kp.NewMicroservice(cfg, loggerConfig)
	app.Use(middleware.Recovery)
	app.Use(middleware.CORS)

	// Setup HTTP routes
	// mux := http.NewServeMux()

	// Health check endpoint
	app.GET("/health", func(ctx *kp.Ctx) {
		ctx.L("health")
		ctx.JSON(http.StatusOK, "OK")
	})

	// Register auth routes
	authHandler.RegisterRoutes(app)

	// Register room routes
	roomHandler.RegisterRoutes(app)

	// Register message routes
	messageHandler.RegisterRoutes(app)

	// Register user routes
	userHandler.RegisterRoutes(app)

	// Register notification routes
	notificationHandler.RegisterRoutes(app)

	app.Consume("test", func(ctx *kp.KafkaCtx) {
		ctx.L(ctx.Topic)

		ctx.Log.Flush(200, "")
	})

	// Register WebSocket route
	app.GET("/ws", wsHandler.ServeWS)
	log.Info("Routes registered")
	// Start the microservice
	app.Start()
}
