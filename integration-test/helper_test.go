package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/berikulyBeket/todo-plus/config"
	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/internal/middleware"

	"github.com/bxcodec/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	searchableTimeoutSec = 2
)

var (
	appId         = "123123"
	appKey        = "123123"
	privateAppId  = "456456"
	privateAppKey = "456456"

	signUpUrl            string
	signInUrl            string
	listUrl              string
	listSearchUrl        string
	itemUrl              string
	itemSearchUrl        string
	deleteUserByAdminUrl string
	deleteItemByAdminUrl string
	deleteListByAdminUrl string
)

func init() {
	cfg, err := config.NewConfig("../config/config.yml")
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	baseUrl := "http://" + cfg.HTTP.Host + ":" + cfg.HTTP.Port

	signUpUrl = baseUrl + "/v1/auth/sign-up"
	signInUrl = baseUrl + "/v1/auth/sign-in"

	listUrl = baseUrl + "/v1/api/lists/"
	listSearchUrl = listUrl + "search"

	itemUrl = baseUrl + "/v1/api/items/"
	itemSearchUrl = itemUrl + "search"

	deleteUserByAdminUrl = baseUrl + "/v1/private/api/users"
	deleteItemByAdminUrl = baseUrl + "/v1/private/api/items"
	deleteListByAdminUrl = baseUrl + "/v1/private/api/lists"
}

// sendRequest creates and sends an HTTP request and returns the response
func sendRequest(t *testing.T, method, url string, headers map[string]string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		reqBytes, err := json.Marshal(body)
		assert.NoError(t, err, "Error marshalling request body")
		reqBody = bytes.NewBuffer(reqBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	assert.NoError(t, err, "Error creating request")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	assert.NoError(t, err, "Error sending request")

	return resp
}

// sendRequestWithQuery creates and sends an HTTP request with query params and returns the response
func sendRequestWithQuery(t *testing.T, method, baseUrl string, headers, queryParams map[string]string) *http.Response {
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	finalUrl := fmt.Sprintf("%s?%s", baseUrl, query.Encode())

	req, err := http.NewRequest(method, finalUrl, nil)
	require.NoError(t, err, "Failed to create new HTTP request")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")

	return resp
}

// parseResponse reads and unmarshals the response body into a map
func parseResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Error reading response body")
	fmt.Printf("Response body: %s\n", string(respBody))

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	assert.NoError(t, err, "Error decoding response body")

	return jsonResponse
}

// assertResponse asserts the status code and message in the response
func assertResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedMsg string, jsonResponse map[string]interface{}) {
	assert.Equal(t, expectedStatus, resp.StatusCode, "Status code mismatch")
	assert.Contains(t, jsonResponse, "message", "Response does not contain 'message' field")
	assert.Equal(t, expectedMsg, jsonResponse["message"], "Response message mismatch")
}

// setupTestUser creates a test user and returns the user ID and token
func setupTestUser(t *testing.T) (userId int, token string) {
	userName := faker.Username()
	userPassword := faker.Password()

	userId, err := createTestUser(t, faker.Name(), userName, userPassword)
	assert.NoError(t, err, "Error creating user")

	token = getTestUserToken(t, userName, userPassword)
	assert.NotEmpty(t, token, "Token should not be empty")

	return userId, token
}

// setupTestList creates a test list and returns the list Id
func setupTestList(t *testing.T, token string) *entity.List {
	title := faker.Sentence()
	description := faker.Sentence()

	listId := createTestList(t, token, title, description)
	assert.NotZero(t, listId, "List ID should not be zero")

	list := &entity.List{
		Id:          listId,
		Title:       title,
		Description: description,
	}

	return list
}

