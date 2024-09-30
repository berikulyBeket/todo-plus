package integration_test

import (
	"net/http"
	"sync"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/middleware"

	"github.com/bxcodec/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignUp tests the user registration functionality
func TestSignUp(t *testing.T) {
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedMsg    string
		cleanTestUser  bool
	}{
		{
			name: "Positive case: Successful sign up",
			requestBody: map[string]interface{}{
				"name":     faker.FirstName(),
				"username": faker.Username(),
				"password": faker.Password(),
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "User registered successfully",
			cleanTestUser:  true,
		},
		{
			name: "Negative case: Invalid JSON",
			requestBody: map[string]interface{}{
				"malformed": "{username: missing_quotes}",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid request body",
			cleanTestUser:  false,
		},
		{
			name:           "Negative case: Missing required fields",
			requestBody:    map[string]interface{}{"username": "missing_password"},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid request body",
			cleanTestUser:  false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			headers := map[string]string{
				"Content-Type":          "application/json",
				middleware.HeaderAppID:  appId,
				middleware.HeaderAppKey: appKey,
			}

			resp := sendRequest(t, "POST", signUpUrl, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.cleanTestUser && resp.StatusCode == http.StatusOK {
				userId, err := getIdFromJsonResponse(jsonResponse)
				assert.NoError(t, err, "Error getting user ID from response")
				defer func() {
					err := deleteTestUser(userId)
					assert.NoError(t, err, "Failed to delete test user")
				}()
			}
		})
	}
}

// TestSignIn tests the user authentication functionality
func TestSignIn(t *testing.T) {
	userName := faker.Username()
	userPassword := faker.Password()

	userId, err := createTestUser(t, faker.FirstName(), userName, userPassword)
	assert.NoError(t, err, "Error creating user")
	defer func() {
		err := deleteTestUser(userId)
		require.NoError(t, err, "Failed to delete test user")
	}()

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedMsg    string
		returnsToken   bool
	}{
		{
			name: "Positive case: Successful sign in",
			requestBody: map[string]interface{}{
				"username": userName,
				"password": userPassword,
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Signed in successfully",
			returnsToken:   true,
		},
		{
			name: "Negative case: Missing username",
			requestBody: map[string]interface{}{
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid request body",
			returnsToken:   false,
		},
		{
			name: "Negative case: Invalid username or password",
			requestBody: map[string]interface{}{
				"username": faker.Username() + faker.Timestamp(),
				"password": faker.Password(),
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid username or password",
			returnsToken:   false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testCases {
		testCase := testCase
		wg.Add(1)

		go func(testCase struct {
			name           string
			requestBody    map[string]interface{}
			expectedStatus int
			expectedMsg    string
			returnsToken   bool
		}) {
			defer wg.Done()

			headers := map[string]string{
				"Content-Type":          "application/json",
				middleware.HeaderAppID:  appId,
				middleware.HeaderAppKey: appKey,
			}

			resp := sendRequest(t, "POST", signInUrl, headers, testCase.requestBody)
			jsonResponse := parseResponse(t, resp)
			assertResponse(t, resp, testCase.expectedStatus, testCase.expectedMsg, jsonResponse)

			if testCase.returnsToken && resp.StatusCode == http.StatusOK {
				assert.Contains(t, jsonResponse, "data", "Data field not found")
				data, ok := jsonResponse["data"].(map[string]interface{})
				assert.True(t, ok, "Data field is invalid")
				token, ok := data["token"].(string)
				assert.True(t, ok, "Token field not found or invalid")
				assert.NotEmpty(t, token, "Token should not be empty")
			}
		}(testCase)
	}

	wg.Wait()
}
