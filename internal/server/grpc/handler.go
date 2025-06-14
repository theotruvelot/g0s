package grpc

import (
	"context"

	"github.com/theotruvelot/g0s/internal/server/service"
	"github.com/theotruvelot/g0s/internal/server/storage/metrics"
	"github.com/theotruvelot/g0s/pkg/logger"
	"google.golang.org/grpc"
)

// Handler orchestrates all gRPC handlers
type Handler struct {
	authHandler        *AuthHandler
	metricsHandler     *MetricsHandler
	healthCheckHandler *HealthCheckHandler
	ctx                context.Context
	cancel             context.CancelFunc
}

// New creates a new handler orchestrator
func New(store *metrics.Manager, authService *service.AuthService, healthCheckService *service.HealthCheckService) *Handler {
	ctx, cancel := context.WithCancel(context.Background())

	metricService := service.NewMetricService(store)

	return &Handler{
		authHandler:        NewAuthHandler(authService),
		metricsHandler:     NewMetricsHandler(metricService),
		healthCheckHandler: NewHealthCheckHandler(healthCheckService),
		ctx:                ctx,
		cancel:             cancel,
	}
}

// RegisterServices registers all gRPC services
func (h *Handler) RegisterServices(server *grpc.Server) {
	h.authHandler.RegisterServices(server)
	h.metricsHandler.RegisterServices(server)
	h.healthCheckHandler.RegisterServices(server)
	logger.Debug("All gRPC services registered")
}

// Shutdown gracefully shuts down all handlers
func (h *Handler) Shutdown() {
	logger.Info("Shutting down all gRPC handlers")
	h.authHandler.Shutdown()
	h.metricsHandler.Shutdown()
	h.healthCheckHandler.Shutdown()
	h.cancel()
}

// NotifyShutdown notifies all handlers about server shutdown
func (h *Handler) NotifyShutdown() {
	logger.Info("Notifying all handlers about server shutdown")
	h.authHandler.NotifyShutdown()
	h.metricsHandler.NotifyShutdown()
	h.healthCheckHandler.NotifyShutdown()
	h.cancel()
}
