package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const managementTokenEnv = "API_AUTH_TOKEN"

// RequireManagementToken protects management endpoints with a static bearer token.
// This is intentionally minimal, but it is still much safer than exposing write and
// operational endpoints without any authentication.
func RequireManagementToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		requiredToken := strings.TrimSpace(os.Getenv(managementTokenEnv))
		if requiredToken == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "management API authentication is not configured",
			})
			return
		}

		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing bearer token",
			})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token != requiredToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid bearer token",
			})
			return
		}

		c.Next()
	}
}
