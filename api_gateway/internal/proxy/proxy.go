package proxy

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// Handler возвращает Gin-хендлер, проксирующий запросы на targetURL
func Handler(targetURL string) (gin.HandlerFunc, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("proxy: invalid URL %s: %w", targetURL, err)
	}

	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(target)

		if userID, exists := c.Get("user_sso_id"); exists {
			c.Request.Header.Set("X-User-ID", fmt.Sprint(userID))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}
