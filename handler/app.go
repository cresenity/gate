package handler

import (
	"net/http"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

func GetInfo(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data: dtf.AppInfo{
				Name:    config.AppConfig.Name,
				Version: config.AppConfig.Version,
			},
		},
	)
}

func GetConfiguration(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data: dtf.Configuration{
				DefaultIp: "127.0.0.1",
			},
		},
	)
}

func SetDefaultIpAddress(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: "TODO : implement set default ip address",
		},
	)
}
