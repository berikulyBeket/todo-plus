package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns the CORS middleware with the provided configuration.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = allowedOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "AppId", "AppKey"}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length", "Authorization"}

	return cors.New(config)
}
