package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func FavProxyMiddleware(favServiceURL string) gin.HandlerFunc {
	fmt.Printf("DEBUG: FavProxyMiddleware initialized with URL: %s\n", favServiceURL)

	return func(c *gin.Context) {
		_, exists := c.Get("user_sso_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Формируем URL для запроса
		var targetURL string
		if strings.HasSuffix(favServiceURL, "/api/v1") {
			// Если favServiceURL уже содержит /api/v1, просто добавляем путь
			relPath := c.Request.URL.Path[7:] // Убираем начальные /api/v1
			targetURL = fmt.Sprintf("%s%s", favServiceURL, relPath)
		} else {
			// Иначе используем полный путь
			targetURL = fmt.Sprintf("%s%s", favServiceURL, c.Request.URL.Path)
		}

		if c.Request.URL.RawQuery != "" {
			targetURL = fmt.Sprintf("%s?%s", targetURL, c.Request.URL.RawQuery)
		}

		fmt.Printf("DEBUG: Proxy target URL: %s\n", targetURL)

		// Создаем новый запрос
		proxyReq, err := http.NewRequestWithContext(
			c.Request.Context(),
			c.Request.Method,
			targetURL,
			nil, // Тело запроса добавим позже
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
			c.Abort()
			return
		}

		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
				c.Abort()
				return
			}
			// Восстанавливаем тело запроса для Gin
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			// Устанавливаем тело для проксируемого запроса
			proxyReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			// Устанавливаем Content-Length
			proxyReq.ContentLength = int64(len(bodyBytes))
			
			fmt.Printf("DEBUG: Request body: %s\n", string(bodyBytes))
		}

		// Копируем заголовки запроса
		for key, values := range c.Request.Header {
			// Пропускаем только заголовок хоста
			if strings.ToLower(key) != "host" {
				for _, value := range values {
					proxyReq.Header.Add(key, value)
				}
			}
		}

		// Убедимся, что заголовок Authorization передается
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			proxyReq.Header.Set("Authorization", authHeader)
			fmt.Printf("DEBUG: Forwarding Authorization header: %s\n", authHeader)
		} else {
			fmt.Printf("WARNING: No Authorization header found in request\n")
		}

		proxyReq.Header.Set("Content-Type", "application/json")

		// Выполняем запрос к микросервису
		client := &http.Client{}
		proxyResp, err := client.Do(proxyReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach favourites microservice"})
			c.Abort()
			return
		}
		defer proxyResp.Body.Close()

		// Копируем тело ответа
		respBody, err := io.ReadAll(proxyResp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response from favourites microservice"})
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: Response from favourites service: status=%d, body=%s\n", proxyResp.StatusCode, string(respBody))

		// Копируем заголовки ответа
		for key, values := range proxyResp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// Явно устанавливаем CORS заголовки
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, X-Requested-With")

		// Устанавливаем статус код ответа
		c.Writer.WriteHeader(proxyResp.StatusCode)

		// Отправляем ответ клиенту
		c.Writer.Write(respBody)
	}
}