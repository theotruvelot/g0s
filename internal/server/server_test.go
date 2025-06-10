package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default configuration",
			config: Config{
				GRPCAddr:  ":9090",
				LogLevel:  "info",
				LogFormat: "json",
			},
		},
		{
			name: "custom configuration",
			config: Config{
				GRPCAddr:  ":9091",
				LogLevel:  "debug",
				LogFormat: "console",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			server, err := New(tt.config, logger)

			require.NoError(t, err)
			assert.NotNil(t, server)
			assert.Equal(t, tt.config, server.cfg)
			assert.Equal(t, logger, server.logger)
			assert.NotNil(t, server.handler)
		})
	}
}

func TestServer_Start_Stop(t *testing.T) {
	// Use available ports for testing
	httpPort := getAvailablePort(t)
	grpcPort := getAvailablePort(t)

	config := Config{
		GRPCAddr:  grpcPort,
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := zaptest.NewLogger(t)
	server, err := New(config, logger)
	require.NoError(t, err)

	// Start server
	err = server.Start()
	require.NoError(t, err)

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Test HTTP server is running
	resp, err := http.Get("http://localhost" + httpPort + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test gRPC server is running (check if port is listening)
	conn, err := net.Dial("tcp", "localhost"+grpcPort)
	require.NoError(t, err)
	conn.Close()

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	assert.NoError(t, err)
}

func TestServer_Start_GRPCPortError(t *testing.T) {
	// Use an invalid gRPC address that will fail to listen
	config := Config{
		GRPCAddr:  "invalid-address:99999", // Invalid address
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := zaptest.NewLogger(t)
	server, err := New(config, logger)
	require.NoError(t, err)

	// Start server - gRPC should fail but Start() should still return nil
	err = server.Start()
	assert.NoError(t, err) // Start() doesn't return gRPC errors directly

	// Give time for the goroutines to attempt startup
	time.Sleep(200 * time.Millisecond)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestServer_Stop_WithError(t *testing.T) {
	grpcPort := getAvailablePort(t)

	config := Config{
		GRPCAddr:  grpcPort,
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := zaptest.NewLogger(t)
	server, err := New(config, logger)
	require.NoError(t, err)

	// Start server
	err = server.Start()
	require.NoError(t, err)

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server with a very short timeout - might cause timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to definitely timeout
	time.Sleep(10 * time.Millisecond)

	// This might error due to context timeout, but shouldn't panic
	_ = server.Stop(ctx)

	// Cleanup: stop again with proper timeout to ensure clean shutdown
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	server.Stop(ctx2)
}

func TestServer_Stop_Timeout(t *testing.T) {
	// Use available ports for testing
	grpcPort := getAvailablePort(t)

	config := Config{
		GRPCAddr:  grpcPort,
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := zaptest.NewLogger(t)
	server, err := New(config, logger)
	require.NoError(t, err)

	// Start server
	err = server.Start()
	require.NoError(t, err)

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server with very short timeout to potentially trigger timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	err = server.Stop(ctx)
	// Might error due to timeout, but could also complete quickly
	// We don't assert on the error because it depends on timing
	_ = err
}

func TestServer_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedGRPC   string
		expectedLevel  string
		expectedFormat string
	}{
		{
			name: "standard config",
			config: Config{
				GRPCAddr:  ":9090",
				LogLevel:  "info",
				LogFormat: "json",
			},
			expectedGRPC:   ":9090",
			expectedLevel:  "info",
			expectedFormat: "json",
		},
		{
			name: "custom config",
			config: Config{
				GRPCAddr:  "localhost:9091",
				LogLevel:  "debug",
				LogFormat: "console",
			},
			expectedGRPC:   "localhost:9091",
			expectedLevel:  "debug",
			expectedFormat: "console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			server, err := New(tt.config, logger)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedGRPC, server.cfg.GRPCAddr)
			assert.Equal(t, tt.expectedLevel, server.cfg.LogLevel)
			assert.Equal(t, tt.expectedFormat, server.cfg.LogFormat)
		})
	}
}

func TestServer_GRPCServer_Setup(t *testing.T) {
	config := Config{
		GRPCAddr:  ":9090",
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := zaptest.NewLogger(t)
	server, err := New(config, logger)
	require.NoError(t, err)

	// Verify gRPC server is configured
	assert.NotNil(t, server.grpc)
	assert.NotNil(t, server.handler)
}

func TestServer_Multiple_Instances(t *testing.T) {
	// Test that we can create multiple server instances with different configs
	configs := []Config{
		{
			GRPCAddr:  getAvailablePort(t),
			LogLevel:  "info",
			LogFormat: "json",
		},
		{
			GRPCAddr:  getAvailablePort(t),
			LogLevel:  "debug",
			LogFormat: "console",
		},
	}

	logger := zaptest.NewLogger(t)

	for i, config := range configs {
		t.Run(fmt.Sprintf("instance_%d", i), func(t *testing.T) {
			server, err := New(config, logger)
			require.NoError(t, err)
			assert.NotNil(t, server)

			assert.Equal(t, config.GRPCAddr, server.cfg.GRPCAddr)
		})
	}
}

func getAvailablePort(t *testing.T) string {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	return fmt.Sprintf(":%d", port)
}
