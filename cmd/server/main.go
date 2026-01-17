package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"realtime-chat-system/config"
	"realtime-chat-system/internal/handler"
	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/internal/websocket"
	"realtime-chat-system/pkg/auth"
	"realtime-chat-system/pkg/database"
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
	var _ repository.TypingRepository
	var _ repository.SessionRepository
	var _ repository.RateLimitRepository
	
	// Initialize presence repository (choose between memory or Redis)
	var presenceRepo repository.PresenceRepository
	if redisClient != nil {
		presenceRepo = repository.NewRedisPresenceRepository(redisClient)
		messageCacheRepo = repository.NewRedisMessageCacheRepository(redisClient)
		// typingRepo = repository.NewRedisTypingRepository(redisClient)
		// sessionRepo = repository.NewRedisSessionRepository(redisClient)
		// rateLimitRepo = repository.NewRedisRateLimitRepository(redisClient)
		log.Info("Using Redis repositories")
	} else {
		presenceRepo = repository.NewMemoryPresenceRepository()
		// For memory implementations, we could create memory versions
		// For now, set to nil to disable caching features
		messageCacheRepo = nil
		log.Info("Using memory presence repository, caching disabled")
	}
	
	log.Info("Repositories initialized")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiration)
	log.Info("JWT manager initialized")

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)
	roomService := service.NewRoomService(roomRepo)
	messageService := service.NewMessageService(messageRepo, roomRepo, messageCacheRepo)
	presenceService := service.NewPresenceService(presenceRepo, log)
	notificationService := service.NewNotificationService(notificationRepo)
	log.Info("Services initialized")

	// Initialize WebSocket hub and message handler
	hub := websocket.NewHub(nil, presenceService, log) // We'll set the message handler after creating it
	wsMessageHandler := websocket.NewWSMessageHandler(hub, messageService, roomService, presenceService, notificationService, log)
	hub.SetMessageHandler(wsMessageHandler)
	log.Info("WebSocket hub initialized")

	// Start hub in background
	go hub.Run()
	log.Info("WebSocket hub started")

	// Start presence heartbeat monitor
	go presenceService.StartHeartbeatMonitor(context.Background())
	log.Info("Presence heartbeat monitor started")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)
	roomHandler := handler.NewRoomHandler(roomService, authMiddleware, hub)
	messageHandler := handler.NewMessageHandler(messageService, roomService, authMiddleware)
	notificationHandler := handler.NewNotificationHandler(notificationService, authMiddleware)
	userHandler := handler.NewUserHandler(userRepo, roomService, authMiddleware, hub)
	wsHandler := websocket.NewHandler(hub, authService, roomService, presenceService, log)
	log.Info("Handlers initialized")

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Register auth routes
	authHandler.RegisterRoutes(mux)
	
	// Register room routes
	roomHandler.RegisterRoutes(mux)
	
	// Register message routes
	messageHandler.RegisterRoutes(mux)
	
	// Register user routes
	userHandler.RegisterRoutes(mux)
	
	// Register notification routes
	notificationHandler.RegisterRoutes(mux)
	
	// Register WebSocket route
	mux.HandleFunc("/ws", wsHandler.ServeWS)
	log.Info("Routes registered")

	// Setup HTTP server with CORS middleware
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      middleware.CORS(mux),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped")
}
