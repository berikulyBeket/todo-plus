package utils

import (
	"github.com/gin-gonic/gin"
)

const (
	statusOk    = "ok"
	statusError = "error"
)

// Error response structure
type ErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Success response structure
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Sends an error response
func NewErrorResponse(c *gin.Context, statusCode int, message string, errors map[string]string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Status:  statusError,
		Message: message,
		Errors:  errors,
	})
}

// Sends a success response
func NewSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  statusOk,
		Message: message,
		Data:    data,
	})
}
