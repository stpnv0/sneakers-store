package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

type requestIDCtxKey struct{}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-ID", rid)

		// Сохраняем request_id в стандартном контексте Go,
		// чтобы он был доступен через c.Request.Context() в gRPC-клиентах.
		ctx := context.WithValue(c.Request.Context(), requestIDCtxKey{}, rid)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) string {
	rid, _ := ctx.Value(requestIDCtxKey{}).(string)
	return rid
}

func SlogAccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		rid, _ := c.Get(RequestIDKey)
		ridStr, _ := rid.(string)

		log.Info("http request",
			slog.String("request_id", ridStr),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func SlogRecovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				rid, _ := c.Get(RequestIDKey)
				ridStr := ""
				if rid != nil {
					ridStr = rid.(string)
				}
				log.Error("panic recovered",
					slog.String("request_id", ridStr),
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}
