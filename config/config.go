package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Presence     PresenceConfig
	JWT          JWTConfig
	Kafka        KafkaConfig
	LoggerConfig LoggerConfig
}

type KafkaConfig struct {
	Brokers        string
	RequestTimeout time.Duration
	RetryBackoff   time.Duration
	MaxRetries     int
	GroupID        string

	AutoCreateTopics bool

	BatchSize    int
	BatchBytes   int
	BatchTimeout int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Name            string
	Version         string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Timeout  time.Duration
}

// PresenceConfig holds presence service configuration
type PresenceConfig struct {
	Backend string // "redis" or "memory"
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type RotationConfig struct {
	MaxSize    int64 // Maximum size in bytes before rotation (default: 100MB)
	MaxAge     int   // Maximum number of days to retain old logs (default: 30)
	MaxBackups int   // Maximum number of backup files to keep (default: 10)
	Compress   bool  // Whether to compress rotated files (default: true)
}
type LogOutputConfig struct {
	Path    string
	Console bool
	File    bool
}
type LoggerConfig struct {
	Summary  LogOutputConfig
	Detail   LogOutputConfig
	Rotation RotationConfig
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Name:            getEnv("SERVER_NAME", "Real-time Chat System"),
			Version:         Version,
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "chat_system"),
			Timeout:  getDurationEnv("MONGODB_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
			Timeout:  getDurationEnv("REDIS_TIMEOUT", 5*time.Second),
		},
		Presence: PresenceConfig{
			Backend: getEnv("PRESENCE_BACKEND", "memory"), // "redis" or "memory"
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
		Kafka: KafkaConfig{
			Brokers:          getEnv("KAFKA_BROKERS", "localhost:29092"),
			RequestTimeout:   getDurationEnv("KAFKA_REQUEST_TIMEOUT", 10*time.Second),
			RetryBackoff:     getDurationEnv("KAFKA_RETRY_BACKOFF", 500*time.Millisecond),
			MaxRetries:       getIntEnv("KAFKA_MAX_RETRIES", 5),
			GroupID:          getEnv("KAFKA_GROUP_ID", "chat_system_group"),
			AutoCreateTopics: getEnv("KAFKA_AUTO_CREATE_TOPICS", "true") == "true",
			BatchSize:        getIntEnv("KAFKA_BATCH_SIZE", 100),
			BatchBytes:       getIntEnv("KAFKA_BATCH_BYTES", 1048576), // 1 MB
			BatchTimeout:     getIntEnv("KAFKA_BATCH_TIMEOUT_MS", 10), // in milliseconds
		},
		LoggerConfig: LoggerConfig{
			Summary: LogOutputConfig{
				Path:    getEnv("LOG_SUMMARY_PATH", "./logs/summary/"),
				Console: getEnv("LOG_SUMMARY_CONSOLE", "true") == "true",
				File:    getEnv("LOG_SUMMARY_FILE", "true") == "true",
			},
			Detail: LogOutputConfig{
				Path:    getEnv("LOG_DETAIL_PATH", "./logs/detail/"),
				Console: getEnv("LOG_DETAIL_CONSOLE", "true") == "true",
				File:    getEnv("LOG_DETAIL_FILE", "true") == "true",
			},
			Rotation: RotationConfig{
				// MaxSize:    50 * 1024 * 1024, // 50MB
				MaxSize:   int64(getIntEnv("LOG_ROTATION_MAX_SIZE_MB", 100)) * 1024 * 1024, // in bytes
				MaxAge:     7,                // 7 days
				MaxBackups: getIntEnv("LOG_ROTATION_MAX_BACKUPS", 10),
				Compress:   getEnv("LOG_ROTATION_COMPRESS", "true") == "true",
			},
		},
	}

	if os.Getenv("KAFKA_BROKERS") != "" {
		kc := KafkaConfig{
			Brokers:          os.Getenv("KAFKA_BROKERS"),
			RequestTimeout:   getDurationEnv("KAFKA_REQUEST_TIMEOUT", 10*time.Second),
			RetryBackoff:     getDurationEnv("KAFKA_RETRY_BACKOFF", 500*time.Millisecond),
			MaxRetries:       getIntEnv("KAFKA_MAX_RETRIES", 5),
			GroupID:          getEnv("KAFKA_GROUP_ID", "chat_system_group"),
			AutoCreateTopics: getEnv("KAFKA_AUTO_CREATE_TOPICS", "true") == "true",
			BatchSize:        getIntEnv("KAFKA_BATCH_SIZE", 100),
			BatchBytes:       getIntEnv("KAFKA_BATCH_BYTES", 1048576), // 1 MB
			BatchTimeout:     getIntEnv("KAFKA_BATCH_TIMEOUT_MS", 10), // in milliseconds
		}
		cfg.Kafka = kc
	}

	return cfg
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv gets duration from environment variable or returns default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getIntEnv gets integer from environment variable or returns default
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
