package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	token := ""
	if strings.HasPrefix(header, "Bearer ") {
		token = strings.TrimPrefix(header, "Bearer ")
	} else if header == "" {
		token, _ = c.GetQuery("apiKey")
	}

	valid := config.AppConfig.ApiKey != "" &&
		subtle.ConstantTimeCompare([]byte(token), []byte(config.AppConfig.ApiKey)) == 1

	if !valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dtf.Response{
			Message: "Authentication failed, invalid API KEY",
		})
		return
	}

	c.Next()
}
