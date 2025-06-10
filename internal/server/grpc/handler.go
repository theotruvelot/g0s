package grpc

import (
	"context"
	"github.com/google/uuid"
	"github.com/theotruvelot/g0s/internal/server/storage/metrics"
	"sync"
	"time"

	health "github.com/theotruvelot/g0s/pkg/proto/health"
	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Handler struct {
	logger *zap.Logger
	store  *metrics.MetricsManager
	pb.UnimplementedMetricServiceServer
	health.UnimplementedHealthServiceServer
	ctx    context.Context
	cancel context.CancelFunc
}

type ClientInfo struct {
	ID          string
	Hostname    string
	IPAddress   string
	ConnectedAt time.Time
}

var (
	clients     = make(map[string]ClientInfo)
	clientsLock sync.Mutex
)

func New(logger *zap.Logger, store *metrics.MetricsManager) *Handler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Handler{
		logger: logger,
		store:  store,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (h *Handler) RegisterServices(server *grpc.Server) {
	pb.RegisterMetricServiceServer(server, h)
	health.RegisterHealthServiceServer(server, h)
	h.logger.Debug("gRPC services registered")
}

func (h *Handler) Shutdown() {
	h.cancel()
}

func (h *Handler) NotifyShutdown() {
	h.logger.Info("Notifying clients about server shutdown")
	h.cancel()
}

func (h *Handler) StreamMetrics(stream pb.MetricService_StreamMetricsServer) error {
	h.logger.Info("New metrics stream started")

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-h.ctx.Done():
			response := &pb.MetricsResponse{
				Status:  "shutdown",
				Message: "Server is shutting down",
			}
			if err := stream.Send(response); err != nil {
				h.logger.Error("Failed to send shutdown notification", zap.Error(err))
			}
			cancel()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stream terminated", zap.Error(ctx.Err()))
			return status.Error(codes.Canceled, "stream terminated")
		default:
			metrics, err := stream.Recv()
			if err != nil {
				h.logger.Error("Error receiving metrics", zap.Error(err))
				return status.Error(codes.Internal, "failed to receive metrics")
			}

			h.logger.Debug("Received metrics",
				zap.String("hostname", metrics.Host.Hostname),
				zap.Time("timestamp", metrics.Timestamp.AsTime()),
				zap.Int("cpu_count", len(metrics.Cpu)),
				zap.Int("disk_count", len(metrics.Disk)),
				zap.Int("network_count", len(metrics.Network)),
				zap.Int("docker_count", len(metrics.Docker)))

			// Store metrics in VictoriaMetrics
			if err := h.store.StoreAllMetrics(metrics); err != nil {
				h.logger.Error("Failed to store metrics", zap.Error(err))
				return status.Error(codes.Internal, "failed to store metrics")
			}

			if err := stream.Send(&pb.MetricsResponse{
				Status:  "ok",
				Message: "Metrics received and stored successfully",
			}); err != nil {
				h.logger.Error("Error sending response", zap.Error(err))
				return status.Error(codes.Internal, "failed to send response")
			}
		}
	}
}

func (h *Handler) GetMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsPayload, error) {
	h.logger.Info("GetMetrics called",
		zap.String("host_filter", req.HostFilter),
		zap.String("metric_type", req.MetricType))

	return nil, status.Error(codes.Unimplemented, "method GetMetrics not implemented")
}

func (h *Handler) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *Handler) Watch(req *health.HealthCheckRequest, stream health.HealthService_WatchServer) error {
	h.logger.Info("New health watch stream started")

	ctx := stream.Context()

	clientID := uuid.New().String()
	p, _ := peer.FromContext(ctx)

	// Register client
	clientsLock.Lock()
	clients[clientID] = ClientInfo{
		ID:          clientID,
		Hostname:    req.Hostname,
		IPAddress:   p.Addr.String(),
		ConnectedAt: time.Now(),
	}
	clientsLock.Unlock()
	h.logger.Debug("Client connected", zap.String("client_id", clientID), zap.String("hostname", req.Hostname), zap.String("ip_address", p.Addr.String()))

	// Clean
	defer func() {
		clientsLock.Lock()
		delete(clients, clientID)
		clientsLock.Unlock()
		h.logger.Debug("Client disconnected", zap.String("client_id", clientID))
		h.logger.Info("Health watch stream terminated", zap.String("client_id", clientID))
	}()

	// Send initial health status
	if err := stream.Send(&health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}); err != nil {
		h.logger.Error("Error sending initial health status", zap.Error(err))
		return status.Error(codes.Internal, "failed to send health status")
	}

	// Wait for context cancellation or server shutdown
	select {
	case <-ctx.Done():
		return status.Error(codes.Canceled, "stream context canceled")
	case <-h.ctx.Done():
		// Server is shutting down, notify client
		h.logger.Info("Server is shutting down, notifying client")

		_ = stream.Send(&health.HealthCheckResponse{
			Status: health.HealthCheckResponse_NOT_SERVING,
		})
		return nil
	}
}
