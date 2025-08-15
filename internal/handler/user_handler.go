package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"server/internal/proto"
)

type UserHandler struct {
	userClient proto.UserServiceClient
}

func NewUserHandler(userClient proto.UserServiceClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req proto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.Register(context.Background(), &req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.AlreadyExists {
				c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req proto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.Login(context.Background(), &req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.Unauthenticated {
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": resp.Token})
}