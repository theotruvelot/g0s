package grpc

import (
	"context"

	"github.com/theotruvelot/g0s/internal/server/service"
	"github.com/theotruvelot/g0s/pkg/logger"
	pb "github.com/theotruvelot/g0s/pkg/proto/metric"
	"google.golang.org/grpc"
)

type MetricsHandler struct {
	pb.UnimplementedMetricServiceServer
	service *service.MetricService
}

func NewMetricsHandler(store *service.MetricService) *MetricsHandler {
	return &MetricsHandler{
		service: store,
	}
}

func (h *MetricsHandler) RegisterServices(server *grpc.Server) {
	pb.RegisterMetricServiceServer(server, h)
	logger.Debug("Metrics gRPC service registered")
}

func (h *MetricsHandler) Shutdown() {
	h.service.Shutdown()
}

func (h *MetricsHandler) NotifyShutdown() {
	h.service.NotifyShutdown()
}

func (h *MetricsHandler) SendStreamMetrics(stream pb.MetricService_StreamMetricsServer) error {
	return h.service.SendStreamMetrics(stream)
}

func (h *MetricsHandler) GetMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsPayload, error) {
	return h.service.GetMetrics(ctx, req)
}
