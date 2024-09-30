package integration_test

import (
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/middleware"

	"github.com/stretchr/testify/require"
)

// TestDeleteUserByAdmin tests deleting a by an admin
func TestDeleteUserByAdmin(t *testing.T) {
	userId, _ := setupTestUser(t)
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	testCases := []struct {
		name           string
		userId         string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Positive case: Successful delete user",
			userId:         strconv.Itoa(userId),
			expectedStatus: http.StatusOK,
			expectedMsg:    "User deleted successfully",
		},
		{
			name:           "Negative case: Invalid userId param",
			userId:         "invalidUserId",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid userId param",
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

			url := deleteUserByAdminUrl + "/" + testCase.userId
			resp := sendRequest(t, "DELETE", url, headers, nil)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)
		}()
	}

	wg.Wait()
}
