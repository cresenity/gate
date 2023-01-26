package router

import "github.com/gin-gonic/gin"

func InitializeRouter() (router *gin.Engine) {
	router = gin.Default()
	router.GET("/")
	router.GET("info")

	apiRoute := router.Group("api")
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
