package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berikulyBeket/todo-plus/internal/middleware"
	appauth "github.com/berikulyBeket/todo-plus/pkg/app_auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper function to create the test router
func setupRouter(appAuth appauth.Interface, accessType string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.AppAuth(accessType, appAuth))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return router
}

// TestAppAuth tests the AppAuth middleware
func TestAppAuth(t *testing.T) {
	publicAppId := "publicAppId"
	publicAppKey := "publicAppKey"
	privateAppId := "privateAppId"
	privateAppKey := "privateAppKey"

	appAuthInstance := appauth.New(publicAppId, publicAppKey, privateAppId, privateAppKey)

	testCases := []struct {
		name         string
		accessType   string
		appId        string
		appKey       string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Valid Public Credentials",
			accessType:   appauth.PublicAccess,
			appId:        publicAppId,
			appKey:       publicAppKey,
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"success"}`,
		},
		{
			name:         "Invalid Public Credentials",
			accessType:   appauth.PublicAccess,
			appId:        "invalidAppId",
			appKey:       "invalidAppKey",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Unauthorized: Invalid appId or appKey"}`,
		},
		{
			name:         "Valid Private Credentials",
			accessType:   appauth.PrivateAccess,
			appId:        privateAppId,
			appKey:       privateAppKey,
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"success"}`,
		},
		{
			name:         "Empty Credentials",
			accessType:   appauth.PublicAccess,
			appId:        "",
			appKey:       "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Unauthorized: Empty appId or appKey"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			router := setupRouter(appAuthInstance, testCase.accessType)
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			if testCase.appId != "" {
				req.Header.Set(middleware.HeaderAppID, testCase.appId)
			}
			if testCase.appKey != "" {
				req.Header.Set(middleware.HeaderAppKey, testCase.appKey)
			}

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, testCase.expectedCode, resp.Code)
			assert.JSONEq(t, testCase.expectedBody, resp.Body.String())
		})
	}
}
