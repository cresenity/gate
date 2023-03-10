package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cresenity/gate/config"
	"github.com/cresenity/gate/router"
	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadAppConfig()

	if !config.AppConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.AppConfig.Port),
		Handler:        router.InitializeRouter(),
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
