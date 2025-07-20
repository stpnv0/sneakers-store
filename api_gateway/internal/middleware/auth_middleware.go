package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "user_sso_id"
)

// AuthMiddleware проверяет JWT токен и устанавливает ID пользователя в контекст запроса
func AuthMiddleware(appSecret string) gin.HandlerFunc {
	if appSecret == "" {
		log.Fatal("app secret cannot be empty")
	}

	return func(c *gin.Context) {
		// Получение токена из заголовка Authorization
		authHeader := c.GetHeader(authorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Проверка формата "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		//верификация с секретом
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(appSecret), nil
		})

		if err != nil {
			log.Printf("Token validation error: %v", err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is expired"})
			} else {
				log.Printf("Detailed token error: %+v", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			}
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Берем uid из токена
		uid, ok := claims["uid"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}

		// Конвертируем uid в число для корректной обработки
		userID, ok := uid.(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID has invalid format"})
			return
		}

		userSSOID := int(userID)

		// Устанавливаем ID пользователя в контекст
		c.Set(userCtx, userSSOID)
		log.Printf("User authenticated successfully. UID: %v", userSSOID)

		c.Next()
	}
}
