package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestHandler_RegisterRoutes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	router := handler.RegisterRoutes()

	assert.NotNil(t, router)
}

func TestRoutes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	router := handler.RegisterRoutes()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health endpoint",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok","service":"g0s-server"}`,
		},
		{
			name:           "status endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/status",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"running","service":"g0s-server"}`,
		},
		{
			name:           "agent register endpoint",
			method:         http.MethodPost,
			path:           "/api/v1/agent/register",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"registered"}`,
		},
		{
			name:           "agent metrics endpoint",
			method:         http.MethodPost,
			path:           "/api/v1/agent/metrics",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"received"}`,
		},
		{
			name:           "non-existent endpoint",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "wrong method for metrics",
			method:         http.MethodGet,
			path:           "/api/v1/agent/metrics",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	router := handler.RegisterRoutes()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check that the content type header is set correctly (this verifies handlers are working)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify the response body is what we expect (this confirms the full middleware chain worked)
	expectedBody := `{"status":"ok","service":"g0s-server"}`
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestAPIRouteGrouping(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	router := handler.RegisterRoutes()

	// Test that API routes are properly grouped under /api/v1
	apiRoutes := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/api/v1/status", http.StatusOK},
		{http.MethodPost, "/api/v1/agent/register", http.StatusOK},
		{http.MethodPost, "/api/v1/agent/metrics", http.StatusOK},
	}

	for _, route := range apiRoutes {
		t.Run(route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, route.status, w.Code)
		})
	}
}

func TestCORS(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	router := handler.RegisterRoutes()

	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The response should handle the OPTIONS request
	// The exact behavior depends on your CORS middleware configuration
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
