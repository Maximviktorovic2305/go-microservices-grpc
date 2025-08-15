package main

import (
	"fmt"
	"log"
	"net"
	"server/internal/config"         
	"server/internal/models"
	"server/internal/proto"
	"server/internal/repository"
	"server/internal/service"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Подключение к базе данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Автоматическая миграция
	db.AutoMigrate(&models.User{})
	log.Println("Database migration completed")

	// 3. Инициализация репозитория и сервиса
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserServiceServer(userRepo, cfg.JWTSecret)

	// 4. Запуск gRPC-сервера
	port := fmt.Sprintf(":%d", cfg.UserServicePort)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, userService)

	log.Printf("UserService listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}