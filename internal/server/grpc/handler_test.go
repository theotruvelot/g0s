package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestHandler_RegisterServices(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "successful service registration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := New(logger)
			server := grpc.NewServer()

			// This should not panic
			assert.NotPanics(t, func() {
				handler.RegisterServices(server)
			})

			// Verify server is still valid
			assert.NotNil(t, server)
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
	server := grpc.NewServer()

	// This should not panic and should work correctly
	assert.NotPanics(t, func() {
		handler.RegisterServices(server)
	})
}

func TestHandler_MultipleRegistrations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := New(logger)
	server := grpc.NewServer()

	// Register services multiple times should not panic
	assert.NotPanics(t, func() {
		handler.RegisterServices(server)
		handler.RegisterServices(server)
	})
}
