package v1

import (
	"net/http"
	"strconv"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/middleware"
	"github.com/berikulyBeket/todo-plus/utils"

	"github.com/gin-gonic/gin"
)

// createItem godoc
// @Summary Create a new item
// @Description Create a new item in a list for the authenticated user
// @Tags items
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Param input body entity.Item true "Item data"
// @Success 201 {object} utils.SuccessResponse "Item created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input or listId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to create item"
// @Router /api/lists/{id}/items/ [post]
func (h *Handler) CreateItem(c *gin.Context) {
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

	var input entity.Item
	if err := c.BindJSON(&input); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to bind JSON: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid input", map[string]string{
			"body": "Invalid or malformed JSON",
		})
		return
	}

	if _, err := h.Usecases.List.GetOneById(c.Request.Context(), userId, listId); err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, "Validation failed", map[string]string{
			"validation": err.Error(),
		})
		return
	}

	itemId, err := h.Usecases.Item.Create(c.Request.Context(), userId, listId, &input)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to create item: %s", err)
		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to create item", map[string]string{
			"database": "Error during item creation",
		})
		return
	}

	h.Metrics.IncrementCreatedItems()

	utils.NewSuccessResponse(c, http.StatusCreated, "Item created successfully", map[string]interface{}{
		"id": itemId,
	})
}

// getAllItems godoc
// @Summary Get all items in a list
// @Description Retrieve all items in a specific list for the authenticated user
// @Tags items
// @Security BearerAuth
// @Produce json
// @Param id path int true "List ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.Item} "Items retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid listId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to retrieve items"
// @Router /api/lists/{id}/items/ [get]
func (h *Handler) GetAllItems(c *gin.Context) {
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

	if _, err := h.Usecases.List.GetOneById(c.Request.Context(), userId, listId); err != nil {
		utils.NewErrorResponse(c, http.StatusBadRequest, "Validation failed", map[string]string{
			"validation": err.Error(),
		})
		return
	}

	items, err := h.Usecases.Item.GetAll(c.Request.Context(), userId, listId)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"list_id": listId,
		}).Errorf("failed to retrieve items: %s", err)
		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve items", map[string]string{
			"database": "Error during item retrieval",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "Items retrieved successfully", items)
}

// getItemById godoc
// @Summary Get an item by ID
// @Description Retrieve a specific item by ID for the authenticated user
// @Tags items
// @Security BearerAuth
// @Produce json
// @Param id path int true "Item ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=entity.Item} "Item retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid itemId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "Item not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to retrieve item"
// @Router /api/items/{id} [get]
func (h *Handler) GetItemById(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid itemId parameter: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid itemId param", map[string]string{
			"param": "itemId must be a valid integer",
		})
		return
	}

	item, err := h.Usecases.Item.GetOneById(c.Request.Context(), userId, itemId)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrItemNotFound {
			h.Logger.WithFields(map[string]interface{}{
				"user_id": userId,
				"item_id": itemId,
			}).Warn("item not found")
			utils.NewErrorResponse(c, http.StatusNotFound, "Item not found", map[string]string{
				"itemId": "The requested item does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
		}).Errorf("failed to retrieve item: %s", err)
		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve item", map[string]string{
			"database": "Error during item retrieval",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "Item retrieved successfully", item)
}

// updateItem godoc
// @Summary Update an item
// @Description Update a specific item by ID for the authenticated user
// @Tags items
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param input body entity.UpdateItemInput true "Updated item data"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "Item updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid input or itemId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "Item not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to update item"
// @Router /api/items/{id} [put]
func (h *Handler) UpdateItem(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid itemId parameter: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid itemId param", map[string]string{
			"param": "itemId must be a valid integer",
		})
		return
	}

	listId, err := utils.ParseRequiredParamAsInt(c, "list_id")
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
		}).Errorf("invalid listId param: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	var input entity.UpdateItemInput
	if err := c.BindJSON(&input); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
			"list_id": listId,
		}).Errorf("failed to bind JSON: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid input", map[string]string{
			"body": "Invalid or malformed JSON",
		})

		return
	}

	if err := input.Validate(); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
			"list_id": listId,
		}).Warn("validation failed")

		utils.NewErrorResponse(c, http.StatusBadRequest, "Validation failed", map[string]string{
			"validation": err.Error(),
		})

		return
	}

	err = h.Usecases.Item.UpdateOneById(c.Request.Context(), userId, listId, itemId, input)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrItemNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "Item not found", map[string]string{
				"itemId": "The requested item does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
			"list_id": listId,
		}).Errorf("failed to update item: %s", err)
		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to update item", map[string]string{
			"database": "Error during item update",
		})
		return
	}

	utils.NewSuccessResponse(c, http.StatusOK, "Item updated successfully", nil)
}