// setupTestItem creates a test item and returns the item Id
func setupTestItem(t *testing.T, token string, listId int) *entity.Item {
	headers := map[string]string{
		"Content-Type":                 "application/json",
		middleware.HeaderAppID:         appKey,
		middleware.HeaderAppKey:        appId,
		middleware.AuthorizationHeader: "Bearer " + token,
	}

	title := faker.Sentence()
	description := faker.Sentence()

	requestBody := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	url := listUrl + strconv.Itoa(listId) + "/items/"
	resp := sendRequest(t, "POST", url, headers, requestBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201 Created")

	jsonResponse := parseResponse(t, resp)
	itemId, err := getIdFromJsonResponse(jsonResponse)
	assert.NoError(t, err, "Error getting item ID from response")

	return &entity.Item{
		Id:          itemId,
		Title:       title,
		Description: description,
	}
}

// createTestUser creates a new user for testing and returns the user Id
func createTestUser(t *testing.T, name, username, password string) (int, error) {
	requestBody := map[string]string{
		"name":     name,
		"username": username,
		"password": password,
	}

	headers := map[string]string{
		"Content-Type":          "application/json",
		middleware.HeaderAppID:  appId,
		middleware.HeaderAppKey: appKey,
	}

	resp := sendRequest(t, "POST", signUpUrl, headers, requestBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to create test user")

	jsonResponse := parseResponse(t, resp)
	userId, err := getIdFromJsonResponse(jsonResponse)
	assert.NoError(t, err, "Error getting user ID from response")

	return userId, nil
}

// createTestList creates a new list for testing and returns the list Id
func createTestList(t *testing.T, token, title, description string) int {
	requestBody := map[string]string{
		"title":       title,
		"description": description,
	}

	headers := map[string]string{
		"Content-Type":                 "application/json",
		middleware.AuthorizationHeader: "Bearer " + token,
		middleware.HeaderAppID:         appKey,
		middleware.HeaderAppKey:        appId,
	}

	resp := sendRequest(t, "POST", listUrl, headers, requestBody)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Error reading response body")
	fmt.Printf("Response body: %s\n", string(respBody))

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code mismatch")
	assert.NotZero(t, len(respBody), "Response body is empty")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	assert.NoError(t, err, "Error decoding response body")

	assert.Contains(t, jsonResponse, "message")
	assert.Equal(t, "List created successfully", jsonResponse["message"])

	listId, err := getIdFromJsonResponse(jsonResponse)
	assert.NoError(t, err, "Error getting list ID from response")

	return listId
}

// getTestUserToken authenticates the test user and returns the token
func getTestUserToken(t *testing.T, userName, password string) string {
	requestBody := map[string]interface{}{
		"username": userName,
		"password": password,
	}

	headers := map[string]string{
		"Content-Type":          "application/json",
		middleware.HeaderAppID:  appId,
		middleware.HeaderAppKey: appKey,
	}

	resp := sendRequest(t, "POST", signInUrl, headers, requestBody)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Error reading response body")
	fmt.Printf("Response body: %s\n", string(respBody))

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code mismatch")
	assert.NotZero(t, len(respBody), "Response body is empty")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	assert.NoError(t, err, "Error decoding response body")

	assert.Contains(t, jsonResponse, "message")
	assert.Equal(t, "Signed in successfully", jsonResponse["message"])

	data, ok := jsonResponse["data"].(map[string]interface{})
	assert.True(t, ok, "Data field not found or invalid")
	token, ok := data["token"].(string)
	assert.True(t, ok, "Token field not found or invalid")
	assert.NotEmpty(t, token, "Token should not be empty")

	return token
}

// getIdFromJsonResponse extracts the 'id' from the JSON response
func getIdFromJsonResponse(jsonResponse map[string]interface{}) (int, error) {
	data, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("Data field is missing or invalid")
	}

	idFloat, ok := data["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("ID field not found or invalid")
	}

	return int(idFloat), nil
}

// deleteTestUser deletes a test user
func deleteTestUser(userId int) error {
	url := fmt.Sprintf("%s/%d", deleteUserByAdminUrl, userId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set(middleware.HeaderAppID, privateAppId)
	req.Header.Set(middleware.HeaderAppKey, privateAppKey)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// deleteTestList deletes a test list
func deleteTestList(listId int) error {
	url := fmt.Sprintf("%s/%d", deleteListByAdminUrl, listId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set(middleware.HeaderAppID, privateAppId)
	req.Header.Set(middleware.HeaderAppKey, privateAppKey)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// deleteTestItem deletes a test item
func deleteTestItem(itemId int) error {
	url := fmt.Sprintf("%s/%d", deleteItemByAdminUrl, itemId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set(middleware.HeaderAppID, privateAppId)
	req.Header.Set(middleware.HeaderAppKey, privateAppKey)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
	}

	return nil
}
