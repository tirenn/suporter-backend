package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware dynamically checks and applies CORS headers based on configured allowed origins.
func CORSMiddleware(allowedOriginsStr string) gin.HandlerFunc {
	// If empty or "*", allow all origins (default open mode)
	cleanStr := strings.TrimSpace(allowedOriginsStr)
	if cleanStr == "" || cleanStr == "*" {
		return func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Suporter-Key, X-Suporter-Signature, X-Suporter-Timestamp, X-Is-Test")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusOK)
				return
			}
			c.Next()
		}
	}

	// Parse comma-separated list of allowed origins
	originMap := make(map[string]bool)
	for _, o := range strings.Split(cleanStr, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			originMap[trimmed] = true
		}
	}

	return func(c *gin.Context) {
		reqOrigin := c.Request.Header.Get("Origin")

		// Check if request origin is in the allowed whitelist
		if originMap[reqOrigin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if strings.HasPrefix(c.Request.URL.Path, "/overlay/") || strings.HasPrefix(c.Request.URL.Path, "/api/v1/webhooks/") || strings.HasPrefix(c.Request.URL.Path, "/api/v1/donations") {
			// Always allow public OBS browser source widgets and public donation/webhook callbacks
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Suporter-Key, X-Suporter-Signature, X-Suporter-Timestamp, X-Is-Test")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
