package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"server/internal/proto"
)

type TodoHandler struct {
	todoClient proto.TodoServiceClient
}

func NewTodoHandler(todoClient proto.TodoServiceClient) *TodoHandler {
	return &TodoHandler{todoClient: todoClient}
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var req proto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserId = userID.(string)

	resp, err := h.todoClient.CreateTodo(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *TodoHandler) GetTodos(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	req := &proto.GetTodosRequest{UserId: userID.(string)}

	resp, err := h.todoClient.GetTodos(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get todos"})
		return
	}

	c.JSON(http.StatusOK, resp.Todos)
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	todoID := c.Param("id")
	var req proto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = todoID
	req.UserId = userID.(string)

	resp, err := h.todoClient.UpdateTodo(context.Background(), &req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.PermissionDenied {
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
				return
			}
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	todoID := c.Param("id")
	req := &proto.DeleteTodoRequest{Id: todoID, UserId: userID.(string)}

	_, err := h.todoClient.DeleteTodo(context.Background(), req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.PermissionDenied {
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
				return
			}
			if st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	c.Status(http.StatusNoContent)
}