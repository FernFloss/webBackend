package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DB         DBConfig
	RabbitMQ   RabbitMQConfig
	QueueName  string
	GinMode    string
	ServerPort string // HTTP server port
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL string
}

// GetDSN returns the PostgreSQL connection string
func (c *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

// LoadConfig loads configuration from environment variables
// It also tries to load from .env file if it exists
func LoadConfig() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	config := &Config{}

	// Load database configuration
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	config.DB = DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "auditorium_db"),
	}

	// Load RabbitMQ configuration
	config.RabbitMQ = RabbitMQConfig{
		URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}

	// Load queue name
	config.QueueName = getEnv("QUEUE_NAME", "camera_events")

	// Load Gin mode
	config.GinMode = getEnv("GIN_MODE", "debug")

	// Load server port
	config.ServerPort = getEnv("SERVER_PORT", "8080")

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
