package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := Config{
		ServerURL: "http://example.com",
		Token:     "test-token",
		Interval:  time.Second,
	}

	service := New(cfg)
	assert.NotNil(t, service)
	assert.Equal(t, cfg.Interval, service.interval)
	assert.False(t, service.lastCheck)
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		expectHealthy  bool
	}{
		{
			name:           "healthy server",
			serverResponse: http.StatusOK,
			expectHealthy:  true,
		},
		{
			name:           "unhealthy server",
			serverResponse: http.StatusServiceUnavailable,
			expectHealthy:  false,
		},
		{
			name:           "server error",
			serverResponse: http.StatusInternalServerError,
			expectHealthy:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, _healthEndpoint, r.URL.Path)
				w.WriteHeader(tt.serverResponse)
			}))
			defer server.Close()

			service := New(Config{
				ServerURL: server.URL,
				Token:     "test-token",
				Interval:  time.Second,
			})

			ctx := context.Background()
			service.check(ctx)

			assert.Equal(t, tt.expectHealthy, service.IsHealthy())
		})
	}
}

func TestHealthCheckService(t *testing.T) {
	// Create a test server that alternates between healthy and unhealthy
	healthy := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if healthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		healthy = !healthy
	}))
	defer server.Close()

	service := New(Config{
		ServerURL: server.URL,
		Token:     "test-token",
		Interval:  100 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Start the service in a goroutine
	go service.Start(ctx)

	// Wait for initial check
	time.Sleep(50 * time.Millisecond)
	firstCheck := service.IsHealthy()

	// Wait for second check
	time.Sleep(100 * time.Millisecond)
	secondCheck := service.IsHealthy()

	assert.NotEqual(t, firstCheck, secondCheck, "Health status should alternate")
}

func TestHealthCheckServiceContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := New(Config{
		ServerURL: server.URL,
		Token:     "test-token",
		Interval:  time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		service.Start(ctx)
		close(done)
	}()

	// Wait for initial check
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for service to stop
	select {
	case <-done:
		// Service stopped as expected
	case <-time.After(time.Second):
		t.Fatal("Service did not stop after context cancellation")
	}
}

func TestHealthCheckWithServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := New(Config{
		ServerURL: server.URL,
		Token:     "test-token",
		Interval:  time.Second,
	})

	ctx := context.Background()
	service.check(ctx)

	assert.False(t, service.IsHealthy(), "Service should be unhealthy when server returns error")
}

func TestHealthCheckWithServerDown(t *testing.T) {
	// Use a non-existent server URL
	service := New(Config{
		ServerURL: "http://localhost:12345",
		Token:     "test-token",
		Interval:  time.Second,
	})

	ctx := context.Background()
	service.check(ctx)

	assert.False(t, service.IsHealthy(), "Service should be unhealthy when server is down")
}
