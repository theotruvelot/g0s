package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		timeout     time.Duration
		wantTimeout time.Duration
	}{
		{
			name:        "with custom timeout",
			baseURL:     "http://example.com",
			token:       "test-token",
			timeout:     5 * time.Second,
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "with zero timeout uses default",
			baseURL:     "http://example.com",
			token:       "test-token",
			timeout:     0,
			wantTimeout: _defaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.token, tt.timeout)
			assert.NotNil(t, client)
			assert.Equal(t, tt.baseURL, client.baseURL)
			assert.Equal(t, tt.token, client.token)
			assert.Equal(t, tt.wantTimeout, client.httpClient.Timeout)
		})
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       io.Reader
		token      string
		setupMock  func() *httptest.Server
		wantErr    bool
		wantStatus int
	}{
		{
			name:   "successful GET request",
			method: http.MethodGet,
			path:   "/test",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/test", r.URL.Path)
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "successful POST request with body",
			method: http.MethodPost,
			path:   "/test",
			body:   strings.NewReader(`{"test":"data"}`),
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/test", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					w.WriteHeader(http.StatusCreated)
				}))
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "request with auth token",
			method: http.MethodGet,
			path:   "/test",
			token:  "test-token",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "nil context",
			method:  http.MethodGet,
			path:    "/test",
			wantErr: true,
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Cette fonction ne devrait jamais être appelée car le contexte est nil
					t.Error("Server should not be called with nil context")
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			var baseURL string

			if tt.setupMock != nil {
				server = tt.setupMock()
				defer server.Close()
				baseURL = server.URL
			} else {
				baseURL = "http://example.com"
			}

			client := NewClient(baseURL, tt.token, 0)
			ctx := context.Background()
			if tt.wantErr {
				ctx = nil
			}

			resp, err := client.Do(ctx, tt.method, tt.path, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				var httpErr *HTTPError
				assert.ErrorAs(t, err, &httpErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestClient_Convenience_Methods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "", 0)
	ctx := context.Background()

	t.Run("GET", func(t *testing.T) {
		resp, err := client.Get(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.Post(ctx, "/test", strings.NewReader("test"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.Put(ctx, "/test", strings.NewReader("test"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.Delete(ctx, "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{
		Method: http.MethodGet,
		URL:    "http://example.com",
		Err:    fmt.Errorf("connection refused"),
	}

	expected := "HTTP request failed [GET http://example.com]: connection refused"
	assert.Equal(t, expected, err.Error())
}
