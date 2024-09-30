package integration_test

import (
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/berikulyBeket/todo-plus/internal/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateItem tests the creation of items
func TestCreateItem(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Error deleting test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Error deleting test list")
	}()

	testCases := []struct {
		name           string
		listId         string
		expectedStatus int
		expectedMsg    string
		requestBody    map[string]interface{}
		sendUserToken  bool
		cleanTestItem  bool
	}{
		{
			name:   "Positive case: Successful item creation",
			listId: strconv.Itoa(list.Id),
			requestBody: map[string]interface{}{
				"title":       "Test Item Title",
				"description": "Test Item Description",
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "Item created successfully",
			sendUserToken:  true,
			cleanTestItem:  true,
		},
		{
			name:   "Negative case: Unauthorized access",
			listId: strconv.Itoa(list.Id),
			requestBody: map[string]interface{}{
				"title":       "Test Item Title",
				"description": "Test Item Description",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
			cleanTestItem:  false,
		},
		{
			name:   "Negative case: Invalid listId",
			listId: "invalidID",
			requestBody: map[string]interface{}{
				"title":       "Test Item Title",
				"description": "Test Item Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
			sendUserToken:  true,
			cleanTestItem:  false,
		},
		{
			name:   "Negative case: Invalid JSON format",
			listId: strconv.Itoa(list.Id),
			requestBody: map[string]interface{}{
				"title":       "Test Item Title",
				"description": 123, // Should be a string
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid input",
			sendUserToken:  true,
			cleanTestItem:  false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				"Content-Type":          "application/json",
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			url := listUrl + testCase.listId + "/items/"
			resp := sendRequest(t, "POST", url, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.cleanTestItem && resp.StatusCode == http.StatusCreated {
				itemId, err := getIdFromJsonResponse(jsonResponse)
				assert.NoError(t, err, "Error getting item ID from response")
				defer func() {
					err := deleteTestItem(itemId)
					require.NoError(t, err, "Failed to delete test item")
				}()
			}
		}()
	}

	wg.Wait()
}

// TestGetAllItems tests retrieving all items in a list
func TestGetAllItems(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name              string
		listId            string
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: Successful retrieval of items",
			listId:            strconv.Itoa(list.Id),
			expectedStatus:    http.StatusOK,
			expectedMsg:       "Items retrieved successfully",
			sendUserToken:     true,
			expectedDataField: true,
		},
		{
			name:              "Negative case: Unauthorized access",
			listId:            strconv.Itoa(list.Id),
			expectedStatus:    http.StatusUnauthorized,
			expectedMsg:       "Unauthorized",
			sendUserToken:     false,
			expectedDataField: false,
		},
		{
			name:              "Negative case: Invalid listId",
			listId:            "invalidID",
			expectedStatus:    http.StatusBadRequest,
			expectedMsg:       "Invalid listId param",
			sendUserToken:     true,
			expectedDataField: false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			url := listUrl + testCase.listId + "/items/"
			resp := sendRequest(t, "GET", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.expectedDataField {
				assert.Contains(t, jsonResponse, "data", "Data field is missing")
				assert.NotEmpty(t, jsonResponse["data"], "Data field is empty")
			}
		}()
	}

	wg.Wait()
}

// TestGetItemById tests retrieving an item by its Id
func TestGetItemById(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name              string
		itemId            string
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: Item retrieved successfully",
			itemId:            strconv.Itoa(item.Id),
			expectedStatus:    http.StatusOK,
			expectedMsg:       "Item retrieved successfully",
			sendUserToken:     true,
			expectedDataField: true,
		},
		{
			name:              "Negative case: Unauthorized access",
			itemId:            strconv.Itoa(item.Id),
			expectedStatus:    http.StatusUnauthorized,
			expectedMsg:       "Unauthorized",
			sendUserToken:     false,
			expectedDataField: false,
		},
		{
			name:              "Negative case: Invalid item ID",
			itemId:            "invalidID",
			expectedStatus:    http.StatusBadRequest,
			expectedMsg:       "Invalid itemId param",
			sendUserToken:     true,
			expectedDataField: false,
		},
		{
			name:              "Negative case: Item not found",
			itemId:            "999999",
			expectedStatus:    http.StatusNotFound,
			expectedMsg:       "Item not found",
			sendUserToken:     true,
			expectedDataField: false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			url := itemUrl + testCase.itemId
			resp := sendRequest(t, "GET", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.expectedDataField {
				itemIdFromResponse, err := getIdFromJsonResponse(jsonResponse)
				assert.NoError(t, err, "Error getting item ID from response")
				assert.Equal(t, testCase.itemId, strconv.Itoa(itemIdFromResponse))
			}
		}()
	}

	wg.Wait()
}

// TestUpdateItem tests updating an item
func TestUpdateItem(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name           string
		listId         string
		itemId         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedMsg    string
		sendUserToken  bool
	}{
		{
			name:   "Positive case: Item updated successfully",
			listId: "3",
			itemId: strconv.Itoa(item.Id),
			requestBody: map[string]interface{}{
				"title":       "Updated Item Title",
				"description": "Updated Item Description",
				"done":        true,
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Item updated successfully",
			sendUserToken:  true,
		},
		{
			name:   "Negative case: Unauthorized access",
			listId: "3",
			itemId: strconv.Itoa(item.Id),
			requestBody: map[string]interface{}{
				"title":       "Updated Item Title",
				"description": "Updated Item Description",
				"done":        true,
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
		},
		{
			name:   "Negative case: Invalid list Id",
			listId: "",
			itemId: strconv.Itoa(item.Id),
			requestBody: map[string]interface{}{
				"title":       "Updated Item Title",
				"description": "Updated Item Description",
				"done":        true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
			sendUserToken:  true,
		},
		{
			name:   "Negative case: Invalid item Id",
			listId: "3",
			itemId: "invalidID",
			requestBody: map[string]interface{}{
				"title":       "Updated Item Title",
				"description": "Updated Item Description",
				"done":        true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid itemId param",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Empty request body",
			listId:         "3",
			itemId:         strconv.Itoa(item.Id),
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
			sendUserToken:  true,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				"Content-Type":          "application/json",
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			url := itemUrl + testCase.itemId + "?list_id=" + testCase.listId
			resp := sendRequest(t, "PUT", url, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestDeleteItem tests deleting an item
func TestDeleteItem(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name           string
		listId         string
		itemId         string
		expectedStatus int
		expectedMsg    string
		sendUserToken  bool
	}{
		{
			name:           "Positive case: Item deleted successfully",
			listId:         "3",
			itemId:         strconv.Itoa(item.Id),
			expectedStatus: http.StatusOK,
			expectedMsg:    "Item deleted successfully",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Unauthorized access",
			listId:         "3",
			itemId:         strconv.Itoa(item.Id),
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
		},
		{
			name:           "Negative case: Invalid list Id",
			listId:         "",
			itemId:         strconv.Itoa(item.Id),
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Invalid item Id",
			listId:         "3",
			itemId:         "invalidID",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid itemId param",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Item not found",
			listId:         "3",
			itemId:         "999999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Item not found",
			sendUserToken:  true,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			url := itemUrl + testCase.itemId + "?list_id=" + testCase.listId
			resp := sendRequest(t, "DELETE", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestDeleteItemByAdmin tests deleting an item by an admin
func TestDeleteItemByAdmin(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name           string
		itemId         string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Positive case: Item deleted successfully by admin",
			itemId:         strconv.Itoa(item.Id),
			expectedStatus: http.StatusOK,
			expectedMsg:    "Item deleted successfully",
		},
		{
			name:           "Negative case: Invalid item ID",
			itemId:         "invalidID",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid itemId param",
		},
		{
			name:           "Negative case: Item not found",
			itemId:         "999999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Item not found",
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				middleware.HeaderAppID:  privateAppId,
				middleware.HeaderAppKey: privateAppKey,
			}

			url := deleteItemByAdminUrl + "/" + testCase.itemId
			resp := sendRequest(t, "DELETE", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestSearchItems tests searching items
func TestSearchItems(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	list := setupTestList(t, token)
	defer func() {
		err := deleteTestList(list.Id)
		require.NoError(t, err, "Failed to delete test list")
	}()

	item := setupTestItem(t, token, list.Id)
	defer func() {
		err := deleteTestItem(item.Id)
		require.NoError(t, err, "Failed to delete test item")
	}()

	testCases := []struct {
		name              string
		searchQuery       string
		listId            *int
		done              *bool
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: Items searched successfully",
			searchQuery:       item.Title,
			listId:            &list.Id,
			done:              nil,
			expectedStatus:    http.StatusOK,
			expectedMsg:       "Item searched successfully",
			sendUserToken:     true,
			expectedDataField: true,
		},
		{
			name:              "Negative case: Empty searchText param",
			searchQuery:       "",
			listId:            nil,
			done:              nil,
			expectedStatus:    http.StatusBadRequest,
			expectedMsg:       "Empty searchText param",
			sendUserToken:     true,
			expectedDataField: false,
		},
		{
			name:        "Negative case: Unauthorized access",
			searchQuery: "test",
			listId:      nil,

			expectedStatus:    http.StatusUnauthorized,
			expectedMsg:       "Unauthorized",
			sendUserToken:     false,
			expectedDataField: false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func() {
			defer wg.Done()

			headers := map[string]string{
				middleware.HeaderAppID:  appKey,
				middleware.HeaderAppKey: appId,
			}
			if testCase.sendUserToken {
				headers[middleware.AuthorizationHeader] = "Bearer " + token
			}

			queryParams := map[string]string{
				"search_text": testCase.searchQuery,
			}
			if testCase.listId != nil {
				queryParams["list_id"] = strconv.Itoa(*testCase.listId)
			}
			if testCase.done != nil {
				queryParams["done"] = strconv.FormatBool(*testCase.done)
			}

			if testCase.expectedDataField {
				time.Sleep(searchableTimeoutSec * time.Second)
			}

			resp := sendRequestWithQuery(t, "GET", itemSearchUrl, headers, queryParams)
			jsonResponse := parseResponse(t, resp)

			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.expectedDataField {
				assert.Contains(t, jsonResponse, "data", "Data field is missing")
				assert.NotEmpty(t, jsonResponse["data"], "Data field is empty")
			}
		}()
	}

	wg.Wait()
}
