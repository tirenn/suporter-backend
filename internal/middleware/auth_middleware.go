package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/service"
)

const UserIDKey = "user_id"
const UserEmailKey = "user_email"

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)

		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := val.(uint64)
	return userID, ok
}