// deleteItem godoc
// @Summary Delete an item
// @Description Delete a specific item by ID for the authenticated user
// @Tags items
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "Item deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid itemId param"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 404 {object} utils.ErrorResponse "Item not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to delete item"
// @Router /api/items/{id} [delete]
func (h *Handler) DeleteItem(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid itemId parameter: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid itemId param", map[string]string{
			"param": "itemId must be a valid integer",
		})
		return
	}

	listId, err := utils.ParseRequiredParamAsInt(c, "list_id")
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
		}).Errorf("invalid listId param: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "listId must be a valid integer",
		})
		return
	}

	err = h.Usecases.Item.DeleteOneById(c.Request.Context(), userId, listId, itemId)
	if err != nil {
		if err == utils.ErrUserNotOwner || err == utils.ErrItemNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "Item not found", map[string]string{
				"itemId": "The requested item does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
			"item_id": itemId,
			"list_id": listId,
		}).Errorf("failed to delete item: %s", err)
		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to delete item", map[string]string{
			"database": "Error during item deletion",
		})
		return
	}

	h.Metrics.IncrementDeletedItems()

	utils.NewSuccessResponse(c, http.StatusOK, "Item deleted successfully", nil)
}

// deleteItemByAdmin godoc
// @Summary Delete an item by admin
// @Description Delete a specific item by ID by admin
// @Tags items
// @Param id path int true "Item ID"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse "Item deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid itemId param"
// @Failure 404 {object} utils.ErrorResponse "Item not found"
// @Failure 500 {object} utils.ErrorResponse "Failed to delete item"
// @Router /private/api/items/{id} [delete]
func (h *Handler) DeleteItemByAdmin(c *gin.Context) {
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.Errorf("invalid itemId parameter: %s", err)

		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid itemId param", map[string]string{
			"param": "itemId must be a valid integer",
		})
		return
	}

	err = h.Usecases.Item.DeleteOneByAdmin(c.Request.Context(), itemId)
	if err != nil {
		if err == utils.ErrItemNotFound {
			utils.NewErrorResponse(c, http.StatusNotFound, "Item not found", map[string]string{
				"itemId": "The requested item does not exist",
			})
			return
		}

		h.Logger.WithFields(map[string]interface{}{
			"item_id": itemId,
		}).Errorf("failed to delete item: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to delete item", map[string]string{
			"database": "Error during item deletion",
		})

		return
	}

	h.Metrics.IncrementDeletedItems()

	utils.NewSuccessResponse(c, http.StatusOK, "Item deleted successfully", nil)
}

// searchItems godoc
// @Summary Search items
// @Description Search for items of the authenticated user by a query parameter
// @Tags items
// @Security BearerAuth
// @Produce json
// @Param query query string true "Search query"
// @Param appId header string true "Application ID"
// @Param appKey header string true "Application Key"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.Item} "Items searched successfully"
// @Failure 400 {object} utils.ErrorResponse "Empty query parameter"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Failed to search items"
// @Router /api/items/search [get]
func (h *Handler) SearchItems(c *gin.Context) {
	userId, err := middleware.GetUserId(c)
	if err != nil {
		h.Logger.Errorf("failed to get user ID: %s", err)
		utils.NewErrorResponse(c, http.StatusUnauthorized, "Unauthorized", map[string]string{
			"auth": "User authentication failed or user not logged in",
		})
		return
	}

	var listId *int
	err = utils.ParseOptionalParamAsInt(c, "list_id", &listId)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf("invalid listId param: %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid listId param", map[string]string{
			"param": "list_id must be a valid integer",
		})
		return
	}

	var done *bool
	err = utils.ParseOptionalParamAsBool(c, "done", &done)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": userId,
		}).Errorf(" : %s", err)
		utils.NewErrorResponse(c, http.StatusBadRequest, "Invalid done param", map[string]string{
			"param": "done must be a valid boolean",
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

	items, err := h.Usecases.Item.Search(c.Request.Context(), userId, listId, done, searchText)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":     userId,
			"search_text": searchText,
		}).Errorf("failed to search items: %s", err)

		utils.NewErrorResponse(c, http.StatusInternalServerError, "Failed to search items", map[string]string{
			"database": "Error during item search",
		})
		return
	}

	h.Metrics.IncrementSearchedItems()

	utils.NewSuccessResponse(c, http.StatusOK, "Item searched successfully", items)
}
