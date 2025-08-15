package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"server/internal/proto"
)

func AuthMiddleware(userClient proto.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(401, gin.H{"error": "Token is missing"})
			c.Abort()
			return
		}

		// Вызов gRPC-сервиса для валидации токена
		resp, err := userClient.ValidateToken(context.Background(), &proto.ValidateTokenRequest{Token: token})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Unauthenticated {
					c.JSON(401, gin.H{"error": "Invalid or expired token"})
					c.Abort()
					return
				}
			}
			c.JSON(500, gin.H{"error": "Failed to validate token"})
			c.Abort()
			return
		}

		if !resp.IsValid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Если токен валиден, сохраняем user_id и role в контексте Gin
		c.Set("user_id", resp.UserId)
		c.Set("user_role", resp.Role)

		c.Next()
	}
}