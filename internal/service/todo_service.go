package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"server/internal/models"
	"server/internal/proto"
	"server/internal/repository"
)

type TodoServiceServer struct {
	proto.UnimplementedTodoServiceServer
	todoRepo repository.TodoRepository
	userClient proto.UserServiceClient // Клиент для gRPC-сервиса User
}

func NewTodoServiceServer(todoRepo repository.TodoRepository, userClient proto.UserServiceClient) *TodoServiceServer {
	return &TodoServiceServer{
		todoRepo: todoRepo,
		userClient: userClient,
	}
}

func (s *TodoServiceServer) CreateTodo(ctx context.Context, req *proto.CreateTodoRequest) (*proto.TodoItem, error) {
	// В реальном проекте здесь будет проверка токена.
	// пока будем считать user_id переданным.
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format")
	}

	todo := &models.Todo{
		UserID: uint(userID),
		Title:  req.Title,
	}

	if err := s.todoRepo.CreateTodo(ctx, todo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create todo: %v", err)
	}

	return &proto.TodoItem{
		Id:     fmt.Sprintf("%d", todo.ID),
		UserId: fmt.Sprintf("%d", todo.UserID),
		Title:  todo.Title,
		Completed: todo.Completed,
	}, nil
}

func (s *TodoServiceServer) GetTodos(ctx context.Context, req *proto.GetTodosRequest) (*proto.GetTodosResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format")
	}

	todos, err := s.todoRepo.GetTodosByUserID(ctx, uint(userID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get todos: %v", err)
	}

	var todoItems []*proto.TodoItem
	for _, todo := range todos {
		todoItems = append(todoItems, &proto.TodoItem{
			Id:     fmt.Sprintf("%d", todo.ID),
			UserId: fmt.Sprintf("%d", todo.UserID),
			Title:  todo.Title,
			Completed: todo.Completed,
		})
	}

	return &proto.GetTodosResponse{Todos: todoItems}, nil
}

func (s *TodoServiceServer) UpdateTodo(ctx context.Context, req *proto.UpdateTodoRequest) (*proto.TodoItem, error) {
	todoID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid todo ID format")
	}
	
	todo, err := s.todoRepo.GetTodoByID(ctx, uint(todoID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "todo not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get todo: %v", err)
	}

	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format")
	}
	if todo.UserID != uint(userID) {
		return nil, status.Errorf(codes.PermissionDenied, "you don't have permission to update this todo")
	}

	todo.Title = req.Title
	todo.Completed = req.Completed

	if err := s.todoRepo.UpdateTodo(ctx, todo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update todo: %v", err)
	}

	return &proto.TodoItem{
		Id:     fmt.Sprintf("%d", todo.ID),
		UserId: fmt.Sprintf("%d", todo.UserID),
		Title:  todo.Title,
		Completed: todo.Completed,
	}, nil
}

func (s *TodoServiceServer) DeleteTodo(ctx context.Context, req *proto.DeleteTodoRequest) (*proto.DeleteTodoResponse, error) {
	todoID, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid todo ID format")
	}
	
	todo, err := s.todoRepo.GetTodoByID(ctx, uint(todoID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "todo not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get todo: %v", err)
	}
	
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format")
	}
	if todo.UserID != uint(userID) {
		return nil, status.Errorf(codes.PermissionDenied, "you don't have permission to delete this todo")
	}

	if err := s.todoRepo.DeleteTodo(ctx, todo.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete todo: %v", err)
	}

	return &proto.DeleteTodoResponse{Message: "Todo deleted successfully"}, nil
}