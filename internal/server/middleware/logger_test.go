package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theotruvelot/g0s/pkg/logger"
)

func TestRequestLogger(t *testing.T) {
	// Initialize logger for testing
	logger.InitLogger(logger.Config{
		Level:     "debug",
		Format:    "json",
		Component: "test",
	})

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		contentType    string
	}{
		{
			name:           "GET request",
			method:         "GET",
			path:           "/health",
			expectedStatus: 200,
		},
		{
			name:           "POST request with JSON body",
			method:         "POST",
			path:           "/api/v1/test",
			body:           `{"name":"test","value":123}`,
			expectedStatus: 201,
			contentType:    "application/json",
		},
		{
			name:           "PUT request with large body",
			method:         "PUT",
			path:           "/api/v1/large",
			body:           strings.Repeat("a", 5000), // Larger than maxBodyLogSize
			expectedStatus: 200,
			contentType:    "text/plain",
		},
		{
			name:           "DELETE request",
			method:         "DELETE",
			path:           "/api/v1/test/123",
			expectedStatus: 204,
		},
		{
			name:           "Request returning error",
			method:         "GET",
			path:           "/error",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with middleware
			r := chi.NewRouter()
			r.Use(middleware.RequestID)
			r.Use(RequestLogger())

			// Add test handlers
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte(`{"status":"ok"}`))
			})

			r.Post("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				w.Write([]byte(`{"id":1}`))
			})

			r.Put("/api/v1/large", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte(`{"received":true}`))
			})

			r.Delete("/api/v1/test/123", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(204)
			})

			r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
				w.Write([]byte(`{"error":"internal server error"}`))
			})

			// Create request
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPrepareBodyForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty body",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "small body",
			input:    []byte("small content"),
			expected: "small content",
		},
		{
			name:     "large body gets truncated",
			input:    bytes.Repeat([]byte("a"), maxBodyLogSize+100),
			expected: string(bytes.Repeat([]byte("a"), maxBodyLogSize)) + "... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareBodyForLogging(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLogLevelForStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   string // We'll check the string representation
	}{
		{"2xx success", 200, "info"},
		{"2xx created", 201, "info"},
		{"3xx redirect", 301, "info"},
		{"4xx client error", 400, "warn"},
		{"4xx not found", 404, "warn"},
		{"5xx server error", 500, "error"},
		{"5xx internal error", 503, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := getLogLevelForStatus(tt.statusCode)
			assert.Equal(t, tt.expected, level.String())
		})
	}
}

func TestCaptureRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "nil body",
			body:     "",
			expected: "",
		},
		{
			name:     "simple JSON",
			body:     `{"key":"value"}`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "empty string",
			body:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))

			if tt.body == "" {
				req.Body = nil
			}

			bodyBytes := captureRequestBody(req)

			if tt.expected == "" {
				assert.Empty(t, bodyBytes)
			} else {
				assert.Equal(t, tt.expected, string(bodyBytes))

				// Verify body is restored and can be read again
				restoredBody := make([]byte, len(tt.expected))
				n, err := req.Body.Read(restoredBody)
				require.NoError(t, err)
				assert.Equal(t, len(tt.expected), n)
				assert.Equal(t, tt.expected, string(restoredBody))
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: recorder,
		statusCode:     200,
		body:           &bytes.Buffer{},
	}

	// Test WriteHeader
	rw.WriteHeader(201)
	assert.Equal(t, 201, rw.statusCode)
	assert.Equal(t, 201, recorder.Code)

	// Test Write
	testData := []byte("test response")
	n, err := rw.Write(testData)
	require.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, string(testData), rw.body.String())
	assert.Equal(t, string(testData), recorder.Body.String())
}
