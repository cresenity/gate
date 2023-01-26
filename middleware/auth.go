package middleware

import (
	"net/http"
	"strings"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	if token == "" {
		token, _ = c.GetQuery("apiKey")
	}

	if token != config.AppConfig.ApiKey {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dtf.Response{
			Message: "Authentication failed, invalid API KEY",
		})
		return
	}

	c.Next()
}
