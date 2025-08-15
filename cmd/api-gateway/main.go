package main

import (
	"fmt"
	"log"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"server/internal/config"
	"server/internal/handler"
	"server/internal/middleware"
	"server/internal/proto"
)

func main() {
	cfg := config.LoadConfig()

	// Настройка gRPC-клиентов для микросервисов
	// Клиент для UserService
	userServiceAddr := fmt.Sprintf("localhost:%d", cfg.UserServicePort)
	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to user service: %v", err)
	}
	defer userConn.Close()
	userClient := proto.NewUserServiceClient(userConn)

	// Клиент для TodoService
	todoServiceAddr := fmt.Sprintf("localhost:%d", cfg.TodoServicePort)
	todoConn, err := grpc.Dial(todoServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to todo service: %v", err)
	}
	defer todoConn.Close()
	todoClient := proto.NewTodoServiceClient(todoConn)

	// Инициализация Gin-роутера
	router := gin.Default()

	// Инициализация хэндлеров
	userHandler := handler.NewUserHandler(userClient)
	todoHandler := handler.NewTodoHandler(todoClient)

	// Маршруты без аутентификации
	router.POST("/api/register", userHandler.Register)
	router.POST("/api/login", userHandler.Login)

	// Маршруты, требующие аутентификации
	authGroup := router.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(userClient))
	{
		// Маршруты для TodoService
		authGroup.POST("/todos", todoHandler.CreateTodo)
		authGroup.GET("/todos", todoHandler.GetTodos)
		authGroup.PUT("/todos/:id", todoHandler.UpdateTodo)
		authGroup.DELETE("/todos/:id", todoHandler.DeleteTodo)
	}

	// Запуск REST-сервера
	log.Println("API Gateway listening on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}