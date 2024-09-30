package middleware

import (
	"net/http"

	appAuth "github.com/berikulyBeket/todo-plus/pkg/app_auth"

	"github.com/gin-gonic/gin"
)

const (
	HeaderAppID  = "appId"
	HeaderAppKey = "appKey"
)

// AppAuth is a middleware that handles authentication based on appId and appKey
func AppAuth(accessType string, appAuth appAuth.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetHeader(HeaderAppID)
		appKey := c.GetHeader(HeaderAppKey)

		if appId == "" || appKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Empty appId or appKey"})
			c.Abort()
			return
		}

		if appAuth.Validate(appId, appKey, accessType) {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid appId or appKey"})
		c.Abort()
	}
}
