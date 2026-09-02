package router

import (
	gincollector "github.com/cresenity/devcloud-collector-client-golang/gin"
	"github.com/cresenity/gate/config"
	"github.com/cresenity/gate/handler"
	"github.com/cresenity/gate/middleware"
	"github.com/gin-gonic/gin"
)

func InitializeRouter() (router *gin.Engine) {
	// gin.New() + Logger() + gincollector.Recovery() alih-alih gin.Default(): sama
	// seperti gin.Recovery() bawaan (memulihkan panic lalu menjawab 500), tapi
	// panic-nya juga dilaporkan ke devcloud - jaring pengaman untuk galat yang
	// lolos jadi panic sungguhan, di luar tiga bekas log.Fatal/log.Panicln yang
	// sudah diganti balasan JSON biasa di handler/domain.go.
	router = gin.New()
	router.Use(gin.Logger(), gincollector.Recovery(config.Collector))
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
