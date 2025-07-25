package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProxyHandler создает обработчик для проксирования запросов к целевому сервису
func ProxyHandler(targetURL, stripPrefix string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("некорректный URL %s: %v", targetURL, err)
	}

	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		// Настраиваем прокси
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.Host = target.Host

				// Обрабатываем путь запроса
				if stripPrefix != "" {
					// Удаляем префикс и очищаем путь
					req.URL.Path = path.Join(target.Path, strings.TrimPrefix(req.URL.Path, stripPrefix))
				}

				// Добавляем ID пользователя из контекста
				if userID, exists := c.Get("user_sso_id"); exists {
					req.Header.Set("X-User-ID", fmt.Sprint(userID))
				}
			},
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
