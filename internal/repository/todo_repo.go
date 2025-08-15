package repository

import (
	"context"
	"gorm.io/gorm"
	"server/internal/models"
)

type TodoRepository interface {
	CreateTodo(ctx context.Context, todo *models.Todo) error
	GetTodosByUserID(ctx context.Context, userID uint) ([]*models.Todo, error)
	GetTodoByID(ctx context.Context, id uint) (*models.Todo, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) error
	DeleteTodo(ctx context.Context, id uint) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) CreateTodo(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepository) GetTodosByUserID(ctx context.Context, userID uint) ([]*models.Todo, error) {
	var todos []*models.Todo
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepository) GetTodoByID(ctx context.Context, id uint) (*models.Todo, error) {
	var todo models.Todo
	if err := r.db.WithContext(ctx).First(&todo, id).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

func (r *todoRepository) DeleteTodo(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Todo{}, id).Error
}