package healthcheck

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/theotruvelot/g0s/pkg/logger"
	health "github.com/theotruvelot/g0s/pkg/proto/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type mockHealthServer struct {
	health.UnimplementedHealthServiceServer
	serving bool
}

func (m *mockHealthServer) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	if m.serving {
		return &health.HealthCheckResponse{
			Status: health.HealthCheckResponse_SERVING,
		}, nil
	}
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_NOT_SERVING,
	}, nil
}

func (m *mockHealthServer) Watch(req *health.HealthCheckRequest, stream health.HealthService_WatchServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream canceled")
		default:
			if err := stream.Send(&health.HealthCheckResponse{
				Status: health.HealthCheckResponse_SERVING,
			}); err != nil {
				return err
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func setupTest(t *testing.T) (*grpc.ClientConn, *mockHealthServer) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	mockServer := &mockHealthServer{serving: true}
	health.RegisterHealthServiceServer(s, mockServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	return conn, mockServer
}

func TestHealthCheck(t *testing.T) {
	conn, mockServer := setupTest(t)
	defer conn.Close()

	logger.InitLogger(logger.Config{
		Level:     "debug",
		Format:    "console",
		Component: "test",
	})

	service := New(conn, logger.GetLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test initial health check
	err := service.Start(ctx, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to start health check: %v", err)
	}

	// Wait for health check to complete
	time.Sleep(200 * time.Millisecond)

	if !service.IsHealthy() {
		t.Error("Expected service to be healthy")
	}

	// Test unhealthy state
	mockServer.serving = false
	time.Sleep(200 * time.Millisecond)

	if service.IsHealthy() {
		t.Error("Expected service to be unhealthy")
	}
}

func TestHealthCheckCancellation(t *testing.T) {
	conn, _ := setupTest(t)
	defer conn.Close()

	logger.InitLogger(logger.Config{
		Level:     "debug",
		Format:    "console",
		Component: "test",
	})

	service := New(conn, logger.GetLogger())

	ctx, cancel := context.WithCancel(context.Background())

	err := service.Start(ctx, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to start health check: %v", err)
	}

	// Wait for health check to establish
	time.Sleep(200 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for cancellation to propagate
	time.Sleep(200 * time.Millisecond)

	if service.IsHealthy() {
		t.Error("Expected service to be unhealthy after cancellation")
	}
}
