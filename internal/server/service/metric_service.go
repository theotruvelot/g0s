package service

import (
	"context"

	"github.com/theotruvelot/g0s/internal/server/storage/metrics"
	"github.com/theotruvelot/g0s/pkg/logger"
	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricService struct {
	store  *metrics.Manager
	ctx    context.Context
	cancel context.CancelFunc
}

func NewMetricService(store *metrics.Manager) *MetricService {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricService{
		store:  store,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *MetricService) Shutdown() {
	s.cancel()
}

func (s *MetricService) NotifyShutdown() {
	logger.Info("Notifying metrics clients about server shutdown")
	s.cancel()
}

func (s *MetricService) SendStreamMetrics(stream pb.MetricService_StreamMetricsServer) error {
	logger.Info("New metrics stream started")

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-s.ctx.Done():
			response := &pb.MetricsResponse{
				Status:  "shutdown",
				Message: "Server is shutting down",
			}
			if err := stream.Send(response); err != nil {
				logger.Error("Failed to send shutdown notification", zap.Error(err))
			}
			cancel()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stream terminated", zap.Error(ctx.Err()))
			return status.Error(codes.Canceled, "stream terminated")
		default:
			metrics, err := stream.Recv()
			if err != nil {
				logger.Error("Error receiving metrics", zap.Error(err))
				return status.Error(codes.Internal, "failed to receive metrics")
			}

			logger.Debug("Received metrics",
				zap.String("hostname", metrics.Host.Hostname),
				zap.Time("timestamp", metrics.Timestamp.AsTime()),
				zap.Int("cpu_count", len(metrics.Cpu)),
				zap.Int("disk_count", len(metrics.Disk)),
				zap.Int("network_count", len(metrics.Network)),
				zap.Int("docker_count", len(metrics.Docker)))

			// Store metrics in VictoriaMetrics
			if err := s.store.StoreAllMetrics(metrics); err != nil {
				logger.Error("Failed to store metrics", zap.Error(err))
				return status.Error(codes.Internal, "failed to store metrics")
			}

			if err := stream.Send(&pb.MetricsResponse{
				Status:  "ok",
				Message: "Metrics received and stored successfully",
			}); err != nil {
				logger.Error("Error sending response", zap.Error(err))
				return status.Error(codes.Internal, "failed to send response")
			}
		}
	}
}

func (s *MetricService) GetMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsPayload, error) {
	logger.Info("GetMetrics called",
		zap.String("host_filter", req.HostFilter),
		zap.String("metric_type", req.MetricType))

	return nil, status.Error(codes.Unimplemented, "method GetMetrics not implemented")
}
