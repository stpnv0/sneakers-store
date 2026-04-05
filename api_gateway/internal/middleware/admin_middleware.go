package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminChecker interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

func AdminMiddleware(checker AdminChecker, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserIDFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		isAdmin, err := checker.IsAdmin(c.Request.Context(), userID)
		if err != nil {
			log.Error("failed to check admin status",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify permissions"})
			return
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		c.Next()
	}
}
