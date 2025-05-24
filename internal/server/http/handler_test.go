package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestHandler_HandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful health check",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok","service":"g0s-server"}`,
		},
		{
			name:           "health check with POST method",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok","service":"g0s-server"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := New(logger)

			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.HandleHealth(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHandler_HandleStatus(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful status check",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"running","service":"g0s-server"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := New(logger)

			req := httptest.NewRequest(tt.method, "/api/v1/status", nil)
			w := httptest.NewRecorder()

			handler.HandleStatus(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHandler_HandleMetrics(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful metrics submission",
			method:         http.MethodPost,
			body:           `{"metrics": {"cpu": 50.5}}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"received"}`,
		},
		{
			name:           "empty metrics submission",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"received"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := New(logger)

			req := httptest.NewRequest(tt.method, "/api/v1/agent/metrics", nil)
			w := httptest.NewRecorder()

			handler.HandleMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHandler_HandleAgentRegister(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful agent registration",
			method:         http.MethodPost,
			body:           `{"agent_id": "agent-123", "hostname": "test-host"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"registered"}`,
		},
		{
			name:           "empty registration request",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"registered"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := New(logger)

			req := httptest.NewRequest(tt.method, "/api/v1/agent/register", nil)
			w := httptest.NewRecorder()

			handler.HandleAgentRegister(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHandler_WithLogger(t *testing.T) {
	// Create a custom logger to verify it's being used
	logger := zaptest.NewLogger(t)
	handler := New(logger)

	// Verify the logger is properly set
	assert.Equal(t, logger, handler.logger)

	// Test that the handler works with the logger
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// This should not panic and should work correctly
	require.NotPanics(t, func() {
		handler.HandleHealth(w, req)
	})

	assert.Equal(t, http.StatusOK, w.Code)
}
