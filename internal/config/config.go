package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application-specific configuration
type Config struct {
	DBHost          string
	DBUser          string
	DBPassword      string
	DBName          string // Для UserService
	DBPort          string
	JWTSecret       string
	UserServicePort int
	TodoServicePort int // Добавлено это поле
	TodoDBName      string // Добавлено это поле
}

// LoadConfig reads configuration from environment variables or .env file
func LoadConfig() *Config {
	// Try to load .env file, ignore if not found
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from system environment variables")
	}

	userServicePortStr := os.Getenv("USER_SERVICE_PORT")
	if userServicePortStr == "" {
		userServicePortStr = "50051" // Default value
	}
	userServicePort, err := strconv.Atoi(userServicePortStr)
	if err != nil {
		log.Fatalf("Invalid USER_SERVICE_PORT in .env: %v", err)
	}

	todoServicePortStr := os.Getenv("TODO_SERVICE_PORT")
	if todoServicePortStr == "" {
		todoServicePortStr = "50052" // Default value
	}
	todoServicePort, err := strconv.Atoi(todoServicePortStr)
	if err != nil {
		log.Fatalf("Invalid TODO_SERVICE_PORT in .env: %v", err)
	}

	return &Config{
		DBHost:          os.Getenv("DB_HOST"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		DBPort:          os.Getenv("DB_PORT"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		UserServicePort: userServicePort,
		TodoServicePort: todoServicePort, // Инициализация нового поля
		TodoDBName:      os.Getenv("TODO_DB_NAME"), // Инициализация нового поля
	}
}