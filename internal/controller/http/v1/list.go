package v1

import (
	"net/http"
	"strconv"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/middleware"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
)

// createList godoc
// @Summary Create a new list
// @Description Create a new list for the authenticated user
// @Tags lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body entity.List true "List data"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 201 {object} utils.SuccessResponse "List created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to create list"
// @Router /api/lists/ [post]
func (h *Handler) CreateList(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	var input entity.List
	if err := c.BindJSON(&input); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("failed to bind JSON: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid input", map[string]string{
			"body": "Invalid or malformed JSON",
		})
		return
	}

	listId, err := h.Usecases.List.Create(c.Request.Context(), userId, &input)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("failed to create list: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to create list", map[string]string{
			"database": "Error during list creation",
		})
		return
	}

	h.Metrics.IncrementCreatedLists()

	utils.NewSuccessResponse(c, http.StatusCreated, "List created successfully", map[string]interface{}{
		"id": listId,
	})
}

// getAllLists godoc
// @Summary Get all lists
// @Description Retrieve all lists for the authenticated user
// @Tags lists
// @Security BearerAuth
// @Produce json
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.List} "Lists retrieved successfully"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to retrieve lists"
// @Router /api/lists/ [get]
func (h *Handler) GetAllLists(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	lists, err := h.Usecases.List.GetAll(c.Request.Context(), userId)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("failed to retrieve lists: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve lists", map[string]string{
			"database": "Error during list retrieval",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "Lists retrieved successfully", lists)
}

// getListById godoc
// @Summary Get a list by ID
// @Description Retrieve a specific list by ID for the authenticated user
// @Tags lists
// @Security BearerAuth
// @Produce json
// @Param id path int true "List ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=entity.List} "List retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid listId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "List not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to retrieve list"
// @Router /api/lists/{id} [get]
func (h *Handler) GetListById(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid listId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	list, err := h.Usecases.List.GetOneById(c.Request.Context(), userId, listId)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrListNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "List not found", map[string]string{
				"listId": "The requested list does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to retrieve list: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve list", map[string]string{
			"database": "Error during list retrieval",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "List retrieved successfully", list)
}

// updateList godoc
// @Summary Update a list
// @Description Update a specific list by ID for the authenticated user
// @Tags lists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param input body entity.UpdateListInput true "Updated list data"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "List updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input or listId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "List not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to update list"
// @Router /api/lists/{id} [put]
func (h *Handler) UpdateList(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid listId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	var newTodoInput entity.UpdateListInput
	if err := c.BindJSON(&newTodoInput); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to bind JSON: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid input", map[string]string{
			"body": "Invalid or malformed JSON",
		})

		return
	}

	if err := newTodoInput.Validate(); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Warn("validation failed")

		utils.NewErrorResponse(c, http.StatusBadRequest, "Validation failed", map[string]string{
			"validation": err.Error(),
		})

		return
	}

	err = h.Usecases.List.UpdateOneById(c.Request.Context(), userId, listId, newTodoInput)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrListNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "List not found", map[string]string{
				"listId": "The requested list does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to update list: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to update list", map[string]string{
			"database": "Error during list update",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "List updated successfully", nil)
}

// deleteList godoc
// @Summary Delete a list
// @Description Delete a specific list by Id for the authenticated user
// @Tags lists
// @Security BearerAuth
// @Param id path int true "List ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "List deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid listId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "List not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to delete list"
// @Router /api/lists/{id} [delete]
func (h *Handler) DeleteList(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid listId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	err = h.Usecases.List.DeleteOneById(c.Request.Context(), userId, listId)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrListNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "List not found", map[string]string{
				"listId": "The requested list does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to delete list: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to delete list", map[string]string{
			"database": "Error during list deletion",
		})
		return
	}

	h.Metrics.IncrementDeletedLists()

	utils.NewSuccessResponse(c, http.StatusOK, "List deleted successfully", nil)
}

// deleteListByAdmin godoc
// @Summary Delete a list by admin
// @Description Delete a specific list by Id by admin
// @Tags lists
// @Param id path int true "List ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "List deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid listId param"
// @Failure 404 {object} utils.ErrorResponse "List not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to delete list"
// @Router /private/api/lists/{id} [delete]
func (h *Handler) DeleteListByAdmin(c *gin.Context) {
	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.Errorf("invalid listId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	err = h.Usecases.List.DeleteOneByAdmin(c.Request.Context(), listId)
	if err != nil {
		if err == utils.ErrListNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "List not found", map[string]string{
				"listId": "The requested list does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"list_id": listId,
		}).Errorf("failed to delete list: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to delete list", map[string]string{
			"database": "Error during list deletion",
		})
		return
	}

	h.Metrics.IncrementDeletedLists()

	utils.NewSuccessResponse(c, http.StatusOK, "List deleted successfully", nil)
}

// searchLists godoc
// @Summary Search lists
// @Description Search for lists of the authenticated user by a query parameter
// @Tags lists
// @Security BearerAuth
// @Produce json
// @Param query query string true "Search query"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.List} "Lists searched successfully"
// @Failure 400 {object} utils.ErrorResponse "Empty query parameter"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to search lists"
// @Router /api/lists/search [get]
func (h *Handler) SearchLists(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	searchText, err := utils.ParseRequiredQueryParam(c, "search_text")
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("empty searchText param: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Empty searchText param", map[string]string{
			"param": "Query parameter 'search_text' is missing or empty",
		})
		return
	}

	lists, err := h.Usecases.List.Search(c.Request.Context(), userId, searchText)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":     userId,
			"search_text": searchText,
		}).Errorf("failed to search lists: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to search lists", map[string]string{
			"database": "Error during list search",
		})
		return
	}

	h.Metrics.IncrementSearchedLists()

	utils.NewSuccessResponse(c, http.StatusOK, "Lists searched successfully", lists)
}
