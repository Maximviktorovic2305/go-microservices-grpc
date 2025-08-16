package main

import (
	"fmt"
	"log"
	"net"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"server/internal/config"
	"server/internal/models"
	"server/internal/proto"
	"server/internal/repository"
	"server/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.TodoDBName,
		cfg.DBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	db.AutoMigrate(&models.Todo{})
	log.Println("Database migration for TodoService completed")
	
	userServiceAddr := fmt.Sprintf("localhost:%d", cfg.UserServicePort)
	conn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to user service: %v", err)
	}
	defer conn.Close()
	userClient := proto.NewUserServiceClient(conn)

	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoServiceServer(todoRepo, userClient)

	todoPort := fmt.Sprintf(":%d", cfg.TodoServicePort)
	lis, err := net.Listen("tcp", todoPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterTodoServiceServer(grpcServer, todoService)

	log.Printf("TodoService listening on port %s", todoPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}