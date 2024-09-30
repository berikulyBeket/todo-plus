package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/utils"
)

// SignUp handles user registration.
// @Summary Create a new user
// @Description Register a new user by providing name, username, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "User registered successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 500 {object} utils.ErrorResponse "Failed to create user"
// @Router /auth/sign-up [post]
func (h *Handler) SignUp(c *gin.Context) {
	var input entity.User

	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Errorf("validation failed: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]string{
			"body": "Invalid or malformed JSON",
		})
		return
	}

	userId, err := h.Usecases.Auth.CreateUser(c.Request.Context(), input)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"username": input.Username,
		}).Errorf("failed to create user: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to create user", map[string]string{
			"database": err.Error(),
		})
		return
	}

	h.Metrics.IncrementSignedUpUsers()
	h.Logger.WithFields(map[string]interface{}{
		"user_id":  userId,
		"username": input.Username,
	}).Info("user signed up successfully")

	utils.NewSuccessResponse(c, http.StatusOK, "User registered successfully", map[string]interface{}{
		"id": userId,
	})
}

// signInInput represents the expected input for user authentication.
type signInInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignIn handles user authentication.
// @Summary Sign in a user
// @Description Log in with username and password to receive a token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body signInInput true "User credentials"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "Signed in successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Invalid username or password"
// @Failure 500 {object} utils.ErrorResponse "Failed to retrieve user"
// @Router /auth/sign-in [post]
func (h *Handler) SignIn(c *gin.Context) {
	var input signInInput

	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Errorf("validation failed: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]string{
			"body": "Invalid or malformed JSON",
		})
		return
	}

	user, err := h.Usecases.Auth.AuthenticateUser(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		if err == utils.ErrUserNotFound {
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Invalid username or password", map[string]string{
				"credentials": "Invalid username or password",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"username": input.Username,
		}).Errorf("failed to retrieve user: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user", map[string]string{
			"database": err.Error(),
		})
		return
	}

	token, err := h.Usecases.Auth.GenerateToken(c.Request.Context(), user.Id)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":  user.Id,
			"username": input.Username,
		}).Errorf("failed to generate token: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to generate token", map[string]string{
			"token": err.Error(),
		})
		return
	}

	h.Metrics.IncrementSignedInUsers()
	h.Logger.WithFields(map[string]interface{}{
		"user_id":  user.Id,
		"username": input.Username,
	}).Info("user signed in successfully")

	utils.NewSuccessResponse(c, http.StatusOK, "Signed in successfully", map[string]interface{}{
		"token": token,
	})
}
