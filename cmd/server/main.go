package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"realtime-chat-system/config"
	"realtime-chat-system/pkg/database"
	"realtime-chat-system/pkg/logger"

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

	// TODO: Initialize repositories, services, and handlers

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Setup basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
