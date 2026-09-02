package router

import (
	"github.com/cresenity/gate/handler"
	"github.com/cresenity/gate/middleware"
	"github.com/gin-gonic/gin"
)

func InitializeRouter() (router *gin.Engine) {
	router = gin.Default()
	// CORS harus berjalan sebelum Auth (dan sebelum grup api) supaya preflight OPTIONS
	// tanpa Authorization tidak ditolak lebih dulu (BUG-07), dan supaya jalur yang tidak
	// cocok rute apa pun tetap mendapat header CORS lewat rantai NoRoute bawaan gin.
	router.Use(middleware.CORS)

	apiRoute := router.Group("api")
	apiRoute.Use(middleware.Auth)

	apiRoute.GET("info", handler.GetInfo)
	configRoute := apiRoute.Group("config")
	{
		configRoute.GET("", handler.GetConfiguration)
		configRoute.POST("ip/:ipAddress", handler.SetDefaultIpAddress)
	}

	sslRoute := apiRoute.Group("domain")
	{
		sslRoute.POST(":domain/:ipAddress", handler.InstallSsl)
		sslRoute.PUT(":domain/:ipAddress", handler.UpdateDomain)
		sslRoute.DELETE(":domain", handler.DeleteDomain)
		sslRoute.GET("status/:domain", handler.GetDomainStatus)
		sslRoute.GET("status/all", handler.GetAllDomainStatus)
	}

	return
}
