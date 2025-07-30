package proxy

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ProxyHandler(targetURL string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("некорректный URL %s: %v", targetURL, err)
	}

	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(target)

		if userID, exists := c.Get("user_sso_id"); exists {
			c.Request.Header.Set("X-User-ID", fmt.Sprint(userID))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
