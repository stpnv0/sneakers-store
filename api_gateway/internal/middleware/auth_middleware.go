package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "user_sso_id"
)

func AuthMiddleware(appSecret string, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(appSecret), nil
		})
		if err != nil {
			log.Warn("token validation error", slog.String("error", err.Error()))
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token is expired"})
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			}
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		uid, ok := claims["uid"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in token"})
			return
		}

		userID, ok := uid.(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user ID has invalid format"})
			return
		}

		c.Set(userCtx, int64(userID))
		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) (int64, error) {
	val, exists := c.Get(userCtx)
	if !exists {
		return 0, fmt.Errorf("user not authenticated")
	}
	uid, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("user ID has invalid type")
	}
	return uid, nil
}
