package v1

import (
	"net/http"
	"strconv"

	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
)

// DeleteUserByAdmin godoc
// @Summary Delete a user
// @Description Deletes a user by their userId
// @Tags user
// @Param id path int true "User ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "User deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid userId param"
// @Failure 404 {object} utils.ErrorResponse "User not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to delete user"
// @Router /private/api/users/{id} [delete]
func (h *Handler) DeleteUserByAdmin(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.Errorf("invalid userId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid userId param", map[string]string{
			"param": "userId must be a valid integer",
		})
		return
	}

	err = h.Usecases.User.DeleteOneByAdmin(c.Request.Context(), userId)
	if err != nil {
		if err == utils.ErrUserNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "User not found", map[string]string{
				"userId": "The requested user does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("failed to delete user: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", map[string]string{
			"database": "Error during user deletion",
		})
		return
	}

	h.Logger.WithFields(map[string]interface{}{
		"user_id": userId,
	}).Info("user deleted successfully")
	h.Metrics.IncrementDeletedUsers()

	utils.NewSuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
