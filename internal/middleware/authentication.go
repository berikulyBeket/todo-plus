package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	UserIdCtx           = "userId"
)

// Authentication is a middleware function that handles authentication by parsing a JWT token from the Authorization header
func Authentication(authUseCase usecase.Auth, logger logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			logger.Error("empty auth header")
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
				"token": "empty auth header",
			})
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			logger.Error("invalid auth header")
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
				"token": "invalid auth header",
			})
			return
		}

		userId, err := authUseCase.ParseToken(c.Request.Context(), headerParts[1])
		if err != nil {
			logger.Error("invalid auth header")
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
				"token": err.Error(),
			})
			return
		}

		c.Set(UserIdCtx, userId)
		c.Next()
	}
}

// GetUserId retrieves the user ID from the context
func GetUserId(c *gin.Context) (int, error) {
	userId, ok := c.Get(UserIdCtx)
	if !ok {
		return 0, errors.New("user id not found")
	}

	userIdInt, ok := userId.(int)
	if !ok {
		return 0, errors.New("invalid userId")
	}

	return userIdInt, nil
}
