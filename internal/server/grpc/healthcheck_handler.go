package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/theotruvelot/g0s/internal/server/service"
	"github.com/theotruvelot/g0s/pkg/logger"
	"github.com/theotruvelot/g0s/pkg/proto/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type HealthCheckHandler struct {
	health.UnimplementedHealthServiceServer
	service *service.HealthCheckService
}

func NewHealthCheckHandler(svc *service.HealthCheckService) *HealthCheckHandler {
	return &HealthCheckHandler{
		service: svc,
	}
}

func (h *HealthCheckHandler) RegisterServices(server *grpc.Server) {
	health.RegisterHealthServiceServer(server, h)
	logger.Debug("Health check gRPC service registered")
}

func (h *HealthCheckHandler) Shutdown() {
	h.service.Shutdown()
}

func (h *HealthCheckHandler) NotifyShutdown() {
	h.service.NotifyShutdown()
}

func (h *HealthCheckHandler) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return h.service.Check(ctx, req)
}

func (h *HealthCheckHandler) Watch(req *health.HealthCheckRequest, stream health.HealthService_WatchServer) error {
	logger.Info("New health watch stream started")
	ctx := stream.Context()
	clientID := uuid.New().String()
	p, _ := peer.FromContext(ctx)
	ip := ""
	if p != nil {
		ip = p.Addr.String()
	}
	return h.service.Watch(ctx, clientID, req.Hostname, ip, func(status health.HealthCheckResponse_ServingStatus) error {
		return stream.Send(&health.HealthCheckResponse{Status: status})
	})
}
