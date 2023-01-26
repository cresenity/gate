package router

import (
	"github.com/cresenity/gate/middleware"
	"github.com/gin-gonic/gin"
)

func InitializeRouter() (router *gin.Engine) {
	router = gin.Default()
	router.GET("/")

	apiRoute := router.Group("api")
	apiRoute.Use(
		middleware.Auth,
		middleware.CORS,
	)

	apiRoute.GET("info")
	configRoute := apiRoute.Group("config")
	{
		configRoute.GET("")
		configRoute.POST("ip/:ipAddress")
	}

	sslRoute := apiRoute.Group("ssl")
	{
		sslRoute.POST("install/:domain/:ipAddress")
		sslRoute.PUT("update/:domain/:ipAddress")
		sslRoute.GET("status/:domain")
		sslRoute.GET("status/all")
	}

	return
}
