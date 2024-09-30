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

// TestCreateList tests the creation of lists
func TestCreateList(t *testing.T) {
	userId, token := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	testCases := []struct {
		name           string
		expectedStatus int
		expectedMsg    string
		requestBody    map[string]interface{}
		sendUserToken  bool
		cleanTestList  bool
	}{
		{
			name: "Positive case: Successfully create list",
			requestBody: map[string]interface{}{
				"title":       "Test List Title",
				"description": "Test List Description",
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "List created successfully",
			sendUserToken:  true,
			cleanTestList:  true,
		},
		{
			name: "Negative case: Unauthorized access",
			requestBody: map[string]interface{}{
				"title":       "Test List Title",
				"description": "Test List Description",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
			cleanTestList:  false,
		},
		{
			name: "Negative case: Invalid or malformed JSON",
			requestBody: map[string]interface{}{
				"title":       "Test List Title",
				"description": 123,
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid input",
			sendUserToken:  true,
			cleanTestList:  false,
		},
		{
			name: "Negative case: Missing required field",
			requestBody: map[string]interface{}{
				"description": "Test List Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid input",
			sendUserToken:  true,
			cleanTestList:  false,
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

			resp := sendRequest(t, "POST", listUrl, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.cleanTestList && resp.StatusCode == http.StatusCreated {
				listId, err := getIdFromJsonResponse(jsonResponse)
				assert.NoError(t, err, "Error getting list ID from response")

				defer func() {
					err := deleteTestList(listId)
					require.NoError(t, err, "Failed to delete test list")
				}()
			}
		}()
	}

	wg.Wait()
}

// TestGetAllLists tests retrieving all lists
func TestGetAllLists(t *testing.T) {
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

	testCases := []struct {
		name              string
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: Lists retrieved successfully",
			expectedStatus:    http.StatusOK,
			expectedMsg:       "Lists retrieved successfully",
			sendUserToken:     true,
			expectedDataField: true,
		},
		{
			name:              "Negative case: Unauthorized access",
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

			resp := sendRequest(t, "GET", listUrl, headers, nil)
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

// TestGetListById tests retrieving a list by its Id
func TestGetListById(t *testing.T) {
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

	testCases := []struct {
		name              string
		listId            string
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: List retrieved successfully",
			listId:            strconv.Itoa(list.Id),
			expectedStatus:    http.StatusOK,
			expectedMsg:       "List retrieved successfully",
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
			name:              "Negative case: Invalid list ID",
			listId:            "invalidID",
			expectedStatus:    http.StatusBadRequest,
			expectedMsg:       "Invalid listId param",
			sendUserToken:     true,
			expectedDataField: false,
		},
		{
			name:              "Negative case: List not found",
			listId:            "999999",
			expectedStatus:    http.StatusNotFound,
			expectedMsg:       "List not found",
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

			url := listUrl + testCase.listId
			resp := sendRequest(t, "GET", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.expectedDataField {
				listIdFromResponse, err := getIdFromJsonResponse(jsonResponse)
				assert.NoError(t, err, "Error getting list ID from response")
				assert.Equal(t, testCase.listId, strconv.Itoa(listIdFromResponse))
			}
		}()
	}

	wg.Wait()
}

// TestUpdateList tests updating a list
func TestUpdateList(t *testing.T) {
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

	testCases := []struct {
		name           string
		listId         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedMsg    string
		sendUserToken  bool
	}{
		{
			name:   "Positive case: List updated successfully",
			listId: strconv.Itoa(list.Id),
			requestBody: map[string]interface{}{
				"title":       "Updated List Title",
				"description": "Updated List Description",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "List updated successfully",
			sendUserToken:  true,
		},
		{
			name:   "Negative case: Unauthorized access",
			listId: strconv.Itoa(list.Id),
			requestBody: map[string]interface{}{
				"title":       "Updated List Title",
				"description": "Updated List Description",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
		},
		{
			name:   "Negative case: Invalid list ID",
			listId: "invalidID",
			requestBody: map[string]interface{}{
				"title":       "Updated List Title",
				"description": "Updated List Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Empty request body",
			listId:         strconv.Itoa(list.Id),
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

			url := listUrl + testCase.listId
			resp := sendRequest(t, "PUT", url, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestDeleteList tests deleting a list
func TestDeleteList(t *testing.T) {
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

	testCases := []struct {
		name           string
		listId         string
		expectedStatus int
		expectedMsg    string
		sendUserToken  bool
	}{
		{
			name:           "Positive case: List deleted successfully",
			listId:         strconv.Itoa(list.Id),
			expectedStatus: http.StatusOK,
			expectedMsg:    "List deleted successfully",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: Unauthorized access",
			listId:         strconv.Itoa(list.Id),
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Unauthorized",
			sendUserToken:  false,
		},
		{
			name:           "Negative case: Invalid list ID",
			listId:         "invalidID",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
			sendUserToken:  true,
		},
		{
			name:           "Negative case: List not found",
			listId:         "999999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "List not found",
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

			url := listUrl + testCase.listId
			resp := sendRequest(t, "DELETE", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestDeleteListByAdmin tests deleting a list by an admin
func TestDeleteListByAdmin(t *testing.T) {
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

	testCases := []struct {
		name           string
		listId         string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Positive case: List deleted successfully by admin",
			listId:         strconv.Itoa(list.Id),
			expectedStatus: http.StatusOK,
			expectedMsg:    "List deleted successfully",
		},
		{
			name:           "Negative case: Invalid list ID",
			listId:         "invalidID",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid listId param",
		},
		{
			name:           "Negative case: List not found",
			listId:         "999999",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "List not found",
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

			url := deleteListByAdminUrl + "/" + testCase.listId
			resp := sendRequest(t, "DELETE", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}

// TestSearchLists tests searching all lists
func TestSearchLists(t *testing.T) {
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

	testCases := []struct {
		name              string
		searchQuery       string
		expectedStatus    int
		expectedMsg       string
		sendUserToken     bool
		expectedDataField bool
	}{
		{
			name:              "Positive case: Lists searched successfully",
			searchQuery:       list.Title,
			expectedStatus:    http.StatusOK,
			expectedMsg:       "Lists searched successfully",
			sendUserToken:     true,
			expectedDataField: true,
		},
		{
			name:              "Negative case: Empty searchText param",
			searchQuery:       "",
			expectedStatus:    http.StatusBadRequest,
			expectedMsg:       "Empty searchText param",
			sendUserToken:     true,
			expectedDataField: false,
		},
		{
			name:              "Negative case: Unauthorized access",
			searchQuery:       "test",
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

			if testCase.expectedDataField {
				time.Sleep(searchableTimeoutSec * time.Second)
			}

			resp := sendRequestWithQuery(t, "GET", listSearchUrl, headers, queryParams)
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
