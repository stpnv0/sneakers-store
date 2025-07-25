package middleware

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ExtractUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			log.Printf("ERROR: X-User-ID header not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("ERROR: Invalid X-User-ID format: %s", userIDStr)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		c.Set("user_sso_id", userID)
		log.Printf("DEBUG: Set user_sso_id in context: %d", userID)
		c.Next()
	}
}
