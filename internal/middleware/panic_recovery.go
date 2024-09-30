package middleware

import (
	"net/http"

	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
)

// PanicRecovery is a middleware that handles panics during request processing and returns an internal server error response
func PanicRecovery(logger logger.Interface) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Errorf("panic occurred: %v", recovered)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", nil)
	})
}
