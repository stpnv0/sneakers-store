package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен и устанавливает ID пользователя в контекст запроса
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение токена из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Проверка формата "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Разбиение токена на части (header.payload.signature)
		tokenParts := strings.Split(tokenString, ".")
		if len(tokenParts) != 3 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Декодирование payload из base64
		payloadJSON, err := decodeBase64Segment(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
			c.Abort()
			return
		}

		// Разбор JSON-данных из payload
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payloadMap); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload format"})
			c.Abort()
			return
		}

		// Поиск поля uid в токене
		uidValue, exists := payloadMap["uid"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in token",
			})
			c.Abort()
			return
		}

		// Преобразование uid в int
		userID, ok := uidValue.(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID has invalid format"})
			c.Abort()
			return
		}

		// Конвертация float64 в int
		userSSOID := int(userID)

		// Сохранение ID пользователя в контексте запроса
		c.Set("user_sso_id", userSSOID)
		c.Next()
	}
}

// Декодирование сегмента base64url в байты
func decodeBase64Segment(seg string) ([]byte, error) {
	// Выравнивание строки до длины, кратной 4
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	// Замена URL-безопасных символов на стандартные для base64
	seg = strings.ReplaceAll(seg, "-", "+")
	seg = strings.ReplaceAll(seg, "_", "/")

	return base64.StdEncoding.DecodeString(seg)
}
