package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
)

// CORS dibangun sekali sebagai gin.HandlerFunc dan dipakai langsung sebagai middleware.
// Sebelumnya cors.New(...) dipanggil di dalam sebuah wrapper yang membuang nilai baliknya
// (BUG-03), sehingga tidak ada header CORS yang pernah terkirim.
var CORS = cors.New(cors.Config{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	AllowHeaders:     []string{"Content-Type", "Content-Length", "Authorization", "Origin"},
	ExposeHeaders:    []string{"Content-Type", "Content-Length"},
	AllowCredentials: true,
	AllowWebSockets:  true,
	MaxAge:           12 * time.Hour,
})
