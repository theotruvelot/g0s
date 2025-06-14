package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

// mockUnaryHandler is a mock gRPC unary handler for testing
func mockUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "test response", nil
}

// mockStreamHandler is a mock gRPC stream handler for testing
func mockStreamHandler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func TestLoggingUnaryInterceptor(t *testing.T) {
	// Create the interceptor
	interceptor := LoggingUnaryInterceptor()

	// Mock server info
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.TestService/TestMethod",
	}

	// Test the interceptor
	resp, err := interceptor(
		context.Background(),
		"test request",
		info,
		mockUnaryHandler,
	)

	// Check that the handler was called and returned expected values
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp != "test response" {
		t.Errorf("Expected 'test response', got %v", resp)
	}
}

func TestLoggingStreamInterceptor(t *testing.T) {
	// Create the interceptor
	interceptor := LoggingStreamInterceptor()

	// Mock server info
	info := &grpc.StreamServerInfo{
		FullMethod:     "/test.TestService/TestStreamMethod",
		IsClientStream: true,
		IsServerStream: true,
	}

	// Mock stream
	stream := &mockServerStream{
		ctx: context.Background(),
	}

	// Test the interceptor
	err := interceptor(
		nil, // srv interface{}
		stream,
		info,
		mockStreamHandler,
	)

	// Check that no error occurred
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// mockServerStream implements grpc.ServerStream for testing
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
